package alor

import (
	"context"
	"github.com/google/uuid"
	"log"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	Config    Config
	Hosts     Hosts
	Token     Token
	Client    *http.Client
	Websocket *Websocket
	mu        sync.Mutex
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

	return &Client{
		Config:    config,
		Hosts:     hosts,
		Token:     NewToken(config.RefreshToken, config.RefreshTokenExp),
		Client:    httpClient,
		Websocket: NewWebsocket(hosts.Websocket),
	}
}

func (c *Client) Connect(ctx context.Context, websocket bool) error {
	err := c.RefreshToken()
	if err != nil {
		return err
	}

	if websocket {
		c.Websocket.StartHealthLoop(ctx, c.Token)
	}

	return nil
}

func (c *Client) Stop(websocket bool) {
	token, err := c.Token.GetAccessToken()
	if err != nil {
		return
	}

	if websocket {
		c.Websocket.StopHealthLoop()
	}

	if err := c.Websocket.RemoveAllSubscribers(token); err != nil {
		log.Println("ahalai mahalai")
	}

	if err := c.Websocket.Disconnect(); err != nil {
		log.Println("abra kadabra")
	}
}

func (c *Client) GetSubscriber(subscriberID uuid.UUID) (*Subscriber, error) {
	return c.Websocket.GetSubscriber(subscriberID)
}

func (c *Client) AddSubscriber(subscriber *Subscriber) error {
	token, err := c.Token.GetAccessToken()
	if err != nil {
		return err
	}

	return c.Websocket.AddSubscriber(token, subscriber)
}

func (c *Client) RemoveSubscriber(subscriberID uuid.UUID) error {
	token, err := c.Token.GetAccessToken()
	if err != nil {
		return err
	}

	return c.Websocket.RemoveSubscriber(token, subscriberID)
}

func (c *Client) RemoveAllSubscribers() error {
	token, err := c.Token.GetAccessToken()
	if err != nil {
		return err
	}

	return c.Websocket.RemoveAllSubscribers(token)
}

func (c *Client) GetSubscribers() []*Subscriber {
	return c.Websocket.GetSubscribers()
}

func (c *Client) GetAllSubscriberBars(subscriberID uuid.UUID) ([]*Bar, error) {
	return c.Websocket.GetAllSubscriberBars(subscriberID)
}
