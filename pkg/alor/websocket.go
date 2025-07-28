package alor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func NewWebsocket(url string) *Websocket {
	return &Websocket{
		url:           url,
		subscribers:   NewSubscribers(),
		subscriptions: NewSubscriptions(),
		queue:         NewChainQueue(10000),
		done:          make(chan struct{}),
		reconnect:     make(chan struct{}, 1),
		ready:         true,
	}
}

type (
	Websocket struct {
		url           string
		connection    *websocket.Conn
		queue         *ChainQueue
		subscribers   Subscribers
		subscriptions Subscriptions
		done          chan struct{}  // Основной канал для остановки всех горутин
		reconnect     chan struct{}  // Канал для инициации переподключения
		isConnecting  atomic.Bool    // Флаг процесса подключения
		isClosing     atomic.Bool    // Флаг процесса отключения
		mu            sync.Mutex     // Защищает connection
		wg            sync.WaitGroup // Для ожидания завершения всех горутин
		ready         bool
	}
)

// Connect безопасно устанавливает соединение с защитой от повторных вызовов
func (ws *Websocket) Connect() error {
	// Проверяем и устанавливаем флаг отключения
	// Если он true, то в данный момент есть активный процесс подключения
	if ws.isClosing.Load() {
		return errors.New("websocket отключается")
	}

	// Проверяем и устанавливаем флаг подключения
	// Если он true, то в данный момент есть активный процесс подключения
	if !ws.isConnecting.CompareAndSwap(false, true) {
		return nil // или можно вернуть ошибку "уже подключается"
	}
	defer ws.isConnecting.Store(false)

	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Если уже подключены - ничего не делаем
	if ws.connection != nil {
		return nil
	}

	log.Printf("Подключение к %s", ws.url)

	conn, _, err := websocket.DefaultDialer.Dial(ws.url, nil)
	if err != nil {
		return fmt.Errorf("error connecting to Websocket Server: %w", err)
	}
	conn.EnableWriteCompression(true)

	ws.connection = conn
	return nil
}

// Close безопасно закрывает соединение с защитой от повторных вызовов
func (ws *Websocket) Close() error {
	// Проверяем и устанавливаем флаг отключения
	// Если он true, то в данный момент есть активный процесс отключения
	if !ws.isClosing.CompareAndSwap(false, true) {
		return nil // или можно вернуть ошибку "уже отключается"
	}
	defer ws.isClosing.Store(false)

	// Блокируем сокет и не даём одновременно
	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Всегда возвращаем канал в состояние готовности после отключения
	defer func() { ws.done = make(chan struct{}) }()

	// Сигнализируем о закрытии
	select {
	case <-ws.done:
		return nil // Уже закрыт
	default:
		close(ws.done)
	}

	// Если уже отключены - ничего не делаем
	if ws.connection == nil {
		log.Println("already closed")
		return nil
	}

	// Сбрасываем подключение при выходе
	defer func() { ws.connection = nil }()

	log.Println("sending close message")
	// Отправляем сообщение о закрытии
	err := ws.connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("Ошибка при отправке CloseMessage:", err)
		return err
	}

	// Закрываем соединение
	err = ws.connection.Close()
	if err != nil {
		log.Println("Ошибка при закрытии соединения:", err)
		return err
	}

	log.Println("waiting all websocket goroutines")
	// Ждем завершения всех горутин
	ws.wg.Wait()
	return nil
}

func (ws *Websocket) Subscribe(token Token, subscriberID SubscriberID, subscription *Subscription) error {
	// Подготавливаем запрос
	requestBytes, err := ws.prepareRequest(token, subscription)
	if err != nil {
		return err
	}

	// отправляем в сокет
	if err := ws.Send(requestBytes); err != nil {
		return err
	}

	// Добавляем подписку в пул
	ws.subscriptions.Add(subscriberID, subscription)

	return nil
}

