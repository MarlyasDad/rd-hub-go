package subscribers

import (
	"context"
	"encoding/json"
	"github.com/MarlyasDad/rd-hub-go/pkg/alor"
	"github.com/google/uuid"
	"log"
	"time"
)

// {
//    "description": "Если нужно описание, то оно должно быть туть",
//    // Параметры инструмента
//    "instrument": {
//        "exchange": "MOEX",
//        "code": "UWGN",
//        "board": "TQBR",
//        "timeframe": 300
//    },
//    // Параметры стратегии
//    "strategy": {
//        // Название стратегии, которую мы хотим использовать
//        // "strategy_name": "blank_strategy",
//        // "strategy_name": "proxy_strategy",
//        "name": "trailing_stop",
//        // Индивидуальные параметры стратегии (опционально, для каждой стратегии свои)
//        "settings": {
//            "close_under": 123.00,
//            "close_above": 456.00,
//            "slippage": 10,
//            "side": "buy",
//            "volume": 50
//        },
//        // Детализация данных (опционально, если не указано, то значения по-умолчанию)
//        "withDelta": true,
//        "withMarketProfile": true,
//        "withOrderBookProfile": true,
//        // Параметры подписок (опционально, если не указано, то значения по-умолчанию)
//        "subscriptions": {
//            "allTradesParams": {
//                "depth": 0,
//                "includeVirtualtrades": false
//            },
//            "orderBookParams": {
//                "depth": 10
//            },
//            "barsParams": {
//                "skipHitory": false,
//                "splitAdjust": true
//            }
//        },
//        // Индикаторы (опционально, если не указано, то не считаются)
//        "indicators": [
//            {
//                // Название индикатора, который мы хотим использовать
//                "name": "EMA",
//                // Индивидуальные параметры индикатора (опционально, для каждого индикатора свои)
//                "params": {
//                    "period": 42
//                }
//            }
//        ],
//        "async": false
//    }
//}

type AddSubscriberParams struct {
	Description string
	Instrument  InstrumentParams `json:"instrument"`
	Strategy    StrategyParams   `json:"strategy"`
	Async       bool             `json:"async"`
}

type InstrumentParams struct {
	Exchange  string
	Code      string
	Board     string
	Timeframe int64
}

type StrategyParams struct {
	Name              string          `json:"name"`
	Settings          json.RawMessage `json:"settings"`
	WithDelta         bool            `json:"with_delta"`
	WithMarketProfile bool            `json:"with_market_profile"`
	WithOrderFlow     bool            `json:"with_order_flow"`
}

type SubscriptionsParams struct {
}

func (s Service) AddSubscriber(ctx context.Context, params AddSubscriberParams) (alor.SubscriberID, error) { // FromAlltrades
	var options []alor.SubscriberOption

	// params

	if params.Strategy.WithDelta {
		options = append(options, alor.WithDelta())
	}

	if params.Strategy.WithMarketProfile {
		options = append(options, alor.WithMarketProfile())
	}

	if params.Strategy.WithOrderFlow {
		options = append(options, alor.WithOrderBookProfile())
	}

	options = append(options, alor.WithAllTradesSubscription(0, 50, false))
	options = append(options, alor.WithOrderBookSubscription(0, 10))
	// options = append(options, alor.WithCustomHandler(testHandler))

	// Подписчик создаётся с ready = false. Все новые события начинают накапливаться в очереди
	testSubscriber := alor.NewSubscriber(
		params.Description,
		alor.Exchange(params.Instrument.Exchange),
		params.Instrument.Code,
		params.Instrument.Board,
		alor.Timeframe(params.Instrument.Timeframe),
		false,
		options...,
	)
	//if err != nil {
	//	return uuid.Nil, err
	//}

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
		log.Println("get data for ", testSubscriber.ID, " offset ", historyParams.Offset)
		events, err := s.brokerClient.GetAllTrades(historyParams)
		if err != nil {
			return testSubscriber.ID, err
		}

		if len(events) == 0 {
			break
		}

		log.Println(len(events), "got events from", testSubscriber.ID, " offset ", historyParams.Offset)

		// Отправляем все данные в подписчика ====>
		for _, event := range events {
			// TODO: HANDLE UNIVESAL
			if err := testSubscriber.HandleHistoryAlltrades(event); err != nil {
				return testSubscriber.ID, err

			}
		}

		time.Sleep(100 * time.Millisecond)

		historyParams.Offset += 10000
	}
	log.Println("get data for ", testSubscriber.ID, " finished")

	// Активируем стратегии
	testSubscriber.SetReady()
	log.Printf("subscriber %s ready to work", testSubscriber.ID)

	// TODO: Получаем с захлёстом в 50 событий
	// Начинаем получать новые события
	if err := s.brokerClient.AddSubscriber(testSubscriber); err != nil {
		return alor.SubscriberID(uuid.Nil), err
	}

	log.Printf("subscriber %s successfully added", testSubscriber.ID)

	return testSubscriber.ID, nil
}
