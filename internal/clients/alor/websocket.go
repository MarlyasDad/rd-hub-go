package alor

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func NewWsClient(url string) *WsClient {
	return &WsClient{
		url: url,
	}
}

type WsClient struct {
	url  string
	conn *websocket.Conn
	// subscribes map[string]string
	done chan interface{}
	// events   chan string
	callback func(event string) error
}

func (c *WsClient) receiveHandler(connection *websocket.Conn) {
	defer close(c.done)
	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			log.Println("Error in receive:", err)
			return
		}
		log.Printf("Received: %s\n", msg)
	}

	// err := conn.WriteMessage(websocket.TextMessage, []byte("Hello from GolangDocs!"))
	// if err != nil {
	// 	log.Println("Error during writing to websocket:", err)
	// 	return
	// }
}

func (c *WsClient) Open() {
	c.done = make(chan interface{}) // Channel to indicate that the receiverHandler is done

	// socketUrl := "ws://localhost:8080" + "/socket"
	conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}

	// defer conn.Close()
	c.conn = conn

	go c.receiveHandler(c.conn)
}

func (c *WsClient) Close() {
	// Close our websocket connection
	err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("Error during closing websocket:", err)
		return
	}

	select {
	case <-c.done:
		log.Println("Receiver Channel Closed! Exiting....")
	case <-time.After(time.Duration(1) * time.Second):
		log.Println("Timeout in closing receiving channel. Exiting....")
	}

	c.conn.Close()
}

func (c *WsClient) AddCallback(callback string) {
	c.callback = nil
}

func (c *WsClient) Subscribe(security string) {
	// err := conn.WriteMessage(websocket.TextMessage, []byte("Hello from GolangDocs!"))
	// if err != nil {
	// 	log.Println("Error during writing to websocket:", err)
	// 	return
	// }
}

func (c *WsClient) Unsubscribe(security string) {
	// err := conn.WriteMessage(websocket.TextMessage, []byte("Hello from GolangDocs!"))
	// if err != nil {
	// 	log.Println("Error during writing to websocket:", err)
	// 	return
	// }
}

func (c *WsClient) HandleEvent(security string) error {
	// получаем событие
	// тип события, тело события, пишем в очередь
	// превратить текст в объект и перенаправить в коллбэк
	event := ""
	return c.callback(event)
}
