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
	healthDone    chan interface{}
	subscribers   map[SubscriberID]*Subscriber
	subscriptions map[string]SubscriptionState
	metrics       map[string]interface{}
	mu            sync.Mutex
}

func (ws *Websocket) Connect(ctx context.Context, token string) error {
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

	ws.done = make(chan interface{})

	go ws.runQueueLoop(ctx)
	go ws.runWebsocketLoop(ctx)

	// Восстанавливаем подписки
	if err := ws.restoreSubscriptions(token); err != nil {
		return fmt.Errorf("error restoring subscriptions: %w", err)
	}

	return nil
}

//func (ws *Websocket) Reconnect(ctx context.Context, token string) error {
//	if err := ws.Disconnect(); err != nil {
//		return err
//	}
//
//	if err := ws.Connect(ctx, token); err != nil {
//		return err
//	}
//
//	return nil
//}

func (ws *Websocket) Disconnect() error {
	if ws.IsConnected() {
		// close(ws.done)

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
	// TODO: ТУТ ОТБИВКИ
	if response.RequestGuid != "" {
		log.Printf("Received: %s\n", msg)
		return nil
	}

	// Если guid пустой, печатаем warning
	if response.Guid == "" {
		return nil
	}

	guidParts := strings.Split(string(response.Guid), "-")
	opcode := Opcode(guidParts[3]) // from guid

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

	// Создать подписку если она не существует
	_, ok := ws.subscriptions[subscription.Guid]
	if !ok {
		ws.subscriptions[subscription.Guid] = SubscriptionState{
			Subscription: subscription,                       // Подписка
			Active:       false,                              // Пришло ли хоть одно сообщение по ней
			Items:        make(map[SubscriberID]*Subscriber), // Подписчики
		}
	}

	// Добавить подписчика в полдписку
	ws.subscriptions[subscription.Guid].Items[SubscriberID(subscriber.ID)] = subscriber

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
