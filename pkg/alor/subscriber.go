package alor

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"time"
)

type Notifier interface {
	Info(message string)
	Warn(message string)
	// SendMessage(clientId string, messageLevel INFO..., text string)
}

type Source interface {
	BarsSubscribe(subscriberID uuid.UUID, exchange Exchange, code string, tf Timeframe, from int, skipHistory bool, splitAdjust bool, instrumentGroup string, format ResponseFormat) (string, error)
	OrderBooksSubscribe(subscriberID uuid.UUID, exchange Exchange, code string, depth int, format ResponseFormat) (string, error)
	AllTradesSubscribe(subscriberID uuid.UUID, exchange Exchange, code string, depth int, format ResponseFormat) (string, error)
	Unsubscribe(subscriberID uuid.UUID, guid string) error
	// IsSubscriptionActive(guid string) bool
}

func NewSubscriber(subscriptions []Subscription, notifier Notifier, dataSource Source) *Subscriber {
	subscriberID := uuid.New()
	return &Subscriber{
		ID:            subscriberID,
		subscriptions: subscriptions,
		notifier:      notifier,   // Куда отправлять оповещения
		source:        dataSource, // Где заказывать данные
	}
}

type Subscriber struct {
	ID               uuid.UUID
	ClientID         uuid.UUID
	subscriptions    []Subscription
	notifier         Notifier // Чтобы отправлять уведомления
	source           Source   // Чтобы делать заявки
	queue            *Queue
	barsHandler      func() error     // Чтобы обрабатывать бары
	orderBookHandler func() error     // Чтобы обрабатывать стаканы
	orderHandler     func() error     // Чтобы обрабатывать обезличенные сделки
	bars             map[string][]Bar // bar
	indicators       []string         // indicator
	autoBarMake      bool
	depthOfMarket    bool
	marketProfile    bool
	delta            bool
	ready            bool
	done             bool
}

func (s Subscriber) GetID() uuid.UUID {
	return s.ID
}

func (s Subscriber) IsDone() bool {
	return s.done
}

func (s Subscriber) HandleEventSync(event Event) error {
	if !s.ready {
		// проверить все подписки
		// проверить все индикаторы

		// При обработке подписок ws сообщать всем подписчикам, что подписка активна -> map[guid]true
	}

	switch event.Opcode {
	case BarsOpcode:
		return s.NewBar(event.Data)
	case AllTradesOpcode:
		return s.NewAllTrades(event.Data)
	case OrderBookOpcode:
		return s.NewOrderBook(event.Data)
	default:
		fmt.Println("no content")
		return nil
	}
}

func (s Subscriber) HandleEventAsync(event Event) error {
	// кто-то должен разбирать очередь и пушить в канал
	// канал должен обрабатываться в селекте
	return s.queue.Enqueue(event)
}

func (s Subscriber) AddInitHandler() error {
	return nil
}

func (s Subscriber) Init() error {
	for _, subscription := range s.subscriptions {
		if subscription.Opcode == BarsOpcode {
			_, _ = s.source.BarsSubscribe(s.ID, subscription.Exchange, subscription.Code, subscription.Tf, subscription.From, false, false, "", SlimResponseFormat)
		}
	}
	return nil
}

func (s Subscriber) DeInit() error {
	return nil
}

func (s Subscriber) AddCandleHandler() error {
	return nil
}

func (s Subscriber) NewBar(eventData json.RawMessage) error {
	if s.barsHandler != nil {
		return s.barsHandler()
	}
	return nil
}

func (s Subscriber) AddOrderBookHandler() error {
	return nil
}

func (s Subscriber) NewOrderBook(eventData json.RawMessage) error {
	var orderBookData OrderBookSlimData
	err := json.Unmarshal(eventData, &orderBookData)
	if err != nil {
		return err
	}

	log.Println(orderBookData)

	if s.autoBarMake && s.depthOfMarket {
		// добавить информацию к текущей свечке
	}

	if s.orderBookHandler != nil {
		return s.orderBookHandler()
	}
	return nil
}

func (s Subscriber) AddAllTradesHandler() error {
	return nil
}

func (s Subscriber) NewAllTrades(eventData json.RawMessage) error {
	var orderData AllTradesSlimData
	err := json.Unmarshal(eventData, &orderData)
	if err != nil {
		return err
	}

	log.Println(orderData)

	// from order flow - обезличенные сделки
	if s.autoBarMake {
		// Добавить основную информацию к свечке
		if s.marketProfile {
			// добавить информацию к текущей свечке
		}
		if s.delta {
			// добавить информацию к текущей свечке
		}
	}

	if s.orderHandler != nil {
		return s.orderHandler()
	}
	return nil
}

func (s Subscriber) CreateBar(eventTime time.Time) error {
	// Создаём новую свечу, когда старая завершена
	// Считаем время текущей свечи по tf
	// Двигаем слайс со свечками
	// Удаляем первый элемент если ёмкость превышена
	// Ёмкость берём как самый большой период индикаторов
	return nil
}

func (s Subscriber) GetLastBar(eventTime time.Time) (*Bar, error) {
	return &Bar{}, nil
}

//func (s Subscriber) Run(ctx context.Context) error {
//
//	for {
//		select {
//		case candle := <-s.candleCh:
//			_ = s.candleHandler(candle)
//		case orderBook := <-s.orderBookCh:
//			_ = s.orderBookHandler(orderBook)
//		case order := <-s.orderCh:
//			_ = s.orderHandler(order)
//		case <-time.After(time.Microsecond * 100):
//			// check candle for end
//			// финалим свечи по таймфрейму
//		case <-ctx.Done():
//			return nil
//		}
//	}
//
//	return nil
//}
