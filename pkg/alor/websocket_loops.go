package alor

import (
	"context"
	"errors"
	"log"
	"time"
)

func (ws *Websocket) runWebsocketLoop(ctx context.Context) {
	log.Println("websocket loop is running")
	defer close(ws.done)

	defer log.Println("websocket loop closed")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, msg, err := ws.conn.ReadMessage()
			if err != nil {
				log.Println("Error in receive:", err)
				return
			}

			// Обрабатываем входящее сообщение
			err = ws.HandleResponse(msg)
			if err != nil {
				log.Println("Error in handle:", err)
				return
			}
		}
	}
}

func (ws *Websocket) runQueueLoop(ctx context.Context) {
	// Включать при включении сокета
	// Не выключать при отключении сокета
	log.Println("websocket queue loop is running")
	defer log.Println("websocket queue loop closed")
	for {
		select {
		case <-ctx.Done():
			return
		case <-ws.done:
			return
		default:
			// do work
			event, err := ws.queue.Dequeue()
			if err != nil {
				if errors.Is(err, ErrQueueUnderFlow) {
					time.Sleep(time.Millisecond * 500)
					continue
				}

				log.Println("Error in receive:", err)
				return
			}

			log.Println("length dequeue", ws.queue.GetLength())

			// Устанавливаем подписку как активную если по ней пришло событие
			if !ws.subscriptions[event.Guid].Active {
				item := ws.subscriptions[event.Guid]
				item.Active = true
				ws.subscriptions[event.Guid] = item
			}

			// Последовательное выполнение может занимать много времени - тогда заменить на асинхронные обработчики
			for _, subscriber := range ws.subscriptions[event.Guid].Items {
				// Блокируем добавление/удаление любых подписчиков пока не пройдёт handle
				// Никто не может поменять subscriptions во время вычисления
				// Добавление или удаление из-за этого может занять продолжительное время

				if subscriber == nil || subscriber.Done {
					// Разблокируем удаление подписчиков
					ws.mu.Unlock()
					continue
				}

				// TODO: Выполнять все вместе параллельно или каждый с асинхронным обработчиком?
				// Синхронное выполнение с задержкой хотя-бы одного воркера может тормозить остальные
				// Отличная идея - выполнять сабскриберы в горутинах. Тогда отставать будет самый нагруженный
				// А легковесные будут пролетать со свистом

				// !Проблема синхронизации между воркерами - один имеет актуальное состояние, а другой нет
				// Как решить - непонятно. Если только сравнивать длину очередей. Если с маленькой погрешностью не отличаются
				// Идея! Если нужно сделать зависимые сабскриберы - делать для них асинхронную оболочку и внутри обрабатывать события синхронно
				// subscribersGroup - интерфейс как у сабскрибера

				// !Проблема отставания от текущей ситуации (решается распараллеливанием) - проверять очередь на количество необработанных элементов
				// Всё, что работает в реалтайме не должно превышать определённый порог загруженности очереди
				// Причём, как общей очереди вебсокета, так и в частной очереди сабскрибера
				// Если очередь переполняется и не разгружается, то отключаем сабскрибера SetDone()
				// Так мы отсекаем самых медленных подписчиков
				// TODO: Нужно сделать проверку очереди вебсокета. При достижении 50тысяч, отключать вебсокет и алертить!
				// Или отключать всех сабскриберов
				// Если очередь больше дельты, не отправлять команды брокеру.
				// Работают только самые шустрые, медленные отключаются

				// В каждом сабскрибере делать свой контекст с отменой от родительского. Отменять горутину когда сабскрибер будет удаляться.

				ws.mu.Lock()

				if err := subscriber.HandleEvent(event); err != nil {
					subscriber.SetDone()
					log.Println(subscriber.ID, "Error in handle:", err)
				}

				// Разблокируем удаление подписчиков
				ws.mu.Unlock()
			}
		}
	}
}

func (ws *Websocket) StartHealthLoop(ctx context.Context, token Token) {
	ws.healthDone = make(chan interface{})

	go func() {
		log.Println("websocket health loop is running")
		defer log.Println("websocket health loop closed")

		workGen := time.NewTicker(time.Second * 1)
		defer workGen.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ws.healthDone:
				return
			case <-workGen.C:
				if !ws.IsConnected() {
					log.Println("try to websocket connect")
					token, err := token.GetAccessToken()
					if err != nil {
						return
					}

					if err := ws.Connect(ctx, token); err != nil {
						log.Println(err)
					}
				}

				// делаем паузу после попытки подключения
				time.Sleep(time.Second * 10)
			}
		}
	}()
}

func (ws *Websocket) StopHealthLoop() {
	close(ws.healthDone)
}
