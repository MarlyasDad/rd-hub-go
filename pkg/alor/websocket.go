package alor

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Subscriber interface {
}

func NewWebsocket(url string) *Websocket {
	return &Websocket{
		url: url,
	}
}

type Callback func(response WsResponse) error

type Websocket struct {
	url           string
	conn          *websocket.Conn
	queue         Queue
	done          chan interface{}
	subscribers   map[uuid.UUID]Subscriber
	subscriptions map[string][]uuid.UUID
	mu            sync.Mutex
	// counter       Counter
}

func (ws *Websocket) runWebsocketLoop(connection *websocket.Conn) {
	defer close(ws.done)
	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			log.Println("Error in receive:", err)
			return
		}

		// Обрабатываем входящее сообщение
		_ = ws.HandleResponse(msg)
	}
}

func (ws *Websocket) Connect() {
	ws.done = make(chan interface{}) // Channel to indicate that the receiverHandler is done

	// socketUrl := "ws://localhost:8080" + "/socket"
	conn, _, err := websocket.DefaultDialer.Dial(ws.url, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}

	// defer conn.Close()
	ws.conn = conn

	go ws.runWebsocketLoop(ws.conn)
}

func (ws *Websocket) Disconnect() {
	// Close our websocket connection
	err := ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("Error during closing websocket:", err)
		return
	}

	select {
	case <-ws.done:
		log.Println("Receiver Channel Closed! Exiting....")
	case <-time.After(time.Duration(1) * time.Second):
		log.Println("Timeout in closing receiving channel. Exiting....")
	}

	_ = ws.conn.Close()
}

func (ws *Websocket) SendMessage(msg []byte) error {
	return ws.conn.WriteMessage(websocket.TextMessage, msg)
}

// HandleResponse
// Ответы от websocket-a бывают двух видов:
// - подтверждение подписки
// - данные
func (ws *Websocket) HandleResponse(msg []byte) error {
	log.Printf("Received: %s\n", msg)

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

	opcode := BarsOpcode // from guid

	// QueueItem? QueueEvent?
	event := Event{
		Opcode: opcode,
		Guid:   response.Guid,
		Data:   response.Data,
	}

	err = ws.queue.Enqueue(event)
	if err != nil {
		return err
	}

	return nil
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

func (ws *Websocket) AddSubscription(subscriberID uuid.UUID, guid string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	_, ok := ws.subscriptions[guid]
	if !ok {
		ws.subscriptions[guid] = make([]uuid.UUID, 0)
	}

	ws.subscriptions[guid] = append(ws.subscriptions[guid], subscriberID)
}

func (ws *Websocket) RemoveSubscription(subscriberID uuid.UUID, guid string) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	_, ok := ws.subscriptions[guid]
	if !ok {
		return errors.New("the subscription is not exists")
	}

	for i, v := range ws.subscriptions[guid] {
		if v == subscriberID {
			ws.subscriptions[guid] = append(ws.subscriptions[guid][:i], ws.subscriptions[guid][i+1:]...)
		}
	}

	return nil
}

func (ws *Websocket) RemoveAllSubscriberSubscriptions(subscriberID uuid.UUID) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	for key, _ := range ws.subscriptions {
		_, ok := ws.subscriptions[key]
		if !ok {
			continue
		}

		for i, v := range ws.subscriptions[key] {
			if v == subscriberID {
				ws.subscriptions[key] = append(ws.subscriptions[key][:i], ws.subscriptions[key][i+1:]...)
			}
		}
	}

	return nil
}

func (ws *Websocket) AddSubscriber(subscriber Subscriber) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	// err := subscriber.Init()
}

func (ws *Websocket) RemoveSubscriber(subscriberID uuid.UUID) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	// err := subscriber.DeInit()
}
