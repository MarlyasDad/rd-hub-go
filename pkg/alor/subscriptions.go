package alor

import (
	"fmt"
)

func (c *Client) SubscribeOrderBook(exchange Exchange, code string) error {
	opcode := OrderBookOpcode
	guid := fmt.Sprintf("%s-%s-%s", exchange, code, opcode) // generate guid
	token, err := c.Authorization.AccessToken()
	if err != nil {
		return err
	}

	// WebsocketOrderBookRequest
	request := Request{
		Opcode:    opcode,
		Token:     token,
		Exchange:  exchange,
		Guid:      guid,
		Code:      code,
		Depth:     10,
		Format:    SimpleResponseFormat,
		Frequency: 100,
	}

	return c.Websocket.Subscribe(request)
}

func (c *Client) UnsubscribeOrderBook(guid string) error {
	token, err := c.Authorization.AccessToken()
	if err != nil {
		return err
	}

	request := UnsubscribeRequest{
		Opcode: UnsubscribeOpcode,
		Token:  token,
		GUID:   guid,
	}
	return c.Websocket.Unsubscribe(request)
}
