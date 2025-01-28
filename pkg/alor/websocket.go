package alor

import (
	"encoding/json"
	"github.com/google/uuid"
	"log"
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
	subscriptions map[string][]uuid.UUID
	subscribers   map[uuid.UUID]Subscriber
	counter       Counter
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

	// err := conn.WriteMessage(websocket.TextMessage, []byte("Hello from GolangDocs!"))
	// if err != nil {
	// 	log.Println("Error during writing to websocket:", err)
	// 	return
	// }
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

func (ws *Websocket) AddSubscriberToList(subscriberID uuid.UUID, guid string) {
	_, ok := ws.subscriptions[guid]
	if !ok {
		ws.subscriptions[guid] = make([]uuid.UUID, 0)
	}

	ws.subscriptions[guid] = append(ws.subscriptions[guid], subscriberID)
}

func (ws *Websocket) SendMessage(msg []byte) error {
	return ws.conn.WriteMessage(websocket.TextMessage, msg)
}

func (ws *Websocket) Unsubscribe(subscriberID uuid.UUID, guid string) {
	_, ok := ws.subscriptions[guid]
	if !ok {
		return
	}

	ws.subscriptions[guid] = ws.RemoveSubscriber(ws.subscriptions[guid], subscriberID)

	return
}

func (ws *Websocket) RemoveSubscriber(list []uuid.UUID, item uuid.UUID) []uuid.UUID {
	for i, v := range list {
		if v == item {
			list = append(list[:i], list[i+1:]...)
		}
	}
	return list
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

// Подписка на стакан
// {
// "opcode": "OrderBookGetAndSubscribe",
// "code": "SBER",
// "depth": 10,
// "exchange": "MOEX",
// "format": "Simple",
// "frequency": 0,
// "guid": "c328fcf1-e495-408a-a0ed-e20f95d6b813",
// "token": "eyJhbGciOiJ..."
// }

// Минимальное значение параметра Frequency зависит от выбранного формата возвращаемого JSON-объекта:
// Simple — 25 миллисекунд
// Slim — 10 миллисекунд
// Heavy — 500 миллисекунд
//type Request struct {
//	Token     string         `json:"token"`               // Access Токен для авторизации запроса
//	Code      string         `json:"code"`                // Код финансового инструмента (Тикер)
//	Depth     int            `json:"depth,omitempty"`     // Глубина стакана. Стандартное и максимальное значение — 20 (20х20).
//	Exchange  Exchange       `json:"exchange,omitempty"`  // Биржа
//	Format    ResponseFormat `json:"format"`              // Формат представления возвращаемых данных
//	Frequency int            `json:"frequency,omitempty"` // Частота (интервал) передачи данных сервером. Сервер вернёт последние данные по запросу за тот временной интервал, который указан в качестве значения параметра. Пример: биржа передаёт данные каждые 2 мс, но, при значении параметра 10 мс, сервер вернёт только последнее значение, отбросив предыдущие.
//	Opcode    Opcode         `json:"opcode"`              // Код выполняемой операции
//	Guid      string         `json:"guid"`                // Не более 50 символов. Уникальный идентификатор сообщений создаваемой подписки. Все входящие сообщения, соответствующие этой подписке, будут иметь такое значение поля guid
//}

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

type AllTradesData struct {
	Data struct {
		Id        int64     `json:"id"`
		Orderno   int64     `json:"orderno"`
		Symbol    string    `json:"symbol"`
		Board     string    `json:"board"`
		Qty       int64     `json:"qty"`
		Price     float64   `json:"price"`
		Time      time.Time `json:"time"`
		Timestamp int64     `json:"timestamp"`
		Oi        int64     `json:"oi"`
		Existing  bool      `json:"existing"`
		Side      OrderSide `json:"side"`
	} `json:"data"`
	Guid string `json:"guid"`
}

type OrderSide string

var (
	SellSide OrderSide = "Sell"
	BuySide  OrderSide = "Buy"
)
