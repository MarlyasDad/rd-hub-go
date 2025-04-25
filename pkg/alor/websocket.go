package alor

import (
	"encoding/json"
	"errors"
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
	Items  map[SubscriberID]bool
}

type Websocket struct {
	url           string
	conn          *websocket.Conn
	queue         *ChainQueue
	done          chan interface{}
	subscribers   map[SubscriberID]*Subscriber
	subscriptions map[GUID]SubscriptionState
	mu            sync.Mutex
}

func (ws *Websocket) runWebsocketLoop(connection *websocket.Conn) {
	defer close(ws.done)
	defer log.Println("websocket loop closed")
	for {
		_, msg, err := connection.ReadMessage()
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

func (ws *Websocket) Connect() error {
	ws.done = make(chan interface{}) // Channel to indicate that the receiverHandler is done

	// socketUrl := "ws://localhost:8080" + "/socket"
	conn, _, err := websocket.DefaultDialer.Dial(ws.url, nil)
	if err != nil {
		log.Println("Error connecting to Websocket Server:", err)
		return err
	}

	conn.EnableWriteCompression(true)

	ws.conn = conn

	go ws.runWebsocketLoop(ws.conn)
	go ws.runQueueLoop()

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

func (ws *Websocket) AddSubscription(subscriberID uuid.UUID, guid string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	_, ok := ws.subscriptions[GUID(guid)]
	if !ok {
		ws.subscriptions[GUID(guid)] = SubscriptionState{
			Active: false,
			Items:  make(map[SubscriberID]bool),
		}
	}

	ws.subscriptions[GUID(guid)].Items[SubscriberID(subscriberID)] = true
}

func (ws *Websocket) RemoveSubscription(subscriberID uuid.UUID, guid string) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	_, ok := ws.subscriptions[GUID(guid)]
	if !ok {
		return errors.New("the subscription is not exists")
	}

	delete(ws.subscriptions[GUID(guid)].Items, SubscriberID(subscriberID))

	//for i, v := range ws.subscriptions[GUID(guid)] {
	//	if v == subscriberID {
	//		ws.subscriptions[guid] = append(ws.subscriptions[guid][:i], ws.subscriptions[guid][i+1:]...)
	//	}
	//}

	return nil
}

//func (ws *Websocket) RemoveAllSubscriberSubscriptions(subscriberID uuid.UUID) error {
//	ws.mu.Lock()
//	defer ws.mu.Unlock()
//
//	for guid, _ := range ws.subscriptions {
//		delete(ws.subscriptions[guid].Items, SubscriberID(subscriberID))
//
//		//for i, v := range ws.subscriptions[key] {
//		//	if v == subscriberID {
//		//		ws.subscriptions[key] = append(ws.subscriptions[key][:i], ws.subscriptions[key][i+1:]...)
//		//	}
//		//}
//	}
//
//	return nil
//}

func (ws *Websocket) AddSubscriber(subscriber *Subscriber) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.subscribers[SubscriberID(subscriber.ID)] = subscriber
}

func (ws *Websocket) RemoveSubscriber(subscriberID uuid.UUID) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	delete(ws.subscribers, SubscriberID(subscriberID))
}

func (ws *Websocket) runQueueLoop() {
	defer log.Println("queue loop closed")
	for {
		select {
		case <-ws.done:
			break
		default:
			event, err := ws.queue.Dequeue()
			if err != nil {
				if errors.Is(err, ErrQueueUnderFlow) {
					time.Sleep(time.Millisecond * 100)
					continue
				}

				log.Println("Error in receive:", err)
				break
			}

			log.Println("length dequeue", ws.queue.GetLength())

			// log.Println(event)

			// Проходим по всем подписчикам
			for subscriber, _ := range ws.subscriptions[event.Guid].Items {
				if !ws.IsConnected() {
					return
				}
				// Handle event
				_ = ws.subscribers[subscriber].HandleEvent(event)
			}
		}
	}
}
