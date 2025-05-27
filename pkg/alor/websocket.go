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
	"time"
)

func NewWebsocket(url string) *Websocket {
	return &Websocket{
		url:           url,
		subscribers:   make(map[SubscriberID]*Subscriber),
		subscriptions: make(map[string]SubscriptionState),
		queue:         NewChainQueue(10000),
	}
}

type (
	GUID         string
	SubscriberID uuid.UUID
)

type SubscriptionState struct {
	Subscription *Subscription
	Active       bool
	Items        map[SubscriberID]*Subscriber
}

type Websocket struct {
	url           string
	conn          *websocket.Conn
	queue         *ChainQueue
	done          chan interface{}
	subscribers   map[SubscriberID]*Subscriber
	subscriptions map[string]SubscriptionState
	metrics       map[string]interface{}
	mu            sync.Mutex
}

func (ws *Websocket) Connect(ctx context.Context, token string) error {
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, ws.url, nil)
	if err != nil {
		close(ws.done)
		return fmt.Errorf("error connecting to Websocket Server: %w", err)
	}

	conn.EnableWriteCompression(true)
	ws.conn = conn
	//ws.conn.SetCloseHandler(func(code int, text string) error {
	//	fmt.Println("Websocket disconnected")
	//	return nil
	//})

	ws.done = make(chan interface{}) // Channel to indicate that the receiverHandler is done
	go ws.runWebsocketLoop(ctx)
	go ws.runQueueLoop(ctx)

	if err := ws.restoreSubscriptions(token); err != nil {
		return fmt.Errorf("error restoring subscriptions: %w", err)
	}

	return nil
}

func (ws *Websocket) Reconnect(ctx context.Context, token string) error {
	if err := ws.Disconnect(); err != nil {
		return err
	}

	if err := ws.Connect(ctx, token); err != nil {
		return err
	}

	return nil
}

func (ws *Websocket) Disconnect() error {
	if ws.IsConnected() {
		close(ws.done)

		err := ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Println("Error during sending websocket close message:", err)
			return err
		}
	}

	if ws.conn != nil {
		err := ws.conn.Close()
		if err != nil {
			log.Println("Error during closing websocket:", err)
			return err
		}
	}

	return nil
}

func (ws *Websocket) IsConnected() bool {
	if ws.done == nil {
		return false
	}

	select {
	case <-ws.done:
		return false
	default:
		return true
	}
}

func (ws *Websocket) SendMessage(msg []byte) error {
	return ws.conn.WriteMessage(websocket.TextMessage, msg)
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

func (ws *Websocket) HandleResponse(msg []byte) error {
	// log.Printf("Received: %s\n", msg)

	var response WsResponse
	err := json.Unmarshal(msg, &response)
	if err != nil {
		return err
	}

	// Проверяем request на ошибку
	// Проверяем, что это, дата или отбивка
	// Если отбивка, запускаем метод обработки отбивки
	// if event == подтверждение подписки
	// То ...
	if response.RequestGuid != "" {
		// обрабатываем статус подписки
		guidParts := strings.Split(string(response.RequestGuid), "-")
		_ = Opcode(guidParts[2]) // from guid
	}

	// Если guid пустой, печатаем warning
	if response.Guid == "" {
		return nil
	}

	guidParts := strings.Split(string(response.Guid), "-")
	opcode := Opcode(guidParts[2]) // from guid

	event := &ChainEvent{
		Opcode: opcode,
		Guid:   response.Guid,
		Data:   response.Data,
	}

	err = ws.queue.Enqueue(event)
	if err != nil {
		return err
	}

	// log.Println("length enqueue", ws.queue.GetLength())

	return nil
}

func (ws *Websocket) AddSubscription(subscription *Subscription, subscriber *Subscriber) {
	//ws.mu.Lock()
	//defer ws.mu.Unlock()

	_, ok := ws.subscriptions[subscription.Guid]
	if !ok {
		ws.subscriptions[subscription.Guid] = SubscriptionState{
			Subscription: subscription,
			Active:       false,
			Items:        make(map[SubscriberID]*Subscriber),
		}
	}

	ws.subscriptions[subscription.Guid].Items[SubscriberID(subscriber.ID)] = subscriber
}

func (ws *Websocket) RemoveSubscription(subscriberID uuid.UUID, guid string) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Получаем саму подписку
	_, ok := ws.subscriptions[guid]
	if !ok {
		return errors.New("the subscription is not exists")
	}

	// Удаляем подписчика из подписки
	delete(ws.subscriptions[guid].Items, SubscriberID(subscriberID))

	// Удаляем подписку если она пустая
	if len(ws.subscriptions[guid].Items) == 0 {
		delete(ws.subscriptions, guid)
	}

	return nil
}

func (ws *Websocket) runWebsocketLoop(ctx context.Context) {
	log.Println("websocket loop is running")
	defer close(ws.done)
	defer log.Println("websocket loop closed")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, msg, err := ws.conn.ReadMessage()
			if err != nil {
				log.Println("Error in receive:", err)
				return
			}

			// Обрабатываем входящее сообщение
			err = ws.HandleResponse(msg)
			if err != nil {
				log.Println("Error in handle:", err)
				return
			}
		}
	}
}

