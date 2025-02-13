package alor

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
)

type Subscription struct {
	Exchange        Exchange
	Code            string
	Board           string
	Tf              Timeframe
	Opcode          Opcode
	Depth           int
	From            int
	SkipHistory     bool
	SplitAdjust     bool
	InstrumentGroup string
	Format          ResponseFormat
	Guid            string
	Ready           bool
}

type OrderSide string

var (
	SellSide OrderSide = "Sell"
	BuySide  OrderSide = "Buy"
)

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

type Exchange string

var (
	MOEXExchange Exchange = "MOEX"
	SPBXExchange Exchange = "SPBX"
)

type ResponseFormat string

var (
	SimpleResponseFormat ResponseFormat = "Simple" // Оригинальный формат данных. Поддерживает устаревшие параметры для обеспечения обратной совместимости
	SlimResponseFormat   ResponseFormat = "Slim"   // Облегчённый формат данных для быстрой передачи сообщений. Не поддерживает устаревшие параметры
	HeavyResponseFormat  ResponseFormat = "Heavy"  // Полный формат данных, развивающийся и дополняющийся новыми полями. Не поддерживает устаревшие параметры
)

type Timeframe int

// 15 — 15 секунд
// 60 — 60 секунд или 1 минута
// 3600 — 3600 секунд или 1 час
// D — сутки (соответствует значению 86400)
// W — неделя (соответствует значению 604800)
// M — месяц (соответствует значению 2592000)
// Y — год (соответствует значению 31536000)

const (
	S15Timeframe Timeframe = 15
	M1Timeframe  Timeframe = 60
	H1Timeframe  Timeframe = 3600
	DTimeframe   Timeframe = 86400
	WTimeframe   Timeframe = 604800
	MTimeframe   Timeframe = 2592000
	YTimeframe   Timeframe = 31536000
)

type OrderBookRequest struct {
	Token     string         `json:"token"`               // Access Токен для авторизации запроса
	Code      string         `json:"code"`                // Код финансового инструмента (Тикер)
	Depth     int            `json:"depth,omitempty"`     // Глубина стакана. Стандартное значение — 20 (20х20), макс 50.
	Exchange  Exchange       `json:"exchange,omitempty"`  // Биржа
	Format    ResponseFormat `json:"format"`              // Формат представления возвращаемых данных
	Frequency int            `json:"frequency,omitempty"` // Частота (интервал) передачи данных сервером. Сервер вернёт последние данные по запросу за тот временной интервал, который указан в качестве значения параметра. Пример: биржа передаёт данные каждые 2 мс, но, при значении параметра 10 мс, сервер вернёт только последнее значение, отбросив предыдущие.
	Opcode    Opcode         `json:"opcode"`              // Код выполняемой операции
	Guid      string         `json:"guid"`                // Не более 50 символов. Уникальный идентификатор сообщений создаваемой подписки. Все входящие сообщения, соответствующие этой подписке, будут иметь такое значение поля guid
}

