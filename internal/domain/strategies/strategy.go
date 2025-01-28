package strategies

import (
	"context"
	"time"
)

type notifier interface {
	Info(message string)
	Warn(message string)
}

type subscriber interface {
	Subscribe()
	Unsubscribe()
	Subscriptions() map[string]string
}

func NewStrategy() Strategy {
	return Strategy{}
}

type Strategy struct {
	timeframe        string
	notifier         notifier
	subscriber       subscriber
	subscriptions    map[string]bool // [name]success
	candleCh         chan bool       //
	candleHandler    func() error
	eventsBuffer     [1000]StrategyEvent
	orderBookCh      chan bool //
	orderBookHandler func() error
	orderCh          chan bool // orders
	orderHandler     func() error
	candles          []string // candle
	orderFlow        bool
	marketProfile    bool
	depthOfMarket    bool
}

func (s Strategy) Run(ctx context.Context) error {
	for {
		select {
		case candle := <-s.candleCh:
			_ = s.candleHandler(candle)
		case orderBook := <-s.orderBookCh:
			_ = s.orderBookHandler(orderBook)
		case order := <-s.orderCh:
			_ = s.orderHandler(order)
		case <-time.After(time.Microsecond * 100):
			// check candle for end
			// финалим свечи по таймфрейму
		case <-ctx.Done():
			return nil
		}
	}

	return nil
}

func (s Strategy) Stop() error {
	return nil
}

func (s Strategy) AddCandleHandler() error {
	return nil
}

func (s Strategy) NewCandle() error {
	if s.candleHandler != nil {
		return s.candleHandler()
	}
	return nil
}

func (s Strategy) AddOrderBookHandler() error {
	return nil
}

func (s Strategy) NewOrderBook() error {
	if s.orderBookHandler != nil {
		return s.orderBookHandler()
	}
	return nil
}

func (s Strategy) AddOrderHandler() error {
	return nil
}

// from order flow
func (s Strategy) NewOrder() error {
	if s.orderFlow {
		// add order
	}

	if s.orderHandler != nil {
		return s.orderHandler()
	}
	return nil
}
