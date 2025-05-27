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

type CustomHandler interface {
	Init() error
	DeInit() error
	SetHistory(history *BarQueue)
	NewBar(data BarsSlimData) error
	NewAllTrades(data AllTradesSlimData) error
	NewOrderBook(data OrderBookSlimData) error
}

type Subscriber struct {
	ID            uuid.UUID                `json:"id"`
	Description   string                   `json:"description"`
	CreatedAt     time.Time                `json:"created_at"`
	Async         bool                     `json:"async"`
	Exchange      Exchange                 `json:"exchange"`
	Code          string                   `json:"code"`
	Board         string                   `json:"board"`
	Timeframe     Timeframe                `json:"timeframe"`
	Subscriptions map[Opcode]*Subscription `json:"subscriptions"`
	CustomHandler CustomHandler            `json:"-"`
	Queue         *ChainQueue              `json:"-"`
	BarsProcessor *BarsProcessor           `json:"-"`
	Ready         bool                     `json:"-"`
	Done          bool                     `json:"done"`
	wg            sync.WaitGroup
}

func NewSubscriber(description string, exchange string, code string, board string, timeframe int64, opts ...SubscriberOption) (*Subscriber, error) {
	// validate exchange
	validExchange := Exchange(exchange)
	// validate timeframe
	validTimeframe := Timeframe(timeframe)

	s := &Subscriber{
		ID:            uuid.New(),
		Description:   description,
		CreatedAt:     time.Now().UTC(),
		Async:         false,
		Exchange:      validExchange,
		Code:          code,
		Board:         board,
		Timeframe:     validTimeframe,
		Queue:         NewChainQueue(10000),
		BarsProcessor: NewBarsProcessor(validTimeframe),
		Subscriptions: make(map[Opcode]*Subscription),
		Ready:         false,
		Done:          false,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

// Options

type SubscriberOption func(subscriber *Subscriber)

func WithDeltaData() SubscriberOption {
	return func(s *Subscriber) {
		s.BarsProcessor.detailing.delta = true
		s.BarsProcessor.detailing.disableBars = true
	}
}

func WithMarketProfileData() SubscriberOption {
	return func(s *Subscriber) {
		s.BarsProcessor.detailing.marketProfile = true
		s.BarsProcessor.detailing.disableBars = true
	}
}

func WithOrderFlowData() SubscriberOption {
	return func(s *Subscriber) {
		s.BarsProcessor.detailing.orderFlow = true
	}
}

func WithAsyncHandle() SubscriberOption {
	return func(s *Subscriber) {
		s.Async = true
	}
}

func WithBarsSubscription(from int64, skipHistory, splitAdjust bool) SubscriberOption {
	// from := int(time.Now().Add(time.Hour*-24).Unix()))
	return func(s *Subscriber) {
		s.Subscriptions[BarsOpcode] = &Subscription{
			Exchange:        s.Exchange,
			Code:            s.Code,
			InstrumentGroup: s.Board,
			Timeframe:       s.Timeframe,
			Opcode:          BarsOpcode,
			SkipHistory:     skipHistory,
			SplitAdjust:     splitAdjust,
			Format:          SlimResponseFormat,
			From:            from,
		}
	}
}

func WithAllTradesSubscription(depth int, includeVirtualTrades bool) SubscriberOption {
	return func(s *Subscriber) {
		s.Subscriptions[AllTradesOpcode] = &Subscription{
			Exchange:             s.Exchange,
			Code:                 s.Code,
			InstrumentGroup:      s.Board,
			Timeframe:            s.Timeframe,
			Opcode:               AllTradesOpcode,
			Depth:                depth, // Если указать, то перед актуальными данными придут данные о последних N сделках.
			IncludeVirtualTrades: includeVirtualTrades,
			Format:               SlimResponseFormat,
		}
	}
}

func WithOrderBookSubscription(depth int) SubscriberOption {
	return func(s *Subscriber) {
		s.Subscriptions[OrderBookOpcode] = &Subscription{
			Exchange:        s.Exchange,
			Code:            s.Code,
			InstrumentGroup: s.Board,
			Timeframe:       s.Timeframe,
			Opcode:          OrderBookOpcode,
			Depth:           depth,
			Format:          SlimResponseFormat,
		}
	}
}

func WithCustomHandler(handler CustomHandler) SubscriberOption {
	return func(s *Subscriber) {
		s.CustomHandler = handler
		s.CustomHandler.SetHistory(s.BarsProcessor.bars)
	}
}

// Setters

//func (s Subscriber) SetSubscriptionGuid(opcode Opcode, guid string) {
//	_, ok := s.Subscriptions[opcode]
//	if !ok {
//		return
//	}
//	if entry, ok := s.Subscriptions[opcode]; ok {
//		entry.Guid = guid
//		s.Subscriptions[opcode] = entry
//	}
//}

func (s *Subscriber) SetID(id uuid.UUID) {
	s.ID = id
}

func (s *Subscriber) SetReady() {
	s.Ready = true
}

func (s *Subscriber) SetDone() {
	s.Done = true
}

func (s *Subscriber) SetWait() {
	s.wg.Add(1)
}

func (s *Subscriber) ReleaseWait() {
	s.wg.Done()
}

// Getters

func (s *Subscriber) GetID() uuid.UUID {
	return s.ID
}

func (s *Subscriber) IsDone() bool {
	return s.Done
}

// GetBarsCloak Получаем BarsCloak
func (s *Subscriber) GetBarsCloak() *BarQueue {
	return s.BarsProcessor.bars
}

// Handlers

func (s *Subscriber) HandleEvent(event *ChainEvent) error {
	if s.Done {
		return nil
	}

	s.wg.Wait()

	// Если не готовы обрабатывать, пишем в очередь и выходим
	if !s.Ready {
		_ = s.Queue.Enqueue(event)
		return nil
	}

	if s.Async {
		return s.handleEventAsync(event)
	} else {
		return s.handleEventSync(event)
	}
}

func (s *Subscriber) HandleHistoryEvent(event *ChainEvent) error {
	if s.Done {
		return nil
	}

	if s.Async {
		return s.handleEventAsync(event)
	} else {
		return s.handleEventSync(event)
	}
}

func (s *Subscriber) HandleHistoryAlltrades(data AllTradesSlimData) error {
	if s.Done {
		return nil
	}

	if s.Async {
		return s.BarsProcessor.NewAllTrades(data)
	} else {
		return s.BarsProcessor.NewAllTrades(data)
	}
}

func (s *Subscriber) handleEventSync(event *ChainEvent) error {
	// defer fmt.Println("Opcode: ", event.Opcode)

	switch event.Opcode {
	case BarsOpcode:
		var barsData BarsSlimData
		err := json.Unmarshal(event.Data, &barsData)
		if err != nil {
			return err
		}

		err = s.BarsProcessor.NewBar(barsData)
		if err != nil {
			return err
		}

		if s.CustomHandler != nil {
			if err := s.CustomHandler.NewBar(barsData); err != nil {
				return err
			}
		}

		return nil
	case AllTradesOpcode:
		var allTradesData AllTradesSlimData
		err := json.Unmarshal(event.Data, &allTradesData)
		if err != nil {
			log.Println("cannot unmarshal Alltrades", string(event.Data))
			return err
		}

		err = s.BarsProcessor.NewAllTrades(allTradesData)
		if err != nil {
			if errors.Is(err, ErrNewBarFound) {
				lastFinalBar, _ := s.BarsProcessor.bars.GetLastFinalizedBar()
				newBarData := BarsSlimData{
					Time:   lastFinalBar.Timestamp,
					Open:   lastFinalBar.Open,
					High:   lastFinalBar.High,
					Low:    lastFinalBar.Low,
					Close:  lastFinalBar.Close,
					Volume: lastFinalBar.Volume,
				}

				if s.CustomHandler != nil {
					if err := s.CustomHandler.NewBar(newBarData); err != nil {
						return err
					}
				}
				return nil
			}
			return err
		}

		if s.CustomHandler != nil {
			if err := s.CustomHandler.NewAllTrades(allTradesData); err != nil {
				return err
			}
		}

		return nil
	case OrderBookOpcode:
		var orderBookData OrderBookSlimData
		err := json.Unmarshal(event.Data, &orderBookData)
		if err != nil {
			return err
		}

		if err := s.BarsProcessor.NewOrderBook(orderBookData); err != nil {
			return err
		}

		if s.CustomHandler != nil {
			if err := s.CustomHandler.NewOrderBook(orderBookData); err != nil {
				return err
			}
		}

		return nil
	}

	fmt.Println("no content")
	return nil
}

func (s *Subscriber) handleEventAsync(event *ChainEvent) error {
	return s.Queue.Enqueue(event)
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