func (ws *Websocket) Unsubscribe(token string, subscriberID SubscriberID, guid GUID) error {
	request := UnsubscribeRequest{
		Opcode: UnsubscribeOpcode,
		Token:  token,
		GUID:   guid,
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	if err := ws.subscriptions.Delete(subscriberID, guid); err != nil {
		return err
	}

	return ws.Send(requestBytes)
}

//func (ws *Websocket) GetSubscribers() []*Subscriber {
//	subscribers := make([]*Subscriber, 0)
//
//	for _, subscriber := range ws.subscribers.All() {
//		subscribers = append(subscribers, subscriber)
//	}
//
//	return subscribers
//}
//
//func (ws *Websocket) GetSubscriptions() []*Subscriber {
//	subscribers := make([]*Subscriber, 0)
//	return subscribers
//}

func (ws *Websocket) Send(message []byte) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	select {
	case <-ws.done:
		return fmt.Errorf("connection is closing")
	default:
	}

	if ws.connection == nil {
		return fmt.Errorf("connection is not established")
	}

	return ws.connection.WriteMessage(websocket.TextMessage, message)
}

// {
// "message": "Handled successfully",
// "httpCode": 200,
// "requestGuid": "c328fcf1-e495-408a-a0ed-e20f95d6b813"
// }
// {
// "requestGuid": "c328fcf1-e495-408a-a0ed-e20f95d6b813",
// "httpCode": 401,
// "message": "Invalid JWT token!"
// }

type WsResponse struct {
	Message     string          `json:"message"`
	Data        json.RawMessage `json:"data"`
	HttpCode    int             `json:"httpCode"`
	RequestGuid string          `json:"requestGuid"`
	Guid        string          `json:"guid"`
}

func (ws *Websocket) Listen() {
	log.Println("running websocket listen loop")
	defer log.Println("websocket listen loop closed")

	ws.wg.Add(1)
	defer ws.wg.Done()

	for {
		select {
		case <-ws.done:
			return
		default:
			// получаем сообщение
			_, message, err := ws.connection.ReadMessage()
			if err != nil {
				log.Println("Ошибка чтения:", err)
				select {
				case ws.reconnect <- struct{}{}:
				case <-ws.done:
				}
				return
			}
			// log.Printf("Получено: %s", message)

			// обрабатываем сообщение
			var response WsResponse
			err = json.Unmarshal(message, &response)
			if err != nil {
				log.Println("unmarshall failed", err)
				return
			}

			err = ws.HandleResponse(response)
			if err != nil {
				log.Println("error in handle:", err)
				return
			}
		}

	}
}

func (ws *Websocket) HandleResponse(message WsResponse) error {
	if message.RequestGuid == "" && message.Guid == "" {
		return errors.New("guid is empty")
	}

	event := &ChainEvent{
		Data: message.Data,
	}

	// отделяем системные сообщения от данных
	if message.RequestGuid != "" {
		event.Type = SystemType
		event.Guid = message.RequestGuid
	}

	if message.Guid != "" {
		event.Type = DataType
		event.Guid = message.Guid
	}

	guidParts := strings.Split(event.Guid, "-")
	event.Opcode = Opcode(guidParts[3])

	err := ws.queue.Enqueue(event)
	if err != nil {
		if errors.Is(err, ErrQueueOverFlow) {
			log.Println("queue too big", ws.queue.GetLength())
			time.Sleep(time.Second * 1)
			ws.ready = false
		}
		return err
	}

	// log.Println("length enqueue", ws.queue.GetLength())

	return nil
}

func (ws *Websocket) ReconnectHandler(ctx context.Context, token Token) {
	log.Println("running websocket reconnect loop")
	defer log.Println("websocket reconnect loop closed")

	ws.wg.Add(1)
	defer ws.wg.Done()

	for {
		select {
		case <-ws.reconnect:
			if ws.isClosing.Load() {
				continue
			}

			log.Println("Попытка переподключения...")

			// Закрываем предыдущее соединение, если оно есть
			// надо ли?
			//if ws.connection != nil {
			//	ws.connection.Close()
			//}

			// Пытаемся подключиться с экспоненциальной задержкой
			retryDelay := time.Second
			maxRetryDelay := 30 * time.Second

			for {
				if ws.isClosing.Load() {
					break
				}

				// TODO: надо ли проверять?
				//select {
				//case <-ws.done:
				//	return
				//default:
				//}

				err := ws.Connect()
				if err == nil {
					log.Println("Переподключение успешно")
					// Запускаем разбор очереди
					go ws.SortQueue(ctx, token)
					// Запускаем слушатель сообщений
					go ws.Listen()

					if err := ws.restoreSubscriptions(token); err != nil {
						log.Println(err)
					}
					break
				}

				log.Printf("Ошибка переподключения: %v, следующая попытка через %v", err, retryDelay)

				select {
				case <-time.After(retryDelay):
				case <-ws.done:
					return
				}

				retryDelay = min(retryDelay*2, maxRetryDelay)
				// retryDelay = time.Duration(min(int64(retryDelay*2), int64(30*time.Second)))
			}
		case <-ws.done:
			return
		}
	}
}

