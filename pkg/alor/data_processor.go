package alor

import (
	"errors"
	"fmt"
	"time"
)

type DataDetailing struct {
	delta, marketProfile, orderBookProfile, disableBars bool
}

func NewDataProcessor(timeframe Timeframe) *DataProcessor {
	return &DataProcessor{
		timeframe: timeframe,
		bars:      NewBarQueue(5000),
		lastBar:   nil,
	}
}

type DataProcessor struct {
	timeframe       Timeframe
	bars            *BarQueue
	Indicators      []string `json:"-"` // Индикаторы
	lastBar         *Bar
	lastAlltradesID int64
	detailing       DataDetailing
}

func (p *DataProcessor) GetLastBar() (*Bar, error) {
	if p.lastBar == nil {
		return nil, errors.New("no available bars")
	}

	return p.lastBar, nil
}

func (p *DataProcessor) NewAllTrades(data AllTradesSlimData) error {
	if !p.detailing.disableBars {
		return nil
	}

	if data.ID < p.lastAlltradesID {
		return nil
	} else {
		p.lastAlltradesID = data.ID
	}

	// лента, лента всех сделок, таблица всех сделок, alltrades, time and sales, T&S
	// log.Println("AllTrades", time.Unix(data.MsTimestamp-(data.MsTimestamp%int64(p.timeframe)), 0))
	eventBarTime := time.UnixMilli(data.Timestamp - (data.Timestamp % int64(p.timeframe*1000)))

	// Создаём бар если его нет или пришёл новый
	if p.lastBar == nil || p.lastBar.Time != eventBarTime {
		newBar := p.NewBarFromAllTradesData(eventBarTime, data)

		err := p.bars.Enqueue(newBar)
		if err != nil {
			return err
		}

		p.lastBar = newBar
	}

	// Заполняем текущий бар
	p.UpdateBarFromAllTradesData(p.lastBar, data)

	// Считаем дополнительные данные
	if p.detailing.delta {
		p.lastBar.Delta.AddValue(data.Qty, data.Side)
	}

	if p.detailing.marketProfile {
		p.lastBar.MarketProfile.AddValue(data.Price, data.Qty, data.Side)
	}

	return nil
}

func (p *DataProcessor) NewBarFromAllTradesData(eventBarTime time.Time, data AllTradesSlimData) *Bar {
	newBar := &Bar{
		Timestamp: eventBarTime.Unix(),
		Time:      eventBarTime,
		Open:      data.Price,
		High:      data.Price,
		Low:       data.Price,
		Close:     data.Price,
		Volume:    data.Qty,
		Delta: Delta{
			Buy:   0,
			Sell:  0,
			Total: 0,
		},
		MarketProfile: MarketProfile{
			POCVolume: 0,
			POCPrice:  0.0,
			Values:    make(map[string]MarketProfileUnit),
		},
		OrderFlow: OrderFlow{
			LastVal:   make(map[string]OrderBookRow),
			ValuesInc: make(map[string]int64),
			ValuesDec: make(map[string]int64),
			TotalInc:  0,
			TotalDec:  0,
		},
	}

	return newBar
}

func (p *DataProcessor) UpdateBarFromAllTradesData(lastBar *Bar, data AllTradesSlimData) {
	if lastBar.Open == 0 {
		lastBar.Open = data.Price
	}

	if p.lastBar.High < data.Price || p.lastBar.High == 0 {
		lastBar.High = data.Price
	}

	if p.lastBar.Low > data.Price || p.lastBar.Low == 0 {
		lastBar.Low = data.Price
	}

	lastBar.Close = data.Price
	lastBar.Volume += data.Qty
}

