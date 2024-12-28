package alor

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func NewWebsocket(url string) *Websocket {
	return &Websocket{
		url: url,
	}
}

type Callback func(response Response) error

type Websocket struct {
	url  string
	conn *websocket.Conn

	done chan interface{}

	callbacks map[Opcode]Callback

	active  map[string]string // active subscriptions
	waiting map[string]string // subscriptions pending confirmation
	failed  map[string]string

	counter Counter
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

func (ws *Websocket) SetCallback(opcode Opcode, callback Callback) {
	ws.callbacks[opcode] = callback
}

func (ws *Websocket) Subscribe(request Request) error {
	// TODO: Add to waiting

	// err := conn.WriteMessage(websocket.TextMessage, []byte("Hello from GolangDocs!"))
	// if err != nil {
	// 	log.Println("Error during writing to websocket:", err)
	// 	return
	// }

	return nil
}

func (ws *Websocket) Unsubscribe(request UnsubscribeRequest) error {
	// TODO: Delete from Active or not here?

	// err := conn.WriteMessage(websocket.TextMessage, []byte("Hello from GolangDocs!"))
	// if err != nil {
	// 	log.Println("Error during writing to websocket:", err)
	// 	return
	// }

	return nil
}

// HandleResponse
// Ответы от websocket-a бывают двух видов:
// - подтверждение подписки
// - данные
func (ws *Websocket) HandleResponse(msg []byte) error {
	ws.counter.Add()
	// получаем событие
	// превратить текст в объект и перенаправить в callback
	log.Printf("Received: %s\n", msg)

	go func() {
		defer ws.counter.Done()

		// msg -> Response
		// get type
		// guid or requestGuid

		// if guid
		// _, ok = ws.callbacks[Opcode]
		//
		// go ws.callbacks[Opcode](data)

		// ws.callbacks(Response{})
	}()

	// msg -> Response
	// get type
	// guid or requestGuid

	// if guid
	// _, ok = ws.callbacks[Opcode]
	//
	// go ws.callbacks[Opcode](data)

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
type Request struct {
	Token     string         `json:"token"`               // Access Токен для авторизации запроса
	Code      string         `json:"code"`                // Код финансового инструмента (Тикер)
	Depth     int            `json:"depth,omitempty"`     // Глубина стакана. Стандартное и максимальное значение — 20 (20х20).
	Exchange  Exchange       `json:"exchange,omitempty"`  // Биржа
	Format    ResponseFormat `json:"format"`              // Формат представления возвращаемых данных
	Frequency int            `json:"frequency,omitempty"` // Частота (интервал) передачи данных сервером. Сервер вернёт последние данные по запросу за тот временной интервал, который указан в качестве значения параметра. Пример: биржа передаёт данные каждые 2 мс, но, при значении параметра 10 мс, сервер вернёт только последнее значение, отбросив предыдущие.
	Opcode    Opcode         `json:"opcode"`              // Код выполняемой операции
	Guid      string         `json:"guid"`                // Не более 50 символов. Уникальный идентификатор сообщений создаваемой подписки. Все входящие сообщения, соответствующие этой подписке, будут иметь такое значение поля guid
}

type Opcode string

var (
	OrderBookOpcode     Opcode = "OrderBookGetAndSubscribe"     // Подписка на биржевой стакан
	BarsOpcode          Opcode = "BarsGetAndSubscribe"          // Подписка на историю цен (свечи)
	Quotes              Opcode = "QuotesSubscribe"              // Подписка на информацию о котировках
	InstrumentsOpcode   Opcode = "InstrumentsGetAndSubscribeV2" // Подписка на изменение информации о финансовых инструментах на выбранной бирже
	AllTradesOpcode     Opcode = "AllTradesGetAndSubscribe"     // Подписка на все сделки
	PositionsOpcode     Opcode = "PositionsGetAndSubscribeV2"   // Подписка на информацию о текущих позициях по торговым инструментам и деньгам
	SummariesOpcode     Opcode = "SummariesGetAndSubscribeV2"   // Подписка на сводную информацию по портфелю
	RisksOpcode         Opcode = "RisksGetAndSubscribe"         // Подписка на сводную информацию по портфельным рискам
	SpectralRisksOpcode Opcode = "SpectraRisksGetAndSubscribe"  // Подписка на информацию по рискам срочного рынка (FORTS)
	TradesOpcode        Opcode = "TradesGetAndSubscribeV2"      // Подписка на информацию о сделках
	OrdersOpcode        Opcode = "OrdersGetAndSubscribeV2"      // Подписка на информацию о текущих заявках на рынке для выбранных биржи и финансового инструмента
	StopOrdersOpcode    Opcode = "StopOrdersGetAndSubscribeV2"  // Подписка на информацию о текущих заявках на рынке для выбранных биржи и финансового инструмента
	UnsubscribeOpcode   Opcode = "Unsubscribe"                  // Отмена существующей подписки
)

type ResponseFormat string

var (
	SimpleResponseFormat ResponseFormat = "Simple" // Оригинальный формат данных. Поддерживает устаревшие параметры для обеспечения обратной совместимости
	SlimResponseFormat   ResponseFormat = "Slim"   // Облегчённый формат данных для быстрой передачи сообщений. Не поддерживает устаревшие параметры
	HeavyResponseFormat  ResponseFormat = "Heavy"  // Полный формат данных, развивающийся и дополняющийся новыми полями. Не поддерживает устаревшие параметры
)

type Exchange string

var (
	MOEXExchange Exchange = "MOEX"
	SPBXExchange Exchange = "SPBX"
)

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

type Response struct {
	Message     string          `json:"message"`
	Data        json.RawMessage `json:"data"`
	HttpCode    int             `json:"httpCode"`
	RequestGuid string          `json:"requestGuid"`
	Guid        string          `json:"guid"`
}

//{
//"data": {
//"snapshot": true,
//"bids": [
//{
//"price": 257.70,
//"volume": 157
//}
//],
//"asks": [
//{
//"price": 257.71,
//"volume": 288
//}
//],
//"timestamp": 1702631123,
//"ms_timestamp": 1702631123780,
//"existing": true
//},
//"guid": "c328fcf1-e495-408a-a0ed-e20f95d6b813"
//}

type OrderBookData struct {
	Snapshot    bool             `json:"snapshot"`
	Bids        []OrderBookQuote `json:"bids"`
	Asks        []OrderBookQuote `json:"asks"`
	Timestamp   int              `json:"timestamp"`
	MsTimestamp int              `json:"ms_timestamp"`
	Existing    bool             `json:"existing"`
	Guid        string           `json:"guid"`
}

type OrderBookQuote struct {
	Price  float64 `json:"price"`
	Volume int64   `json:"volume"`
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

type UnsubscribeRequest struct {
	Opcode Opcode `json:"opcode"`
	Token  string `json:"token"`
	GUID   string `json:"guid"`
}