func (ws *Websocket) restoreSubscriptions(token Token) error {
	for _, subscriptionState := range ws.subscriptions.All() {
		requestBytes, err := ws.prepareRequest(token, subscriptionState.Subscription)
		if err != nil {
			return err
		}

		if err := ws.Send(requestBytes); err != nil {
			return err
		}
	}

	return nil
}

func (ws *Websocket) prepareRequest(token Token, subscription *Subscription) ([]byte, error) {
	switch subscription.Opcode {
	case BarsOpcode:
		return ws.BarsSubscribe(token, subscription)
	case AllTradesOpcode:
		return ws.AllTradesSubscribe(token, subscription)
	case OrderBookOpcode:
		return ws.OrderBooksSubscribe(token, subscription)
	}

	return nil, errors.New("invalid opcode")
}

func (ws *Websocket) SortQueue(ctx context.Context, token Token) {
	log.Println("running websocket queue loop")
	defer log.Println("websocket queue loop closed")

	ws.wg.Add(1)
	defer ws.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return // завершаем когда сервер остановлен
		case <-ws.done:
			return // завершаем когда нажали close
		default:
			event, err := ws.queue.Dequeue()
			if err != nil {
				if errors.Is(err, ErrQueueUnderFlow) {
					time.Sleep(time.Millisecond * 500)
					continue
				}

				log.Println("Error in receive:", err)
				return
			}

			// Удаляем и добавляем подписки между итерациями
			ws.subscriptions.Rebalancing()

			// Удаляем и добавляем подписчиков между итерациями
			ws.subscribers.Rebalancing()

			log.Println("length dequeue", ws.queue.GetLength())

			// Последовательное выполнение может занимать много времени - тогда заменить на асинхронные обработчики
			// for _, subscriber := range ws.subscriptions.GetSubscriptionsByGUID(event.Guid) {
			for _, subscriber := range ws.subscriptions[event.Guid].Items {
				// Блокируем добавление/удаление любых подписчиков пока не пройдёт handle
				// Никто не может поменять subscriptions во время вычисления
				// Добавление или удаление из-за этого может занять продолжительное время

				if subscriber == nil || subscriber.Done {
					// Разблокируем удаление подписчиков
					continue
				}

				// TODO: Выполнять все вместе параллельно или каждый с асинхронным обработчиком?
				// Синхронное выполнение с задержкой хотя-бы одного воркера может тормозить остальные
				// Отличная идея - выполнять сабскриберы в горутинах. Тогда отставать будет самый нагруженный
				// А легковесные будут пролетать со свистом

				// !Проблема синхронизации между воркерами - один имеет актуальное состояние, а другой нет
				// Как решить - непонятно. Если только сравнивать длину очередей. Если с маленькой погрешностью не отличаются
				// Идея! Если нужно сделать зависимые сабскриберы - делать для них асинхронную оболочку и внутри обрабатывать события синхронно
				// subscribersGroup - интерфейс как у сабскрибера

				// !Проблема отставания от текущей ситуации (решается распараллеливанием) - проверять очередь на количество необработанных элементов
				// Всё, что работает в реалтайме не должно превышать определённый порог загруженности очереди
				// Причём, как общей очереди вебсокета, так и в частной очереди сабскрибера
				// Если очередь переполняется и не разгружается, то отключаем сабскрибера SetDone()
				// Так мы отсекаем самых медленных подписчиков
				// TODO: Нужно сделать проверку очереди вебсокета. При достижении 50тысяч, отключать вебсокет и алертить!
				// Или отключать всех сабскриберов
				// Если очередь больше дельты, не отправлять команды брокеру.
				// Работают только самые шустрые, медленные отключаются

				// В каждом сабскрибере делать свой контекст с отменой от родительского. Отменять горутину когда сабскрибер будет удаляться.

				// активируем подписку
				if item, ok := ws.subscriptions[event.Guid]; ok {
					if !item.Active {
						item.Active = true
						ws.subscriptions[event.Guid] = item
					}
				}

				if event.Type == SystemType {
					log.Printf("%+v\n", event)
				}

				if event.Type == DataType {
					if err := subscriber.HandleEvent(event); err != nil {
						subscriber.SetDone()
						log.Println(subscriber.ID, "Error in handle:", err)
					}
				}
			}
		}
	}
}

