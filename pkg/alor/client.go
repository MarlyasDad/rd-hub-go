package alor

import (
	"context"
	"log"
	"net/http"
	"sync"
)

type Client struct {
	Config        Config
	Hosts         Hosts
	Authorization Authorization
	Client        *http.Client
	API           string
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
		API:           "",
	}
}

func (c *Client) Connect(ctx context.Context, websocket bool) error {
	// Get auth token
	err := c.Authorization.Refresh()
	if err != nil {
		return err
	}

	if websocket {
		// Create websocket connection
		err = c.Websocket.Connect()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) Stop() {
	err := c.Websocket.Disconnect()
	if err != nil {
		log.Println("abra")
	}
}
