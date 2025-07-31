package alor

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Strategy interface {
	Handle(opcode Opcode, data interface{}) error
	SetDataProcessor(processor *DataProcessor)
	SetStorage(storage *Storage)
}

func NewStrategy(name string, settings json.RawMessage) (Strategy, error) {
	if name != "" {
		name = "base"
	}

	switch strings.ToLower(name) {
	case "base":
		return &BaseStrategy{}, nil
	default:
		return nil, fmt.Errorf("unknown strategy: %s", name)
	}
}

// !!! интерфейс для dataProcessor, commandBus и messageBus
type BaseStrategy struct {
	Storage    *Storage
	Processor  *DataProcessor
	CommandBus int64
	MessageBus int64
	Handlers   map[Opcode]func(opcode Opcode, data interface{}, processor *DataProcessor, storage *Storage, commandBus, messageBus int64) error
}

func (s *BaseStrategy) SetDataProcessor(processor *DataProcessor) {
	s.Processor = processor
}

func (s *BaseStrategy) SetStorage(storage *Storage) {
	s.Storage = storage
}

func (s *BaseStrategy) Handle(opcode Opcode, data interface{}) error {
	_, ok := s.Handlers[opcode]
	if !ok {
		return ErrNoAvailableHandler
	}

	return s.Handlers[opcode](opcode, data, s.Processor, s.Storage, s.CommandBus, s.MessageBus)
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
