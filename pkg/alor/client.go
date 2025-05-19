package alor

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"log"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	Config      Config
	Hosts       Hosts
	Token       Token
	Client      *http.Client
	Subscribers *map[SubscriberID]*Subscriber
	Websocket   *Websocket
	mu          sync.Mutex
}

func New(config Config) *Client {
	hosts := circuits.Development

	if !config.DevCircuit {
		hosts = circuits.Production
	}

	httpClient := &http.Client{Transport: http.DefaultTransport, Timeout: 30 * time.Second}
	//httpClient := &http.Client{Transport: &http.Transport{
	//	// Proxy: ProxyFromEnvironment,
	//	DialContext: (&net.Dialer{
	//		Timeout:   30 * time.Second,
	//		KeepAlive: 30 * time.Second,
	//	}).DialContext,
	//	ForceAttemptHTTP2:     true,
	//	MaxIdleConns:          100,
	//	IdleConnTimeout:       90 * time.Second,
	//	TLSHandshakeTimeout:   10 * time.Second,
	//	ExpectContinueTimeout: 1 * time.Second,
	//	MaxIdleConnsPerHost:   5,
	//	DisableKeepAlives:     true,
	//}}
	// httpClient := &http.Client{Transport: &http.Transport{}}

	subscribers := make(map[SubscriberID]*Subscriber)

	return &Client{
		Config:      config,
		Hosts:       hosts,
		Token:       NewToken(config.RefreshToken, config.RefreshTokenExp),
		Client:      httpClient,
		Subscribers: &subscribers,
		Websocket:   NewWebsocket(hosts.Websocket),
	}
}

func (c *Client) Connect(ctx context.Context, websocket bool) error {
	err := c.RefreshToken()
	if err != nil {
		return err
	}

	if websocket {
		err = c.Websocket.Connect(ctx)
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

//func (c *Client) AddSubscriberOLD(subscriber *Subscriber) error {
//
//	log.Println("subscriber ", subscriber.ID, "start subscriptions")
//	for opcode, subscription := range subscriber.Subscriptions {
//		switch opcode {
//		case BarsOpcode:
//			_ = c.BarsSubscribe(subscriber.ID, subscription)
//		case AllTradesOpcode:
//			_ = c.AllTradesSubscribe(subscriber.ID, subscription)
//		case OrderBookOpcode:
//			_ = c.OrderBooksSubscribe(subscriber.ID, subscription)
//		default:
//			continue
//		}
//	}
//
//	log.Println("subscriber ", subscriber.ID, "init")
//	if err := subscriber.CustomHandler.Init(); err != nil {
//		return err
//	}
//
//	if err := c.AddSubscriber(subscriber); err != nil {
//		return err
//	}
//
//	//c.mu.Lock()
//	//defer c.mu.Unlock()
//	//
//	//c.Subscribers[SubscriberID(subscriber.ID)] = subscriber
//
//	return nil
//}

//func (c *Client) RemoveSubscriberOLD(subscriberID uuid.UUID) error {
//	subscriber, ok := c.Websocket.subscribers[SubscriberID(subscriberID)]
//	if !ok {
//		// TODO: Error
//		return nil
//	}
//
//	// Больше не принимает события
//	subscriber.SetDone()
//
//	// Отписывается от всех подписок
//	for _, subscription := range subscriber.Subscriptions {
//		_ = c.Unsubscribe(subscriberID, subscription.Guid)
//		// TODO: Error
//	}
//
//	// Завершает handler
//	_ = subscriber.CustomHandler.DeInit()
//	// TODO: Error
//
//	c.mu.Lock()
//	defer c.mu.Unlock()
//
//	delete(c.Subscribers, SubscriberID(subscriberID))
//
//	return nil
//}

func (c *Client) GetSubscriber(subscriberID uuid.UUID) (*Subscriber, error) {
	subscriber, ok := (*c.Subscribers)[SubscriberID(subscriberID)]
	if !ok {
		return nil, errors.New("subscriber not found")
	}

	return subscriber, nil
}

func (c *Client) AddSubscriber(subscriber *Subscriber) error {
	log.Println("subscriber ", subscriber.ID, "start subscriptions", subscriber)
	for opcode, _ := range subscriber.Subscriptions {
		switch opcode {
		case BarsOpcode:
			if err := c.BarsSubscribe(subscriber); err != nil {
				return err
			}
		case AllTradesOpcode:
			if err := c.AllTradesSubscribe(subscriber); err != nil {
				return err
			}
		case OrderBookOpcode:
			if err := c.OrderBooksSubscribe(subscriber); err != nil {
				return err
			}
		default:
			continue
		}
	}

	log.Println("subscriber ", subscriber.ID, "init")
	if subscriber.CustomHandler != nil {
		if err := subscriber.CustomHandler.Init(); err != nil {
			return err
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	(*c.Subscribers)[SubscriberID(subscriber.ID)] = subscriber

	return nil
}

func (c *Client) RemoveSubscriber(subscriberID uuid.UUID) error {
	subscriber, ok := (*c.Subscribers)[SubscriberID(subscriberID)]
	if !ok {
		// TODO: Error
		return nil
	}

	// Больше не принимает события
	subscriber.SetDone()

	// Отписывается от всех подписок
	for _, subscription := range subscriber.Subscriptions {
		if err := c.Unsubscribe(subscriberID, subscription.Guid); err != nil {
			return err
		}
	}

	if subscriber.CustomHandler != nil {
		if err := subscriber.CustomHandler.DeInit(); err != nil {
			return err
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(*c.Subscribers, SubscriberID(subscriberID))

	return nil
}

func (c *Client) RemoveAllSubscribers() error {
	//for _, subscriber := range c.Websocket.subscribers {
	//	_ = c.RemoveSubscriber(subscriber.ID)
	//}

	for _, subscriber := range *c.Subscribers {
		_ = c.RemoveSubscriber(subscriber.ID)
	}

	return nil
}

func (c *Client) GetSubscribers() []*Subscriber {
	subscribers := make([]*Subscriber, 0)

	for _, subscriber := range *c.Subscribers {
		subscribers = append(subscribers, subscriber)
	}

	return subscribers
}

func (c *Client) GetAllSubscriberBars(subscriberID uuid.UUID) ([]*Bar, error) {
	subscriber, ok := (*c.Subscribers)[SubscriberID(subscriberID)]
	if !ok {
		return nil, errors.New("subscriber does not exist")
	}

	return subscriber.BarsProcessor.bars.GetAllBars(), nil
}