func (ws *Websocket) runQueueLoop(ctx context.Context) {
	log.Println("websocket queue loop is running")
	defer log.Println("websocket queue loop closed")
	for {
		select {
		case <-ctx.Done():
			return
		case <-ws.done:
			return
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

			log.Println("length dequeue", ws.queue.GetLength())

			// log.Println(event)

			// Устанавливаем подписку как активную если по ней пришло событие
			if !ws.subscriptions[event.Guid].Active {
				item := ws.subscriptions[event.Guid]
				item.Active = true
				ws.subscriptions[event.Guid] = item
			}

			// Последовательное выполнение может занимать много времени - тогда заменить на асинхронные обработчики
			for _, subscriber := range ws.subscriptions[event.Guid].Items {
				// Блокируем добавление/удаление любых подписчиков пока не пройдёт handle
				// Никто не может поменять subscriptions во время вычисления
				// Добавление или удаление из-за этого может занять продолжительное время
				ws.mu.Lock()

				if subscriber == nil || subscriber.Done {
					// Разблокируем удаление подписчиков
					ws.mu.Unlock()
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

				if err := subscriber.HandleEvent(event); err != nil {
					subscriber.SetDone()
					log.Println(subscriber.ID, "Error in handle:", err)
				}

				// Разблокируем удаление подписчиков
				ws.mu.Unlock()
			}
		}
	}
}

func (ws *Websocket) AddSubscriber(token string, subscriber *Subscriber) error {
	log.Println("subscriber ", subscriber.ID, "start subscriptions", subscriber)

	ws.mu.Lock()
	defer ws.mu.Unlock()

	for _, subscription := range subscriber.Subscriptions {
		if err := ws.Subscribe(token, subscriber, subscription); err != nil {
			return err
		}

		ws.subscriptions[subscription.Guid].Items[SubscriberID(subscriber.ID)] = subscriber
	}

	log.Println("subscriber ", subscriber.ID, "init")
	if subscriber.CustomHandler != nil {
		if err := subscriber.CustomHandler.Init(); err != nil {
			return err
		}
	}

	ws.subscribers[SubscriberID(subscriber.ID)] = subscriber

	return nil
}

func (ws *Websocket) RemoveSubscriber(token string, subscriberID uuid.UUID) error {
	subscriber, ok := ws.subscribers[SubscriberID(subscriberID)]
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

	ws.mu.Lock()
	defer ws.mu.Unlock()

	delete(ws.subscribers, SubscriberID(subscriberID))

	return nil
}

func (ws *Websocket) Subscribe(token string, subscriber *Subscriber, subscription *Subscription) error {
	requestBytes, err := ws.prepareRequest(token, subscription)
	if err != nil {
		return err
	}

	if err := ws.SendMessage(requestBytes); err != nil {
		return err
	}

	ws.AddSubscription(subscription, subscriber)

	return nil
}

func (ws *Websocket) prepareRequest(token string, subscription *Subscription) ([]byte, error) {
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

func (ws *Websocket) Unsubscribe(token string, subscriberID uuid.UUID, guid string) error {
	if err := ws.RemoveSubscription(subscriberID, guid); err != nil {
		return err
	}

	// Выйти если ещё остались подписчики
	_, ok := ws.subscriptions[guid]
	if !ok {
		return errors.New("the subscription is not exists")
	}

	if len(ws.subscriptions[guid].Items) > 0 {
		return nil
	}

	request := UnsubscribeRequest{
		Opcode: UnsubscribeOpcode,
		Token:  token,
		GUID:   guid,
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	return ws.SendMessage(requestBytes)
}

func (ws *Websocket) GetSubscriber(subscriberID uuid.UUID) (*Subscriber, error) {
	subscriber, ok := ws.subscribers[SubscriberID(subscriberID)]
	if !ok {
		return nil, errors.New("subscriber not found")
	}

	return subscriber, nil
}

func (ws *Websocket) GetSubscribers() []*Subscriber {
	subscribers := make([]*Subscriber, 0)

	for _, subscriber := range ws.subscribers {
		subscribers = append(subscribers, subscriber)
	}

	return subscribers
}

func (ws *Websocket) RemoveAllSubscribers(token string) error {
	for _, subscriber := range ws.subscribers {
		_ = ws.RemoveSubscriber(token, subscriber.ID)
	}

	return nil
}

func (ws *Websocket) GetAllSubscriberBars(subscriberID uuid.UUID) ([]*Bar, error) {
	subscriber, ok := ws.subscribers[SubscriberID(subscriberID)]
	if !ok {
		return nil, errors.New("subscriber does not exist")
	}

	return subscriber.BarsProcessor.bars.GetAllBars(), nil
}

func (ws *Websocket) restoreSubscriptions(token string) error {
	for _, subscriptionState := range ws.subscriptions {
		requestBytes, err := ws.prepareRequest(token, subscriptionState.Subscription)
		if err != nil {
			return err
		}

		if err := ws.SendMessage(requestBytes); err != nil {
			return err
		}
	}

	return nil
}
