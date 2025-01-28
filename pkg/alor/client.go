package alor

import (
	"context"
	"net/http"
	"sync"
)

type Client struct {
	Config        Config
	Hosts         Hosts
	Authorization Authorization
	Client        *http.Client
	Websocket     *Websocket
	mu            *sync.Mutex
}

func New(config Config) *Client {
	client := http.Client{}

	var hosts Hosts

	if config.DevCircuit {
		hosts = circuits.Development
	} else {
		hosts = circuits.Production
	}

	return &Client{
		Hosts:         hosts,
		Authorization: NewAuthorization(hosts.Authorization, client, config.RefreshToken, config.RefreshTokenExp),
		Client:        &client,
		Websocket:     NewWebsocket(hosts.Websocket),
	}
}

func (c *Client) Start(ctx context.Context) {
	c.Authorization.Refresh()
	// log.Println(c.Authorization.Token.Info)
	c.Websocket.Connect()
}

func (c *Client) Stop() {
	c.Websocket.Disconnect()
}
