package alor

import (
	"github.com/google/uuid"
)

type AsyncSubscriber struct {
	ID               uuid.UUID
	queue            *Queue
	timeframe        Timeframe
	notifier         Notifier // Чтобы отправлять уведомления
	source           Source   // Чтобы делать заявки
	subscriptions    map[string]bool
	barsHandler      func() error // Чтобы обрабатывать бары
	orderBookHandler func() error // Чтобы обрабатывать стаканы
	orderHandler     func() error // Чтобы обрабатывать обезличенные сделки
	candles          []string     // candle
	autoBarMake      bool
	orderFlow        bool
	marketProfile    bool
	depthOfMarket    bool
}

func (s AsyncSubscriber) GetID() uuid.UUID {
	return s.ID
}

func (s AsyncSubscriber) HandleEvent(event Event) error {
	return s.queue.Enqueue(event)
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
