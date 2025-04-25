package alor

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"log"
	"net/http"
	"sync"
)

type Client struct {
	Config      Config
	Hosts       Hosts
	Token       Token
	Client      *http.Client
	Websocket   *Websocket
	Subscribers map[SubscriberID]*Subscriber
	// Subscriptions map[GUID]SubscriptionState
	mu sync.Mutex
}

func New(config Config) *Client {
	hosts := circuits.Development

	if !config.DevCircuit {
		hosts = circuits.Production
	}

	httpClient := &http.Client{}
	//subscribers := make(map[SubscriberID]*Subscriber)
	//subscriptions := make(map[GUID]SubscriptionState)

	return &Client{
		Config:      config,
		Hosts:       hosts,
		Token:       NewToken(config.RefreshToken, config.RefreshTokenExp),
		Client:      httpClient,
		Websocket:   NewWebsocket(hosts.Websocket),
		Subscribers: make(map[SubscriberID]*Subscriber),
		// Subscriptions: make(map[GUID]SubscriptionState),
	}
}

func (c *Client) Connect(ctx context.Context, websocket bool) error {
	err := c.RefreshToken()
	if err != nil {
		return err
	}

	if websocket {
		err = c.Websocket.Connect()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) Stop() {
	err := c.RemoveAllSubscribers()
	if err != nil {
		log.Println("ahalai mahalai")
	}

	err = c.Websocket.Disconnect()
	if err != nil {
		log.Println("abra kadabra")
	}
}

func (c *Client) AddSubscriber(subscriber *Subscriber) error {
	// TODO: Вынести подписчиков в клиента

	log.Println("subscriber ", subscriber.ID, "start subscriptions")
	for opcode, subscription := range subscriber.Subscriptions {
		switch opcode {
		case BarsOpcode:
			_ = c.BarsSubscribe(subscriber.ID, subscription)
		case AllTradesOpcode:
			_ = c.AllTradesSubscribe(subscriber.ID, subscription)
		case OrderBookOpcode:
			_ = c.OrderBooksSubscribe(subscriber.ID, subscription)
		default:
			continue
		}
	}

	log.Println("subscriber ", subscriber.ID, "init")
	err := subscriber.CustomHandler.Init()
	if err != nil {
		return err
	}

	c.Websocket.AddSubscriber(subscriber)

	//c.mu.Lock()
	//defer c.mu.Unlock()
	//
	//c.Subscribers[SubscriberID(subscriber.ID)] = subscriber

	return nil
}

func (c *Client) RemoveSubscriber(subscriberID uuid.UUID) error {
	subscriber, ok := c.Websocket.subscribers[SubscriberID(subscriberID)]
	if !ok {
		// TODO: Error
		return nil
	}

	// Больше не принимает события
	subscriber.SetDone()

	// Отписывается от всех подписок
	for _, subscription := range subscriber.Subscriptions {
		_ = c.Unsubscribe(subscriberID, subscription.Guid)
		// TODO: Error
	}

	// Завершает handler
	_ = subscriber.CustomHandler.DeInit()
	// TODO: Error

	c.mu.Lock()
	defer c.mu.Unlock()
	// Удаляется из списка подписчиков
	c.Websocket.RemoveSubscriber(subscriberID)

	return nil
}

func (c *Client) RemoveAllSubscribers() error {
	for _, subscriber := range c.Websocket.subscribers {
		_ = c.RemoveSubscriber(subscriber.ID)
	}
	return nil
}

func (c *Client) GetSubscribers() []*Subscriber {
	subscribers := make([]*Subscriber, 0)

	for _, subscriber := range c.Websocket.subscribers {
		subscribers = append(subscribers, subscriber)
	}

	return subscribers
}

func (c *Client) GetAllSubscriberBars(subscriberID uuid.UUID) ([]*Bar, error) {
	subscriber, ok := c.Websocket.subscribers[SubscriberID(subscriberID)]
	if !ok {
		return nil, errors.New("subscriber does not exist")
	}

	return subscriber.BarsProcessor.bars.GetAllBars(), nil
}
