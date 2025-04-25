package barstofile

import "github.com/MarlyasDad/rd-hub-go/pkg/alor"

type Notifier interface {
	SendInfo(message string)
	SendError(message string)
	// SendMessage(clientId string, messageLevel INFO..., text string)
}

type Commands interface {
	MakeOrder() error
	// GetOrderDetails() error
}

type History interface {
	GetActiveBar() alor.Bar
	GetLastFinalizedBar() alor.Bar
	GetSpecificBar(index int64) alor.Bar
	GetBarsRange(start, end int64) []alor.Bar
	// GetIndicatorValue(barIndex int64, name string) int64
}
