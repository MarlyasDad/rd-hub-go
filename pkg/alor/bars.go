package alor

type Bar struct {
	High          float64
	Open          float64
	Close         float64
	Low           float64
	Volume        int64
	Delta         Delta
	MarketProfile MarketProfile
	OrderFlow     OrderFlow
}

type Delta struct {
}

type MarketProfile struct {
	POC float64
}

type OrderFlow struct {
}
