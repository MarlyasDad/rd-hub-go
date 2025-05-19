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
		subscriptions: make(map[GUID]SubscriptionState),
		queue:         NewChainQueue(10000),
	}
}

type (
	GUID         string
	SubscriberID uuid.UUID
)

type SubscriptionState struct {
	Active bool
	Items  map[SubscriberID]*Subscriber
}

type Websocket struct {
	url           string
	conn          *websocket.Conn
	queue         *ChainQueue
	done          chan interface{}
	subscriptions map[GUID]SubscriptionState
	metrics       map[string]interface{}
	mu            sync.Mutex
}

func (ws *Websocket) Connect(ctx context.Context) error {
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, ws.url, nil)
	if err != nil {
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
	go ws.runWebsocketHealthLoop(ctx)
	go ws.runQueueLoop(ctx)

	return nil
}

func (ws *Websocket) Disconnect() error {
	if ws.IsConnected() {
		err := ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Println("Error during closing websocket:", err)
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
	RequestGuid GUID            `json:"requestGuid"`
	Guid        GUID            `json:"guid"`
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

func (ws *Websocket) AddSubscription(subscriber *Subscriber, guid string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	_, ok := ws.subscriptions[GUID(guid)]
	if !ok {
		ws.subscriptions[GUID(guid)] = SubscriptionState{
			Active: false,
			Items:  make(map[SubscriberID]*Subscriber),
		}
	}

	ws.subscriptions[GUID(guid)].Items[SubscriberID(subscriber.ID)] = subscriber
}

func (ws *Websocket) RemoveSubscription(subscriberID uuid.UUID, guid string) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Получаем саму подписку
	_, ok := ws.subscriptions[GUID(guid)]
	if !ok {
		return errors.New("the subscription is not exists")
	}

	// Удаляем подписчика из подписки
	delete(ws.subscriptions[GUID(guid)].Items, SubscriberID(subscriberID))

	// Удаляем подписку если она пустая
	if len(ws.subscriptions[GUID(guid)].Items) == 0 {
		delete(ws.subscriptions, GUID(guid))
	}

	return nil
}

func (ws *Websocket) runWebsocketLoop(ctx context.Context) {
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
	defer log.Println("queue loop closed")
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

func (ws *Websocket) runWebsocketHealthLoop(ctx context.Context) {
	defer log.Println("websocket health loop closed")
	for {
		select {
		case <-ctx.Done():
			return
		case <-ws.done:
			time.Sleep(time.Second * 5)
			log.Println("try to websocket reconnect")
			err := ws.Connect(ctx)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
