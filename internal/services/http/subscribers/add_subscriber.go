package subscribers

import (
	"context"
	"encoding/json"
	"github.com/MarlyasDad/rd-hub-go/pkg/alor"
	"github.com/google/uuid"
	"log"
	"time"
)

type AddSubscriberParams struct {
	Description   string        `json:"description"`
	Instrument    Instrument    `json:"instrument"`
	Strategy      Strategy      `json:"strategy"`
	Subscriptions Subscriptions `json:"subscriptions"`
	Indicators    []Indicator   `json:"indicators"`
	Async         bool          `json:"async"`
}

type Instrument struct {
	Exchange  string `json:"exchange"`
	Code      string `json:"code"`
	Board     string `json:"board"`
	Timeframe int64  `json:"timeframe"`
}

type Strategy struct {
	Name                 string          `json:"name"`
	Settings             json.RawMessage `json:"settings"`
	WithDelta            bool            `json:"withDelta"`
	WithMarketProfile    bool            `json:"withMarketProfile"`
	WithOrderBookProfile bool            `json:"withOrderBookProfile"`
}

type Subscriptions struct {
	AllTrades *AllTradesParams `json:"allTrades"`
	OrderBook *OrderBookParams `json:"orderBook"`
	Bars      *BarsParams      `json:"bars"`
}

type AllTradesParams struct {
	Frequency            int  `json:"frequency"`
	Depth                int  `json:"depth"`
	IncludeVirtualTrades bool `json:"includeVirtualTrades"`
}

type OrderBookParams struct {
	Frequency int `json:"frequency"`
	Depth     int `json:"depth"`
}

type BarsParams struct {
	Frequency   int   `json:"frequency"`
	From        int64 `json:"from"`
	SkipHistory bool  `json:"skipHistory"`
	SplitAdjust bool  `json:"splitAdjust"`
}

type Indicator struct {
	Name     string          `json:"name"`
	Settings json.RawMessage `json:"settings"`
}

func (s Service) AddSubscriber(ctx context.Context, params *AddSubscriberParams) (alor.SubscriberID, error) {
	var options []alor.SubscriberOption

	if params.Strategy.WithDelta {
		options = append(options, alor.WithDelta())
	}

	if params.Strategy.WithMarketProfile {
		options = append(options, alor.WithMarketProfile())
	}

	if params.Strategy.WithOrderBookProfile {
		options = append(options, alor.WithOrderBookProfile())
	}

	if params.Subscriptions.AllTrades != nil {
		options = append(options, alor.WithAllTradesSubscription(params.Subscriptions.AllTrades.Frequency, 50, false))
	}

	if params.Subscriptions.OrderBook != nil {
		options = append(options, alor.WithOrderBookSubscription(params.Subscriptions.OrderBook.Frequency, 10))
	}

	if params.Subscriptions.Bars != nil {
		options = append(options, alor.WithBarsSubscription(params.Subscriptions.Bars.Frequency, 10, params.Subscriptions.Bars.SkipHistory, params.Subscriptions.Bars.SplitAdjust))
	}

	subscriber := alor.NewSubscriber(
		params.Description,
		alor.Exchange(params.Instrument.Exchange),
		params.Instrument.Code,
		params.Instrument.Board,
		alor.Timeframe(params.Instrument.Timeframe),
		params.Async,
		options...,
	)

	strategy, err := alor.NewStrategy(params.Strategy.Name, params.Strategy.Settings)
	if err != nil {
		return alor.SubscriberID{}, err
	}

	subscriber.SetStrategy(strategy)

	// TODO: Add indicators
	// for ...

	// GET данные прошлых сессий
	// Отправляем все данные в подписчика ====>

	// GET данные текущей сессии по alltrades
	from := time.Now().AddDate(0, 0, -1).Unix()
	historyParams := alor.GetAllTradesV2Params{
		Exchange:     alor.MOEXExchange,
		Symbol:       params.Instrument.Code,
		Board:        params.Instrument.Board,
		From:         &from,
		Descending:   false,
		JsonResponse: true,
		Offset:       0,
		Take:         10000,
		// FromID:       &fromID,
	}

	// Получаем данные от брокера <====
	for {
		// TODO: GET *ChainEvent
		log.Println("get data for ", subscriber.ID, " offset ", historyParams.Offset)
		events, err := s.brokerClient.GetAllTrades(historyParams)
		if err != nil {
			return subscriber.ID, err
		}

		if len(events) == 0 {
			break
		}

		log.Println(len(events), "got events from", subscriber.ID, " offset ", historyParams.Offset)

		// Отправляем все данные в подписчика ====>
		for _, event := range events {
			// TODO: HANDLE UNIVESAL
			if err := subscriber.HandleHistoryAlltrades(event); err != nil {
				return subscriber.ID, err

			}
		}

		time.Sleep(100 * time.Millisecond)

		historyParams.Offset += 10000
	}
	log.Println("get data for ", subscriber.ID, " finished")

	// Начинаем получать новые события
	if err := s.brokerClient.AddSubscriber(subscriber); err != nil {
		return alor.SubscriberID(uuid.Nil), err
	}

	log.Printf("subscriber %s successfully added", subscriber.ID)

	return subscriber.ID, nil
}
