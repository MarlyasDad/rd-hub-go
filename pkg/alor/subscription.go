package alor

import (
	"encoding/json"
)

type Subscription struct {
	GUID            GUID            // Идентификатор подписки
	Opcode          Opcode          // Код подписки
	Exchange        Exchange        // Биржа MOEX SPBX
	Code            string          // Тикер (Код финансового инструмента)
	InstrumentGroup string          // Борд
	AllTradesParams AllTradesParams // Параметры для обезличенных сделок
	OrderBookParams OrderBookParams // Параметры для стакана котировок
	BarsParams      BarsParams      // Параметры для баров
}

type AllTradesParams struct {
	Depth                int  // Если указать, то перед актуальными данными придут данные о последних N сделках.
	IncludeVirtualTrades bool // Указывает, нужно ли отправлять виртуальные (индикативные) сделки
	Frequency            int  // Частота (интервал) передачи данных сервером. Сервер вернёт последние данные по запросу за тот временной интервал, который указан в качестве значения параметра. Пример: биржа передаёт данные каждые 2 мс, но, при значении параметра 10 мс, сервер вернёт только последнее значение, отбросив предыдущие.
}

type OrderBookParams struct {
	Depth     int // Глубина стакана. Стандартное значение — 20 (20х20), макс 50.
	Frequency int // Частота (интервал) передачи данных сервером. Сервер вернёт последние данные по запросу за тот временной интервал, который указан в качестве значения параметра. Пример: биржа передаёт данные каждые 2 мс, но, при значении параметра 10 мс, сервер вернёт только последнее значение, отбросив предыдущие.
}

type BarsParams struct {
	Timeframe   Timeframe // Длительность таймфрейма. В качестве значения можно указать точное количество секунд или код таймфрейма
	From        int64     // Дата и время (UTC) для первой запрашиваемой свечи
	SkipHistory bool      // Флаг отсеивания исторических данных
	SplitAdjust bool      // Флаг коррекции исторических свечей инструмента с учётом сплитов, консолидаций и прочих факторов.
	Frequency   int       // Частота (интервал) передачи данных сервером. Сервер вернёт последние данные по запросу за тот временной интервал, который указан в качестве значения параметра. Пример: биржа передаёт данные каждые 2 мс, но, при значении параметра 10 мс, сервер вернёт только последнее значение, отбросив предыдущие.
}

type OrderSide string

var (
	SellSide OrderSide = "sell"
	BuySide  OrderSide = "buy"
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
	S15TF   Timeframe = 15
	M1TF    Timeframe = 60
	M5TF    Timeframe = 300
	M15TF   Timeframe = 900
	H1TF    Timeframe = 3600
	DayTF   Timeframe = 86400
	WeekTF  Timeframe = 604800
	MonthTF Timeframe = 2592000
	YearTF  Timeframe = 31536000
)

type OrderBookRequest struct {
	Token     string         `json:"token"`               // Access Токен для авторизации запроса
	Code      string         `json:"code"`                // Код финансового инструмента (Тикер)
	Depth     int            `json:"depth,omitempty"`     // Глубина стакана. Стандартное значение — 20 (20х20), макс 50.
	Exchange  Exchange       `json:"exchange,omitempty"`  // Биржа
	Format    ResponseFormat `json:"format"`              // Формат представления возвращаемых данных
	Frequency int            `json:"frequency,omitempty"` // Частота (интервал) передачи данных сервером. Сервер вернёт последние данные по запросу за тот временной интервал, который указан в качестве значения параметра. Пример: биржа передаёт данные каждые 2 мс, но, при значении параметра 10 мс, сервер вернёт только последнее значение, отбросив предыдущие.
	Opcode    Opcode         `json:"opcode"`              // Код выполняемой операции
	Guid      GUID           `json:"guid"`                // Не более 50 символов. Уникальный идентификатор сообщений создаваемой подписки. Все входящие сообщения, соответствующие этой подписке, будут иметь такое значение поля guid
}

func (ws *Websocket) prepareOrderBooksRequest(token Token, subscription *Subscription) ([]byte, error) {
	accessToken, err := token.GetAccessToken()
	if err != nil {
		return nil, err
	}

	request := OrderBookRequest{
		Code:      subscription.Code,
		Depth:     subscription.OrderBookParams.Depth,
		Exchange:  subscription.Exchange,
		Format:    SlimResponseFormat,
		Frequency: subscription.OrderBookParams.Frequency,
		Opcode:    subscription.Opcode,
		Guid:      subscription.GUID,
		Token:     accessToken,
	}

	return json.Marshal(request)
}

type OrderBookSimpleData struct {
	Bids        []OrderBookSimpleQuote `json:"bids"`
	Asks        []OrderBookSimpleQuote `json:"asks"`
	MsTimestamp int                    `json:"ms_timestamp"`
	Depth       int                    `json:"depth"`
	Existing    bool                   `json:"existing"`
	// Snapshot      bool                 `json:"snapshot"`   // Deprecated
	// MsTimestamp   int                  `json:"timestamp"`  // Deprecated

}

type OrderBookSimpleQuote struct {
	Price  float64 `json:"price"`
	Volume int64   `json:"volume"`
}

