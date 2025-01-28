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
}

type Source interface {
	Subscribe(exchange Exchange, code string) // Возвращает guid
	Unsubscribe(guid string)
}

func NewSyncSubscriber(tf Timeframe, source Source, notifier Notifier) SyncSubscriber {
	subscriberID := uuid.New()
	return SyncSubscriber{
		ID:        subscriberID,
		timeframe: tf,
		source:    source,   // Где подписаться на данные
		notifier:  notifier, // Куда отправлять события
	}
}

type SyncSubscriber struct {
	ID               uuid.UUID
	timeframe        Timeframe
	notifier         Notifier // Чтобы отправлять уведомления
	source           Source   // Чтобы делать заявки
	subscriptions    map[string]bool
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
	// orderBook        OrderBook        // orderBook
}

func (s SyncSubscriber) GetID() uuid.UUID {
	return s.ID
}

func (s SyncSubscriber) HandleEvent(event Event) error {
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

func (s SyncSubscriber) AddInitHandler() error {
	return nil
}

func (s SyncSubscriber) Init() error {
	return nil
}

func (s SyncSubscriber) AddCandleHandler() error {
	return nil
}

func (s SyncSubscriber) NewBar(eventData json.RawMessage) error {
	if s.barsHandler != nil {
		return s.barsHandler()
	}
	return nil
}

func (s SyncSubscriber) AddOrderBookHandler() error {
	return nil
}

func (s SyncSubscriber) NewOrderBook(eventData json.RawMessage) error {
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

func (s SyncSubscriber) AddAllTradesHandler() error {
	return nil
}

func (s SyncSubscriber) NewAllTrades(eventData json.RawMessage) error {
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

func (s SyncSubscriber) CreateBar(eventTime time.Time) error {
	// Создаём новую свечу, когда старая завершена
	// Считаем время текущей свечи по tf
	// Двигаем слайс со свечками
	// Удаляем первый элемент если ёмкость превышена
	// Ёмкость берём как самый большой период индикаторов
	return nil
}

func (s SyncSubscriber) GetLastBar(eventTime time.Time) (*Bar, error) {
	return &Bar{}, nil
}