func (c *Client) OrderBooksSubscribe(subscriberID uuid.UUID, exchange Exchange, code string, depth int, format ResponseFormat) (string, error) {
	opcode := OrderBookOpcode
	guid := fmt.Sprintf("%s-%s-%s-%d-%s", exchange, code, opcode, depth, format)
	token, err := c.Authorization.AccessToken()
	if err != nil {
		return "", err
	}

	if depth > 50 || depth == 0 {
		depth = 20
	}

	var frequency int
	// Минимальное значение параметра Frequency зависит от выбранного формата возвращаемого JSON-объекта:
	// Simple — 25 миллисекунд
	// Slim — 10 миллисекунд
	// Heavy — 500 миллисекунд
	switch format {
	case SimpleResponseFormat:
		frequency = 25
	case SlimResponseFormat:
		frequency = 10
	case HeavyResponseFormat:
		frequency = 500
	}

	request := OrderBookRequest{
		Token:     token,
		Code:      code,
		Depth:     depth,
		Exchange:  exchange,
		Format:    format,
		Frequency: frequency,
		Opcode:    opcode,
		Guid:      guid,
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	err = c.Websocket.SendMessage(requestBytes)
	if err != nil {
		return "", err
	}

	c.AddSubscription(subscriberID, request.Guid)
	return request.Guid, nil
}

type OrderBookSimpleData struct {
	Snapshot    bool                   `json:"snapshot"`
	Bids        []OrderBookSimpleQuote `json:"bids"`
	Asks        []OrderBookSimpleQuote `json:"asks"`
	Timestamp   int                    `json:"timestamp"`
	MsTimestamp int                    `json:"ms_timestamp"`
	Depth       int                    `json:"depth"`
	Existing    bool                   `json:"existing"`
}

type OrderBookSimpleQuote struct {
	Price  float64 `json:"price"`
	Volume int64   `json:"volume"`
}

type OrderBookSlimData struct {
	Snapshot  bool                 `json:"h"`
	Bids      []OrderBookSlimQuote `json:"b"`
	Asks      []OrderBookSlimQuote `json:"a"`
	Timestamp int                  `json:"t"`
}

type OrderBookSlimQuote struct {
	Price  float64 `json:"p"`
	Volume int64   `json:"v"`
	Yield  int64   `json:"y"`
}

type OrderBookHeavyData struct {
	Bids        []OrderBookSimpleQuote `json:"bids"`
	Asks        []OrderBookSimpleQuote `json:"asks"`
	MsTimestamp int                    `json:"msTimestamp"`
	Depth       int                    `json:"depth"`
	Existing    bool                   `json:"existing"`
}

type OrderBookHeavyQuote struct {
	Price  float64 `json:"price"`
	Volume int64   `json:"volume"`
}

type AllTradesRequest struct {
	Opcode               Opcode         `json:"opcode"`               // Код выполняемой операции
	Depth                int            `json:"depth,omitempty"`      // Если указать, то перед актуальными данными придут данные о последних N сделках.
	IncludeVirtualTrades bool           `json:"includeVirtualTrades"` // Указывает, нужно ли отправлять виртуальные (индикативные) сделки
	Code                 string         `json:"code"`                 // Код финансового инструмента (Тикер)
	Exchange             Exchange       `json:"exchange,omitempty"`   // Биржа
	InstrumentGroup      string         `json:"instrumentGroup"`      // Код режима торгов (Борд). Для Биржи СПБ всегда SPBX
	Format               ResponseFormat `json:"format"`               // Формат представления возвращаемых данных
	Frequency            int            `json:"frequency,omitempty"`  // Частота (интервал) передачи данных сервером. Сервер вернёт последние данные по запросу за тот временной интервал, который указан в качестве значения параметра. Пример: биржа передаёт данные каждые 2 мс, но, при значении параметра 10 мс, сервер вернёт только последнее значение, отбросив предыдущие.
	Guid                 string         `json:"guid"`                 // Не более 50 символов. Уникальный идентификатор сообщений создаваемой подписки. Все входящие сообщения, соответствующие этой подписке, будут иметь такое значение поля guid
	Token                string         `json:"token"`                // Access Токен для авторизации запроса
}

func (c *Client) AllTradesSubscribe(subscriberID uuid.UUID, exchange Exchange, code string, depth int, format ResponseFormat) (string, error) {
	opcode := AllTradesOpcode
	guid := fmt.Sprintf("%s-%s-%s-%s", exchange, code, opcode, format)
	token, err := c.Authorization.AccessToken()
	if err != nil {
		return "", err
	}

	if depth > 50 || depth == 0 {
		depth = 50
	}

	var frequency int
	// Минимальное значение параметра Frequency зависит от выбранного формата возвращаемого JSON-объекта:
	// Simple — 25 миллисекунд
	// Slim — 10 миллисекунд
	// Heavy — 500 миллисекунд
	switch format {
	case SimpleResponseFormat:
		frequency = 25
	case SlimResponseFormat:
		frequency = 10
	case HeavyResponseFormat:
		frequency = 500
	}

	request := AllTradesRequest{
		Opcode:    opcode,
		Token:     token,
		Exchange:  exchange,
		Guid:      guid,
		Code:      code,
		Depth:     depth,
		Format:    format,
		Frequency: frequency,
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	err = c.Websocket.SendMessage(requestBytes)
	if err != nil {
		return "", err
	}

	c.AddSubscription(subscriberID, request.Guid)
	return request.Guid, nil
}

type AllTradesSimpleData struct {
	ID        int       `json:"id"`
	Orderno   int       `json:"orderno"`
	Symbol    string    `json:"symbol"`
	Board     string    `json:"board"`
	Qty       int       `json:"qty"`
	Price     float64   `json:"price"`
	Time      string    `json:"time"`
	Timestamp int       `json:"timestamp"`
	Oi        int       `json:"oi"`
	Existing  bool      `json:"existing"`
	Side      OrderSide `json:"side"`
}

type AllTradesSlimData struct {
	ID        int       `json:"id"`
	Symbol    string    `json:"sym"`
	Board     string    `json:"bd"`
	Qty       int       `json:"q"`
	Price     float64   `json:"px"`
	Timestamp int       `json:"t"`
	Oi        int       `json:"oi"`
	Existing  bool      `json:"h"`
	Side      OrderSide `json:"s"`
}

type AllTradesHeavyData struct {
	ID        int       `json:"id"`
	Symbol    string    `json:"symbol"`
	Board     string    `json:"board"`
	Qty       int       `json:"qty"`
	Price     float64   `json:"price"`
	Time      string    `json:"time"`
	Timestamp int       `json:"timestamp"`
	Oi        int       `json:"oi"`
	Existing  bool      `json:"existing"`
	Side      OrderSide `json:"side"`
}

type BarsRequest struct {
	Opcode          Opcode         `json:"opcode"`              // Код выполняемой операции
	Code            string         `json:"code"`                // Код финансового инструмента (Тикер)
	Tf              Timeframe      `json:"tf"`                  // Длительность таймфрейма. В качестве значения можно указать точное количество секунд или код таймфрейма
	From            int            `json:"from"`                // Дата и время (UTC) для первой запрашиваемой свечи
	SkipHistory     bool           `json:"skipHistory"`         // Флаг отсеивания исторических данных
	SplitAdjust     bool           `json:"splitAdjust"`         // Флаг коррекции исторических свечей инструмента с учётом сплитов, консолидаций и прочих факторов.
	Exchange        Exchange       `json:"exchange,omitempty"`  // Биржа
	InstrumentGroup string         `json:"instrumentGroup"`     // Код режима торгов (Борд). Для Биржи СПБ всегда SPBX
	Format          ResponseFormat `json:"format"`              // Формат представления возвращаемых данных
	Frequency       int            `json:"frequency,omitempty"` // Частота (интервал) передачи данных сервером. Сервер вернёт последние данные по запросу за тот временной интервал, который указан в качестве значения параметра. Пример: биржа передаёт данные каждые 2 мс, но, при значении параметра 10 мс, сервер вернёт только последнее значение, отбросив предыдущие.
	Guid            string         `json:"guid"`                // Не более 50 символов. Уникальный идентификатор сообщений создаваемой подписки. Все входящие сообщения, соответствующие этой подписке, будут иметь такое значение поля guid
	Token           string         `json:"token"`               // Access Токен для авторизации запроса
}

func (c *Client) BarsSubscribe(subscriberID uuid.UUID, exchange Exchange, code string, tf Timeframe, from int, skipHistory bool, splitAdjust bool, instrumentGroup string, format ResponseFormat) (string, error) {
	opcode := BarsOpcode
	guid := fmt.Sprintf("%s-%s-%d-%s-%s", exchange, code, tf, opcode, format)
	token, err := c.Authorization.AccessToken()
	if err != nil {
		return "", err
	}

	var frequency int
	// Минимальное значение параметра Frequency зависит от выбранного формата возвращаемого JSON-объекта:
	// Simple — 25 миллисекунд
	// Slim — 10 миллисекунд
	// Heavy — 500 миллисекунд
	switch format {
	case SimpleResponseFormat:
		frequency = 25
	case SlimResponseFormat:
		frequency = 10
	case HeavyResponseFormat:
		frequency = 500
	}

	request := BarsRequest{
		Opcode:          opcode,
		Code:            code,
		Tf:              tf,
		From:            from,
		SkipHistory:     skipHistory,
		SplitAdjust:     splitAdjust,
		Exchange:        exchange,
		InstrumentGroup: instrumentGroup,
		Format:          format,
		Frequency:       frequency,
		Guid:            guid,
		Token:           token,
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	err = c.Websocket.SendMessage(requestBytes)
	if err != nil {
		return "", err
	}

	c.AddSubscription(subscriberID, request.Guid)
	return request.Guid, nil
}

type BarsSimpleData struct {
	Time   string  `json:"time"`
	Close  float64 `json:"close"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Volume int     `json:"volume"`
}

type BarsSlimData struct {
	Time   string  `json:"t"`
	Close  float64 `json:"c"`
	Open   float64 `json:"o"`
	High   float64 `json:"h"`
	Low    float64 `json:"l"`
	Volume int     `json:"v"`
}

type BarsHeavyData struct {
	Time   string  `json:"time"`
	Close  float64 `json:"close"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Volume int     `json:"volume"`
}

type UnsubscribeRequest struct {
	Opcode Opcode `json:"opcode"`
	Token  string `json:"token"`
	GUID   string `json:"guid"`
}

func (c *Client) Unsubscribe(subscriberID uuid.UUID, guid string) error {
	token, err := c.Authorization.AccessToken()
	if err != nil {
		return err
	}

	err = c.RemoveSubscription(subscriberID, guid)
	if err != nil {
		return err
	}

	if len(c.Websocket.subscriptions[GUID(guid)].Items) > 0 {
		return nil
	}

	delete(c.Websocket.subscriptions, GUID(guid))

	request := UnsubscribeRequest{
		Opcode: UnsubscribeOpcode,
		Token:  token,
		GUID:   guid,
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	return c.Websocket.SendMessage(requestBytes)
}

/*
Остальные подписки реализовать по подобию
Trades, Quotes, ......
*/

func (c *Client) AddSubscription(subscriberID uuid.UUID, guid string) {
	c.Websocket.AddSubscription(subscriberID, guid)
}

func (c *Client) RemoveSubscription(subscriberID uuid.UUID, guid string) error {
	return c.Websocket.RemoveSubscription(subscriberID, guid)
}

func (c *Client) RemoveAllSubscriberSubscriptions(subscriberID uuid.UUID) error {
	return c.Websocket.RemoveAllSubscriberSubscriptions(subscriberID)
}

func (c *Client) AddSubscriber(subscriber *Subscriber) {
	c.Websocket.AddSubscriber(subscriber)
	// Записать guid в subscription
	// Записать статус подписки

	var guid string

	for key, subscription := range subscriber.subscriptions {
		switch subscription.Opcode {
		case BarsOpcode:
			guid, _ = c.BarsSubscribe(
				subscriber.ID,
				subscription.Exchange,
				subscription.Code,
				subscription.Tf,
				subscription.From,
				subscription.SkipHistory,
				subscription.SplitAdjust,
				subscription.InstrumentGroup,
				subscription.Format,
			)
		case AllTradesOpcode:
			guid, _ = c.AllTradesSubscribe(
				subscriber.ID,
				subscription.Exchange,
				subscription.Code,
				subscription.Depth,
				subscription.Format,
			)
		case OrderBookOpcode:
			guid, _ = c.OrderBooksSubscribe(
				subscriber.ID,
				subscription.Exchange,
				subscription.Code,
				subscription.Depth,
				subscription.Format,
			)
		default:
			continue
		}
		c.AddSubscription(subscriber.ID, guid)
		subscriber.subscriptions[key].Guid = guid
	}
}
