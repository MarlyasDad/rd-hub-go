package alor

import (
	"strconv"
	"time"
)

type Bar struct {
	High          float64       `json:"high"`
	Open          float64       `json:"open"`
	Close         float64       `json:"close"`
	Low           float64       `json:"low"`
	Volume        int64         `json:"volume"`
	Time          time.Time     `json:"time"`
	Timestamp     int64         `json:"timestamp"`
	Delta         Delta         `json:"delta"`
	MarketProfile MarketProfile `json:"market_profile"`
	OrderFlow     OrderFlow     `json:"order_flow"`
	Indicators    []bool        `json:"indicators"`
}

type Delta struct {
	Buy   int64 `json:"buy"`
	Sell  int64 `json:"sell"`
	Total int64 `json:"total"`
}

// Setters

func (d *Delta) AddValue(count int64, side OrderSide) {
	if side == BuySide {
		d.Buy += count
		d.Total += count
	} else {
		d.Sell += count
		d.Total -= count
	}
}

type MarketProfileUnit struct {
	Buy   int64 `json:"buy"`
	Sell  int64 `json:"sell"`
	Total int64 `json:"total"`
}

type MarketProfile struct {
	POCVolume int64                        `json:"poc_volume"`
	POCPrice  float64                      `json:"poc_price"`
	Values    map[string]MarketProfileUnit `json:"values"`
}

func (mp *MarketProfile) AddValue(price float64, count int64, side OrderSide) {
	mapKey := Float64ToStringKey(price)
	if _, ok := mp.Values[mapKey]; !ok {
		mp.Values[mapKey] = MarketProfileUnit{}
	}

	unit := mp.Values[mapKey]

	if side == BuySide {
		unit.Buy += count
	} else {
		unit.Sell += count
	}

	unit.Total += count

	if unit.Total > mp.POCVolume {
		mp.POCPrice = price
		mp.POCVolume = unit.Total
	}

	mp.Values[mapKey] = unit
}

type OrderBookRow struct {
	Price  float64   `json:"price"`
	Volume int64     `json:"volume"`
	Side   OrderSide `json:"side"`
}

type OrderFlow struct {
	LastVal       map[string]OrderBookRow `json:"-"`
	ValuesInc     map[string]int64        `json:"values_inc"`
	ValuesDec     map[string]int64        `json:"values_dec"`
	TotalInc      int64                   `json:"total_inc"` // Сколько лотов всего появилось
	TotalDec      int64                   `json:"total_dec"` // Сколько лотов всего исчезло
	LastAsks      map[string]OrderBookRow `json:"last_asks"`
	ValuesAsksDec map[string]int64        `json:"values_asks_dec"`
	ValuesAsksInc map[string]int64        `json:"values_asks_inc"`
	TotalAsksInc  int64                   `json:"total_asks_inc"`
	TotalAsksDec  int64                   `json:"total_asks_dec"`
	LastBids      map[string]OrderBookRow `json:"last_bids"`
	ValuesBidsDec map[string]int64        `json:"values_bids_dec"`
	ValuesBidsInc map[string]int64        `json:"values_bids_inc"`
	TotalBidsInc  int64                   `json:"total_bids_inc"`
	TotalBidsDec  int64                   `json:"total_bids_dec"`
}

// Setters

func (of *OrderFlow) AddValue(asks []OrderBookSlimQuote, bids []OrderBookSlimQuote) {
	newVals := of.ConvertObToMap(asks, bids)
	asksMap := of.ConvertOBSlimToMap(asks)
	bidsMap := of.ConvertOBSlimToMap(bids)

	// Выясняем куда сдвинулся стакан
	minAsksMap := of.CalcAsksMin(asksMap)
	minLastAsksMap := of.CalcAsksMin(of.LastAsks)
	maxBidsMap := of.CalcBidsMax(bidsMap)
	maxLastBidsMap := of.CalcBidsMax(of.LastBids)

	if minAsksMap >= minLastAsksMap {
		// Если стакан сдвинулся вверх или на месте, считаем asks

		// Считаем asksMap на увеличение цены

		// Считаем asksMap на уменьшение цены
		of.CalcAsksMapDec(asksMap)
	}

	if maxBidsMap <= maxLastBidsMap {
		// Если стакан сдвинулся вниз или на месте, считаем bids

		// Считаем bidsMap на увеличение цены

		// Считаем bidsMap на уменьшение цены
		of.CalcBidsMapDec(bidsMap)
	}

	// Считаем общую активность по стакану

	// Тут считаем order_blocks на увеличении цены
	// Инкрементим всё, что появилось или увеличилось цикл по newVal
	for _, quote := range newVals {
		mapKey := Float64ToStringKey(quote.Price)

		lastV, ok := of.LastVal[mapKey]
		if !ok {
			// Если значение появилось
			of.TotalInc += quote.Volume
			of.ValuesInc[mapKey] += quote.Volume
		} else {
			// Если значение увеличилось
			if lastV.Volume < quote.Volume {
				deltaVolume := quote.Volume - lastV.Volume
				of.TotalInc += deltaVolume
				of.ValuesInc[mapKey] += deltaVolume
			}
		}
	}

	// Тут считаем order_blocks на уменьшении цены
	// Декрементим всё, что пропало или уменьшилось
	for _, quote := range of.LastVal {
		mapKey := Float64ToStringKey(quote.Price)

		newV, ok := newVals[mapKey]
		if !ok {
			// Если значение пропало
			of.TotalDec += quote.Volume
			of.ValuesDec[mapKey] += quote.Volume
		} else {
			// Если значение уменьшилось
			if newV.Volume < quote.Volume {
				deltaVolume := quote.Volume - newV.Volume
				of.TotalDec += deltaVolume
				of.ValuesDec[mapKey] += deltaVolume
			}
		}
	}

	of.LastVal = newVals
	of.LastBids = bidsMap
	of.LastAsks = asksMap
}