func (p *DataProcessor) NewOrderBook(data OrderBookSlimData) error {
	if !p.detailing.orderBookProfile {
		return nil
	}

	// log.Println("OrderBook", time.Unix(data.MsTimestamp-(data.MsTimestamp%int64(p.timeframe)), 0))
	// Дата приходит в UnixMilli
	orderBookTime := time.UnixMilli(data.MsTimestamp - (data.MsTimestamp % int64(p.timeframe*1000)))

	// Если свечи нет или стакан не от этой свечи, то пропускаем
	if p.lastBar == nil || p.lastBar.Time != orderBookTime {
		return nil
	}

	if p.detailing.orderBookProfile {
		p.lastBar.OrderFlow.AddValue(data.Asks, data.Bids)
	}

	return nil
}

func (p *DataProcessor) NewBar(data BarsSlimData) error {
	if p.detailing.disableBars {
		return nil
	}

	eventBarTime := time.Unix(data.Time-(data.Time%int64(p.timeframe)), 0)

	fmt.Println("New bart", eventBarTime.String())

	if eventBarTime != p.lastBar.Time {
		newBar := p.NewBarFromBarData(eventBarTime, data)
		_ = p.bars.Enqueue(newBar)
		p.lastBar = newBar

		return nil
	}

	p.UpdateBarFromBarData(p.lastBar, data)

	return nil
}

func (p *DataProcessor) NewBarFromBarData(eventBarTime time.Time, data BarsSlimData) *Bar {
	newBar := &Bar{
		Timestamp: eventBarTime.Unix(),
		Time:      eventBarTime,
		Open:      data.Open,
		High:      data.High,
		Low:       data.Low,
		Close:     data.Close,
		Volume:    data.Volume,
		Delta: Delta{
			Buy:   0,
			Sell:  0,
			Total: 0,
		},
		MarketProfile: MarketProfile{
			POCVolume: 0,
			POCPrice:  0.0,
			Values:    make(map[string]MarketProfileUnit),
		},
		OrderFlow: OrderFlow{
			LastVal:   make(map[string]OrderBookRow),
			ValuesInc: make(map[string]int64),
			ValuesDec: make(map[string]int64),
			TotalInc:  0,
			TotalDec:  0,
		},
	}

	return newBar
}

func (p *DataProcessor) UpdateBarFromBarData(lastBar *Bar, data BarsSlimData) {
	lastBar.High = data.High
	lastBar.Low = data.Low
	lastBar.Close = data.Close
	lastBar.Volume = data.Volume
}

func (p *DataProcessor) NewBlankBar(eventBarTime time.Time) *Bar {
	newBar := &Bar{
		Timestamp: eventBarTime.Unix(),
		Time:      eventBarTime,
		Open:      0,
		High:      0,
		Low:       0,
		Close:     0,
		Volume:    0,
		Delta: Delta{
			Buy:   0,
			Sell:  0,
			Total: 0,
		},
		MarketProfile: MarketProfile{
			POCVolume: 0,
			POCPrice:  0.0,
			Values:    make(map[string]MarketProfileUnit),
		},
		OrderFlow: OrderFlow{
			LastVal:   make(map[string]OrderBookRow),
			ValuesInc: make(map[string]int64),
			ValuesDec: make(map[string]int64),
			TotalInc:  0,
			TotalDec:  0,
		},
	}

	return newBar
}

func (p *DataProcessor) NewBlankBarFromPrevious(eventBarTime time.Time, barClose float64) *Bar {
	newBar := &Bar{
		Timestamp: eventBarTime.Unix(),
		Time:      eventBarTime,
		Open:      barClose,
		High:      barClose,
		Low:       barClose,
		Close:     barClose,
		Volume:    0,
		Delta: Delta{
			Buy:   0,
			Sell:  0,
			Total: 0,
		},
		MarketProfile: MarketProfile{
			POCVolume: 0,
			POCPrice:  0.0,
			Values:    make(map[string]MarketProfileUnit),
		},
		OrderFlow: OrderFlow{
			LastVal:   make(map[string]OrderBookRow),
			ValuesInc: make(map[string]int64),
			ValuesDec: make(map[string]int64),
			TotalInc:  0,
			TotalDec:  0,
		},
	}

	return newBar
}
