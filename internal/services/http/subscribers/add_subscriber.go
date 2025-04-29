package subscribers

import (
	"errors"
	"github.com/MarlyasDad/rd-hub-go/pkg/alor"
	"github.com/google/uuid"
	"log"
	"time"
)

func (s Service) AddSubscriber() (uuid.UUID, error) { // FromAlltrades
	var options []alor.SubscriberOption

	// testHandler := barsToFileCommand.New("UWGN.txt")

	options = append(options, alor.WithDeltaData())
	options = append(options, alor.WithMarketProfileData())
	options = append(options, alor.WithOrderFlowData())
	options = append(options, alor.WithAllTradesSubscription(0, false))
	options = append(options, alor.WithOrderBookSubscription(10))
	// options = append(options, alor.WithCustomHandler(testHandler))

	// Подписчик создаётся с ready = false. Все новые события начинают накапливаться в очереди
	testSubscriber := alor.NewSubscriber(
		"Test UWGN subscriber, timeframe M5, sync",
		alor.MOEXExchange,
		"UWGN",
		"TQBR",
		alor.M5TF,
		options...,
	)

	// Начинаем получать события
	err := s.brokerClient.AddSubscriber(testSubscriber)
	if err != nil {
		return testSubscriber.ID, err
	}

	// GET данные прошлых сессий
	// Отправляем все данные в подписчика ====>

	// GET данные текущей сессии
	from := time.Now().AddDate(0, 0, -1).Unix()
	params := alor.GetAllTradesV2Params{
		Exchange:     alor.MOEXExchange,
		Symbol:       "UWGN",
		Board:        "TQBR",
		From:         &from,
		Descending:   false,
		JsonResponse: true,
		Offset:       0,
		Take:         10000,
		// FromID:       &fromID,
	}

	// Получаем данные от брокера <====
	for {
		log.Println("get data for ", testSubscriber.ID, " offset ", params.Offset)
		events, err := s.brokerClient.GetAllTrades(params)
		if err != nil {
			return testSubscriber.ID, err
		}

		if len(events) == 0 {
			break
		}

		log.Println(len(events), "got events from", testSubscriber.ID, " offset ", params.Offset)

		// Отправляем все данные в подписчика ====>
		for _, event := range events {
			if err := testSubscriber.HandleHistoryAlltrades(event); err != nil {
				if !errors.Is(err, alor.ErrNewBarFound) {
					return testSubscriber.ID, err
				}
			}
		}

		time.Sleep(100 * time.Millisecond)

		params.Offset += 10000
	}
	log.Println("get data for ", testSubscriber.ID, " finished")

	// Блокируем подписчика на прием событий, чтобы мы могли очистить очередь
	testSubscriber.SetWait()
	log.Printf("subscriber %s is blocked", testSubscriber.ID)

	// После всех манипуляций делаем подписчика активным
	testSubscriber.SetReady()
	log.Printf("subscriber %s ready to work", testSubscriber.ID)

	// Отправляем все данные из очереди в подписчика ====>
	for {
		event, err := testSubscriber.Queue.Dequeue()
		if err != nil {
			if errors.Is(err, alor.ErrQueueUnderFlow) {
				log.Printf("subscriber %s queue underflow %s", testSubscriber.ID, err.Error())
			} else {
				log.Printf("subscriber %s deque error %s", testSubscriber.ID, err.Error())
			}

			break
		}

		log.Printf("subscriber %s queue length %d", testSubscriber.ID, testSubscriber.Queue.Len)

		if err := testSubscriber.HandleHistoryEvent(event); err != nil {
			return testSubscriber.ID, err
		}
	}
	log.Printf("subscribers %s queue is empty", testSubscriber.ID)

	// Разблокируем подписчика
	testSubscriber.ReleaseWait()
	log.Printf("subscriber %s successfully added", testSubscriber.ID)

	return testSubscriber.ID, err
}
