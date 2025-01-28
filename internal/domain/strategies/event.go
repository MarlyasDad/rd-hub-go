package strategies

type StrategyEvent struct {
	eventType string
	candle    *string
	order     *string
	orderBook *string
}

func (e StrategyEvent) GetType() string {
	return e.eventType
}

func (e StrategyEvent) GetBody() any {
	switch e.eventType.(type) {
	case CandleEvent:
		return e.candle
	case OrderEvent:
		return e.order
	case OrderBookEvent:
		return e.orderBook
	default:
		return nil
	}
}
