package alor

import (
	"errors"
)

// !!! интерфейс для dataProcessor, commandBus и messageBus
type Strategy interface {
	Handle(data interface{}, processor *DataProcessor, commandBus int64, messageBus int64) error
	SetDataProcessor(processor *DataProcessor)
	SetStorage(storage *Storage)
}

type BaseStrategy struct {
	Storage    *Storage
	Processor  *DataProcessor
	CommandBus int64
	MessageBus int64
}

func (s *BaseStrategy) SetDataProcessor(processor *DataProcessor) {
	s.Processor = processor
}

func (s *BaseStrategy) SetStorage(storage *Storage) {
	s.Storage = storage
}

type OrderBookStrategy struct {
	BaseStrategy
	Opcode     Opcode
	HandleFunc func(data OrderBookSlimData, processor *DataProcessor, commandBus int64, messageBus int64) error
}

func NewOrderBookStrategy(handleFunc func(data OrderBookSlimData, processor *DataProcessor, commandBus int64, messageBus int64) error) *OrderBookStrategy {
	return &OrderBookStrategy{
		Opcode:     OrderBookOpcode,
		HandleFunc: handleFunc,
	}
}

func (h *OrderBookStrategy) Handle(data interface{}, processor *DataProcessor, commandBus int64, messageBus int64) error {
	switch v := data.(type) {
	case OrderBookSlimData:
		return h.HandleFunc(v, processor, commandBus, messageBus)
	default:
		return errors.New("invalid data type")
	}
}

type AllTradesStrategy struct {
	BaseStrategy
	Opcode     Opcode
	HandleFunc func(data AllTradesSlimData, processor *DataProcessor, commandBus int64, messageBus int64) error
}

func NewAllTradesStrategy(handleFunc func(data AllTradesSlimData, processor *DataProcessor, commandBus int64, messageBus int64) error) *AllTradesStrategy {
	return &AllTradesStrategy{
		Opcode:     AllTradesOpcode,
		HandleFunc: handleFunc,
	}
}

func (h *AllTradesStrategy) Handle(data interface{}, processor *DataProcessor, commandBus int64, messageBus int64) error {
	switch v := data.(type) {
	case AllTradesSlimData:
		return h.HandleFunc(v, processor, commandBus, messageBus)
	default:
		return errors.New("invalid data type")
	}
}

type BarsStrategy struct {
	BaseStrategy
	Opcode     Opcode
	HandleFunc func(data BarsSlimData, processor *DataProcessor, commandBus int64, messageBus int64) error
}

func NewBarsStrategy(handleFunc func(data BarsSlimData, processor *DataProcessor, commandBus int64, messageBus int64) error) *BarsStrategy {
	return &BarsStrategy{
		Opcode:     BarsOpcode,
		HandleFunc: handleFunc,
	}
}

func (h *BarsStrategy) Handle(data interface{}, processor *DataProcessor, commandBus int64, messageBus int64) error {
	switch v := data.(type) {
	case BarsSlimData:
		return h.HandleFunc(v, processor, commandBus, messageBus)
	default:
		return errors.New("invalid data type")
	}
}
