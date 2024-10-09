package alor

import (
	"context"
	"log"
	"net/http"
)

type AlorConnector struct {
	Config        Config
	Hosts         AlorHosts
	Authorization Authorization
	Client        int64
	Websocket     int64
}

func New(config Config) *AlorConnector {
	client := http.Client{}

	var hosts AlorHosts

	if config.DevCircuit {
		hosts = circuits.Development
	} else {
		hosts = circuits.Production
	}

	token := Token{
		Refresh:           config.RefreshToken,
		RefreshExpiration: config.RefreshTokenExp,
	}

	return &AlorConnector{
		Authorization: NewAuthorization(hosts.Authorization, client, token),
		Hosts:         hosts,
	}
}

func (c *AlorConnector) Start(ctx context.Context) {
	c.Authorization.Refresh()
	log.Println(c.Authorization.Token.Info)
	// c.Websocket.Connect()
}

func (c *AlorConnector) Stop() {
	// c.Websocket.Disconnect()
}