type OrderBookSlimData struct {
	Bids        []OrderBookSlimQuote `json:"b"`
	Asks        []OrderBookSlimQuote `json:"a"`
	MsTimestamp int64                `json:"t"`
	Existing    bool                 `json:"h"`
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
	Guid                 GUID           `json:"guid"`                 // Не более 50 символов. Уникальный идентификатор сообщений создаваемой подписки. Все входящие сообщения, соответствующие этой подписке, будут иметь такое значение поля guid
	Token                string         `json:"token"`                // Access Токен для авторизации запроса
}

func (ws *Websocket) prepareAllTradesRequest(token Token, subscription *Subscription) ([]byte, error) {
	accessToken, err := token.GetAccessToken()
	if err != nil {
		return nil, err
	}

	request := AllTradesRequest{
		Opcode:    subscription.Opcode,
		Exchange:  subscription.Exchange,
		Guid:      subscription.GUID,
		Code:      subscription.Code,
		Depth:     subscription.AllTradesParams.Depth,
		Format:    SlimResponseFormat,
		Frequency: subscription.AllTradesParams.Frequency,
		Token:     accessToken,
	}

	return json.Marshal(request)
}

type AllTradesSimpleData struct {
	ID        int64     `json:"id"`
	Orderno   int       `json:"orderno"`
	Symbol    string    `json:"symbol"`
	Board     string    `json:"board"`
	Qty       int64     `json:"qty"`
	Price     float64   `json:"price"`
	Time      string    `json:"time"`
	Timestamp int64     `json:"timestamp"`
	Oi        int       `json:"oi"`
	Existing  bool      `json:"existing"`
	Side      OrderSide `json:"side"`
}

type AllTradesSlimData struct {
	ID        int64     `json:"id"`
	EID       string    `json:"eid"`
	Symbol    string    `json:"sym"`
	Board     string    `json:"bd"`
	Qty       int64     `json:"q"`
	Price     float64   `json:"px"`
	Timestamp int64     `json:"t"`
	Oi        int64     `json:"oi"`
	Existing  bool      `json:"h"`
	Side      OrderSide `json:"s"`
}

type AllTradesHeavyData struct {
	ID        int64     `json:"id"`
	Symbol    string    `json:"symbol"`
	Board     string    `json:"board"`
	Qty       int64     `json:"qty"`
	Price     float64   `json:"price"`
	Time      string    `json:"time"`
	Timestamp int64     `json:"timestamp"`
	Oi        int       `json:"oi"`
	Existing  bool      `json:"existing"`
	Side      OrderSide `json:"side"`
}

type BarsRequest struct {
	Opcode          Opcode         `json:"opcode"`              // Код выполняемой операции
	Code            string         `json:"code"`                // Код финансового инструмента (Тикер)
	Tf              Timeframe      `json:"tf"`                  // Длительность таймфрейма. В качестве значения можно указать точное количество секунд или код таймфрейма
	From            int64          `json:"from"`                // Дата и время (UTC) для первой запрашиваемой свечи
	SkipHistory     bool           `json:"skipHistory"`         // Флаг отсеивания исторических данных
	SplitAdjust     bool           `json:"splitAdjust"`         // Флаг коррекции исторических свечей инструмента с учётом сплитов, консолидаций и прочих факторов.
	Exchange        Exchange       `json:"exchange,omitempty"`  // Биржа
	InstrumentGroup string         `json:"instrumentGroup"`     // Код режима торгов (Борд). Для Биржи СПБ всегда SPBX
	Format          ResponseFormat `json:"format"`              // Формат представления возвращаемых данных
	Frequency       int            `json:"frequency,omitempty"` // Частота (интервал) передачи данных сервером. Сервер вернёт последние данные по запросу за тот временной интервал, который указан в качестве значения параметра. Пример: биржа передаёт данные каждые 2 мс, но, при значении параметра 10 мс, сервер вернёт только последнее значение, отбросив предыдущие.
	Guid            GUID           `json:"guid"`                // Не более 50 символов. Уникальный идентификатор сообщений создаваемой подписки. Все входящие сообщения, соответствующие этой подписке, будут иметь такое значение поля guid
	Token           string         `json:"token"`               // Access Токен для авторизации запроса
}

func (ws *Websocket) prepareBarsRequest(token Token, subscription *Subscription) ([]byte, error) {
	accessToken, err := token.GetAccessToken()
	if err != nil {
		return nil, err
	}

	request := BarsRequest{
		Opcode:          subscription.Opcode,
		Code:            subscription.Code,
		Tf:              subscription.BarsParams.Timeframe,
		From:            subscription.BarsParams.From,
		SkipHistory:     subscription.BarsParams.SkipHistory,
		SplitAdjust:     subscription.BarsParams.SplitAdjust,
		Exchange:        subscription.Exchange,
		InstrumentGroup: subscription.InstrumentGroup,
		Format:          SlimResponseFormat,
		Frequency:       subscription.BarsParams.Frequency,
		Guid:            subscription.GUID,
		Token:           accessToken,
	}

	return json.Marshal(request)
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
	Time   int64   `json:"t"`
	Close  float64 `json:"c"`
	Open   float64 `json:"o"`
	High   float64 `json:"h"`
	Low    float64 `json:"l"`
	Volume int64   `json:"v"`
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
	GUID   GUID   `json:"guid"`
}

func (c *Client) Unsubscribe(subscriberID SubscriberID, guid GUID) error {
	token, err := c.Token.GetAccessToken()
	if err != nil {
		return err
	}

	return c.Websocket.Unsubscribe(token, subscriberID, guid)
}

/*
Остальные подписки реализовать по подобию
Trades, Quotes, ......
*/
