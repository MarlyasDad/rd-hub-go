package alor

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log"
	"sync"
	"time"
)

func NewSubscriber(description string, exchange Exchange, code string, board string, timeframe Timeframe, async bool, opts ...SubscriberOption) *Subscriber {
	s := &Subscriber{
		ID:            SubscriberID(uuid.New()),
		Description:   description,
		CreatedAt:     time.Now().UTC(),
		Exchange:      exchange,
		Code:          code,
		Board:         board,
		Timeframe:     timeframe,
		Storage:       newStorage(),
		DataProcessor: NewDataProcessor(timeframe),
		Subscriptions: make(map[Opcode]*Subscription),
		Strategy:      nil,
		Ready:         false,
		Async:         async,
		Queue:         NewChainQueue(10000),
		Done:          false,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// одна стратегия - один процессор
// Как делать хеджи?
// контейнер для стратегий, который ведёт себя как стратегия
// внутри выполняет логику хеджей и рулит несколькими стратегиями
// тогда стратегию нужно сделать интерфейсом!

// table Datasets in DB
// Да!

type SubscriberID uuid.UUID

func (sid SubscriberID) String() string {
	return uuid.UUID(sid).String()
}

// MarshalJSON реализует json.Marshaler
func (sid SubscriberID) MarshalJSON() ([]byte, error) {
	return json.Marshal(uuid.UUID(sid).String())
}

// UnmarshalJSON реализует json.Unmarshaler
func (sid *SubscriberID) UnmarshalJSON(data []byte) error {
	id, err := uuid.ParseBytes(data)
	if err != nil {
		return err
	}

	*sid = SubscriberID(id)
	return nil
}

type Subscriber struct {
	ID            SubscriberID             `json:"id"`
	Description   string                   `json:"description"`
	CreatedAt     time.Time                `json:"created_at"`
	Exchange      Exchange                 `json:"exchange"`
	Code          string                   `json:"code"`
	Board         string                   `json:"board"`
	Timeframe     Timeframe                `json:"timeframe"`
	Subscriptions map[Opcode]*Subscription `json:"subscriptions"` // Подписки на инструменты
	Ready         bool                     `json:"ready"`
	Storage       *Storage                 `json:"storage"` // Для передачи пользовательских состояний между обработчиками
	Strategy      Strategy                 `json:"-"`       // Стратегия
	DataProcessor *DataProcessor           `json:"-"`       // Бары, индикаторы, читает стратегию и добавляет индикаторы, можно добавлять пользовательские индикаторы
	Async         bool                     `json:"async"`   // Асинхронный режим
	Queue         *ChainQueue              `json:"queue"`   // Очередь для асинхронной обработки
	Done          bool                     `json:"done"`
	commandBus    *int
	messageBus    *int
	wg            sync.WaitGroup
}

func (s *Subscriber) Init() error { return nil }

func (s *Subscriber) DeInit() error { return nil }

// Options

type SubscriberOption func(strategy *Subscriber)

// Для DataProcessor

// WithDelta добавляем расчёт дельты
func WithDelta() SubscriberOption {
	return func(s *Subscriber) {
		s.DataProcessor.detailing.delta = true
	}
}

// WithMarketProfile добавляем расчёт профиля
func WithMarketProfile() SubscriberOption {
	return func(s *Subscriber) {
		s.DataProcessor.detailing.marketProfile = true
	}
}

// WithOrderBookProfile добавляем расчёт стакана
func WithOrderBookProfile() SubscriberOption {
	return func(s *Subscriber) {
		s.DataProcessor.detailing.orderBookProfile = true
	}
}

// WithIndicator добавляем расчёт индикатора
func WithIndicator(name string, opts SubscriberOption) SubscriberOption {
	return func(s *Subscriber) {
		// indicator := ...
		// if !ok "нет такого индикатора"
		// s.DataProcessor.Indicators = append(s.DataProcessor.Indicators, indicator)
	}
}

// Subscriptions

func WithAllTradesSubscription(frequency int, depth int, includeVirtualTrades bool) SubscriberOption {
	return func(s *Subscriber) {
		// GUID не меняется за всё время существования подписки
		guid := fmt.Sprintf("%s-%s-%s-%s-%s",
			AllTradesOpcode,
			s.Exchange,
			s.Code,
			s.Board,
			SlimResponseFormat,
		)

		if depth > 50 {
			depth = 50
		}

		// Минимальное значение параметра Frequency зависит от выбранного формата возвращаемого JSON-объекта:
		// Simple — 25 миллисекунд
		// Slim — 10 миллисекунд
		// Heavy — 500 миллисекунд
		if frequency < 10 {
			frequency = 10
		}

		s.Subscriptions[AllTradesOpcode] = &Subscription{
			GUID:            GUID(guid),
			Exchange:        s.Exchange,
			Code:            s.Code,
			InstrumentGroup: s.Board,
			Opcode:          AllTradesOpcode,

			AllTradesParams: AllTradesParams{
				Frequency:            frequency,
				Depth:                depth, // Если указать, то перед актуальными данными придут данные о последних N сделках.
				IncludeVirtualTrades: includeVirtualTrades,
			},
		}
	}
}

func WithOrderBookSubscription(frequency int, depth int) SubscriberOption {
	return func(s *Subscriber) {
		// GUID не меняется за всё время существования подписки
		guid := fmt.Sprintf(
			"%s-%s-%s-%s-%d-%s",
			OrderBookOpcode,
			s.Exchange,
			s.Code,
			s.Board,
			depth,
			SlimResponseFormat,
		)

		if depth > 50 || depth == 0 {
			depth = 10
		}

		// Минимальное значение параметра Frequency зависит от выбранного формата возвращаемого JSON-объекта:
		// Simple — 25 миллисекунд
		// Slim — 10 миллисекунд
		// Heavy — 500 миллисекунд
		if frequency < 10 {
			frequency = 10
		}

		s.Subscriptions[OrderBookOpcode] = &Subscription{
			GUID:            GUID(guid),
			Exchange:        s.Exchange,
			Code:            s.Code,
			InstrumentGroup: s.Board,
			Opcode:          OrderBookOpcode,
			OrderBookParams: OrderBookParams{
				Depth:     depth,
				Frequency: frequency,
			},
		}
	}
}

func WithBarsSubscription(frequency int, from int64, skipHistory, splitAdjust bool) SubscriberOption {
	// from := int(time.Now().Add(time.Hour*-24).Unix()))
	return func(s *Subscriber) {
		// GUID не меняется за всё время существования подписки
		guid := fmt.Sprintf(
			"%s-%s-%s-%s-%d-%s",
			BarsOpcode,
			s.Exchange,
			s.Code,
			s.Board,
			s.Timeframe,
			SlimResponseFormat,
		)

		// Минимальное значение параметра Frequency зависит от выбранного формата возвращаемого JSON-объекта:
		// Simple — 25 миллисекунд
		// Slim — 10 миллисекунд
		// Heavy — 500 миллисекунд
		if frequency < 10 {
			frequency = 10
		}

		s.Subscriptions[BarsOpcode] = &Subscription{
			GUID:            GUID(guid),
			Exchange:        s.Exchange,
			Code:            s.Code,
			InstrumentGroup: s.Board,
			Opcode:          BarsOpcode,
			BarsParams: BarsParams{
				Timeframe:   s.Timeframe,
				From:        from,
				SkipHistory: skipHistory,
				SplitAdjust: splitAdjust,
				Frequency:   frequency,
			},
		}
	}
}

func WithAsyncHandle() SubscriberOption {
	return func(s *Subscriber) {
		s.Async = true
	}
}

func WithStrategy(strategy Strategy) SubscriberOption {
	return func(s *Subscriber) {
		s.Strategy = strategy
	}
}

func WithStorage(storage *Storage) SubscriberOption {
	return func(s *Subscriber) {
		s.Storage = storage
	}
}

func (s *Subscriber) HandleEventSync(event *ChainEvent) error {
	// defer fmt.Println("Opcode: ", event.Opcode)

	switch event.Opcode {
	case BarsOpcode:
		var barsData BarsSlimData
		err := json.Unmarshal(event.Data, &barsData)
		if err != nil {
			return err
		}

		err = s.DataProcessor.NewBar(barsData)
		if err != nil {
			return err
		}

		if s.Ready && s.Strategy != nil {
			if err := s.Strategy.Handle(BarsOpcode, barsData); err != nil {
				if errors.Is(err, ErrNoAvailableHandler) {
					// log.Println(fmt.Errorf("%w for Opcode: %s", err, BarsOpcode))
					return nil
				}

				return err
			}
		}
	case AllTradesOpcode:
		var allTradesData AllTradesSlimData
		err := json.Unmarshal(event.Data, &allTradesData)
		if err != nil {
			log.Println("cannot unmarshal Alltrades", string(event.Data))
			return err
		}

		if err := s.DataProcessor.NewAllTrades(allTradesData); err != nil {
			return err
		}

		if s.Ready && s.Strategy != nil {
			if err := s.Strategy.Handle(AllTradesOpcode, allTradesData); err != nil {
				if errors.Is(err, ErrNoAvailableHandler) {
					// log.Println(fmt.Errorf("%w for Opcode: %s", err, AllTradesOpcode))
					return nil
				}

				return err
			}
		}
	case OrderBookOpcode:
		var orderBookData OrderBookSlimData
		err := json.Unmarshal(event.Data, &orderBookData)
		if err != nil {
			return err
		}

		if err := s.DataProcessor.NewOrderBook(orderBookData); err != nil {
			return err
		}

		if s.Ready && s.Strategy != nil {
			if err := s.Strategy.Handle(OrderBookOpcode, orderBookData); err != nil {
				if errors.Is(err, ErrNoAvailableHandler) {
					// log.Println(fmt.Errorf("%w for Opcode: %s", err, OrderBookOpcode))
					return nil
				}

				return err
			}
		}
	}

	// fmt.Println("no content")
	return nil
}

func (s *Subscriber) SetStrategy(strategy Strategy) {
	strategy.SetDataProcessor(s.DataProcessor)
	strategy.SetStorage(s.Storage)
	s.Strategy = strategy
}

func (s *Subscriber) SetID(id uuid.UUID) {
	s.ID = SubscriberID(id)
}

// setReady Выставляет флаг готовности стратегии к торговле
func (s *Subscriber) setReady() {
	s.Ready = true
}

// setDone Выставляет флаг завершения работы
func (s *Subscriber) setDone() {
	s.Done = true
}

//func (s *Subscriber) SetWait() {
//	s.wg.Add(1)
//}
//
//func (s *Subscriber) ReleaseWait() {
//	s.wg.Done()
//}

// Getters

func (s *Subscriber) GetID() SubscriberID {
	return s.ID
}

func (s *Subscriber) IsDone() bool {
	return s.Done
}

func (s *Subscriber) GetBarsCloak() *BarQueue {
	return s.DataProcessor.bars
}

// Handlers

func (s *Subscriber) HandleEvent(event *ChainEvent) error {
	// Если подписчик завершён, то не обрабатываем новые данные
	if s.Done {
		return nil
	}

	if s.Async {
		return s.Queue.Enqueue(event)
	} else {
		return s.HandleEventSync(event)
	}
}

//func (s *Subscriber) HandleHistoryEvent(event *ChainEvent) error {
//	// Если подписчик завершён, то не обрабатываем новые данные
//	if s.Done {
//		return nil
//	}
//
//	if s.Async {
//		return s.Queue.Enqueue(event)
//	} else {
//		return s.HandleEventSync(event)
//	}
//}

func (s *Subscriber) HandleHistoryAlltrades(data AllTradesSlimData) error {
	// Если подписчик завершён, то не обрабатываем новые данные
	if s.Done {
		return nil
	}

	if s.Async {
		return s.DataProcessor.NewAllTrades(data)
	} else {
		return s.DataProcessor.NewAllTrades(data)
	}
}

//func (s Subscriber) Run(ctx context.Context) error {
//
//  for {
//    select {
//    case <-a.context.Done():
//        return
//    default:
//        return
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