// IsConnected проверяет, активно ли соединение WebSocket
func (ws *Websocket) IsConnected() bool {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.connection == nil {
		return false
	}

	//  _, _, err := ws.connection.NextReader()
	//	if err != nil {
	//		return false
	//	}
	//	return true

	// Простая проверка через отправку ping сообщения
	err := ws.connection.WriteControl(
		websocket.PingMessage,
		[]byte{},
		time.Now().Add(1*time.Second),
	)

	return err == nil
}

func (ws *Websocket) AddSubscriber(token Token, subscriber *Subscriber) error {
	log.Println("subscriber subscribe", subscriber.ID, "start subscriptions", subscriber)

	ws.mu.Lock()
	defer ws.mu.Unlock()

	for _, subscription := range subscriber.Subscriptions {
		// подписываемся в вебсокете
		if err := ws.Subscribe(token, subscriber, subscription); err != nil {
			return err
		}

		// добавляем подписчика в подписку
		ws.subscriptions.Add(subscription)
		// ws.subscriptions[subscription.Guid].Items[SubscriberID(subscriber.ID)] = subscriber
	}

	log.Println("subscriber ", subscriber.ID, "init")
	if subscriber.CustomHandler != nil {
		if err := subscriber.CustomHandler.Init(); err != nil {
			return err
		}
	}

	// добавляем подписчика в список подписчиков
	ws.subscribers.Add(subscriber)
	// ws.subscribers[SubscriberID(subscriber.ID)] = subscriber

	return nil
}

func (ws *Websocket) RemoveSubscriber(token string, subscriberID uuid.UUID) error {
	log.Println("subscriber unsubscribe", subscriberID, "stop subscriptions")

	ws.mu.Lock()
	defer ws.mu.Unlock()

	subscriber, ok := ws.subscribers.Get(SubscriberID(subscriberID))
	if !ok {
		// TODO: Error
		return nil
	}

	// Больше не принимает события
	subscriber.SetDone()

	// Отписывается от всех подписок
	for _, subscription := range subscriber.Subscriptions {
		if err := ws.Unsubscribe(token, subscriberID, subscription.Guid); err != nil {
			return err
		}
	}

	if subscriber.CustomHandler != nil {
		if err := subscriber.CustomHandler.DeInit(); err != nil {
			return err
		}
	}

	delete(ws.subscribers, SubscriberID(subscriberID))

	return nil
}

func (ws *Websocket) GetSubscriber(subscriberID uuid.UUID) (*Subscriber, error) {
	subscriber, ok := ws.subscribers[SubscriberID(subscriberID)]
	if !ok {
		return nil, errors.New("subscriber not found")
	}

	return subscriber, nil
}

func (ws *Websocket) RemoveAllSubscribers(token string) error {
	for _, subscriber := range ws.subscribers {
		_ = ws.RemoveSubscriber(token, subscriber.ID)
	}

	return nil
}

func (ws *Websocket) GetAllStrategyBars(subscriberID uuid.UUID) ([]*Bar, error) {
	subscriber, ok := ws.subscribers[SubscriberID(subscriberID)]
	if !ok {
		return nil, errors.New("subscriber does not exist")
	}

	return subscriber.Strategy.DataProcessor.bars.GetAllBars(), nil
}