func (of *OrderFlow) ConvertObToMap(asks []OrderBookSlimQuote, bids []OrderBookSlimQuote) map[string]OrderBookRow {
	orderBook := make(map[string]OrderBookRow)

	for _, quote := range asks {
		mapKey := Float64ToStringKey(quote.Price)

		orderBook[mapKey] = OrderBookRow{
			Price:  quote.Price,
			Volume: quote.Volume,
			Side:   BuySide,
		}
	}

	for _, quote := range bids {
		mapKey := Float64ToStringKey(quote.Price)

		orderBook[mapKey] = OrderBookRow{
			Price:  quote.Price,
			Volume: quote.Volume,
			Side:   SellSide,
		}
	}

	return orderBook
}

func (of *OrderFlow) ConvertOBSlimToMap(rawOB []OrderBookSlimQuote) map[string]OrderBookRow {
	orderBook := make(map[string]OrderBookRow)

	for _, quote := range rawOB {
		mapKey := Float64ToStringKey(quote.Price)

		orderBook[mapKey] = OrderBookRow{
			Price:  quote.Price,
			Volume: quote.Volume,
			Side:   BuySide,
		}
	}
	return orderBook
}

func (of *OrderFlow) CalcAsksMapDec(asksMap map[string]OrderBookRow) {
	// Тут считаем order_blocks на уменьшении цены
	// Декрементим всё, что пропало или уменьшилось
	for _, quote := range of.LastAsks {
		mapKey := Float64ToStringKey(quote.Price)

		newV, ok := asksMap[mapKey]
		if !ok {
			// Если значение пропало
			of.TotalAsksDec += quote.Volume
			of.ValuesAsksDec[mapKey] += quote.Volume
		} else {
			// Если значение уменьшилось
			if newV.Volume < quote.Volume {
				deltaVolume := quote.Volume - newV.Volume
				of.TotalAsksDec += deltaVolume
				of.ValuesAsksDec[mapKey] += deltaVolume
			}
		}
	}
}

func (of *OrderFlow) CalcBidsMapDec(bidsMap map[string]OrderBookRow) {
	// Тут считаем order_blocks на уменьшении цены
	// Декрементим всё, что пропало или уменьшилось
	for _, quote := range of.LastBids {
		mapKey := Float64ToStringKey(quote.Price)

		newV, ok := bidsMap[mapKey]
		if !ok {
			// Если значение пропало
			of.TotalBidsDec += quote.Volume
			of.ValuesBidsDec[mapKey] += quote.Volume
		} else {
			// Если значение уменьшилось
			if newV.Volume < quote.Volume {
				deltaVolume := quote.Volume - newV.Volume
				of.TotalBidsDec += deltaVolume
				of.ValuesBidsDec[mapKey] += deltaVolume
			}
		}
	}
}

func (of *OrderFlow) CalcAsksMin(asksMap map[string]OrderBookRow) float64 {
	var minPrice float64

	for _, quote := range asksMap {
		if minPrice == 0 {
			minPrice = quote.Price
		}

		if quote.Price < minPrice {
			minPrice = quote.Price
		}
	}

	return minPrice
}

func (of *OrderFlow) CalcBidsMax(bidsMap map[string]OrderBookRow) float64 {
	var maxPrice float64

	for _, quote := range bidsMap {
		if maxPrice == 0 {
			maxPrice = quote.Price
		}

		if quote.Price > maxPrice {
			maxPrice = quote.Price
		}
	}

	return maxPrice
}

func Float64ToStringKey(value float64) string {
	return strconv.FormatFloat(value, 'f', 4, 64) // fmt.Sprint(price)
}
