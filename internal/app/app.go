package app

import (
	"context"
	"fmt"
	appconfig "github.com/MarlyasDad/rd-hub-go/internal/config"
	"github.com/MarlyasDad/rd-hub-go/pkg/scheduler"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"log"
	"os"

	tgBot "github.com/MarlyasDad/rd-hub-go/internal/infra/telegram"
	httpTransport "github.com/MarlyasDad/rd-hub-go/internal/transport/http"
	"github.com/MarlyasDad/rd-hub-go/pkg/alor"
	"log/slog"
	"net/http"
	"sync"
)

type (
	App struct {
		ctx        context.Context
		wg         *sync.WaitGroup
		config     appconfig.Config
		scheduler  *scheduler.Scheduler
		broker     *alor.Client
		tgBot      tgBot.TgClient
		httpServer httpTransport.Server
		zapLogger  *zap.Logger
	}
)

func NewApp(ctx context.Context, wg *sync.WaitGroup, config appconfig.Config) (*App, error) {
	zapL := zap.Must(zap.NewProduction())
	logger := slog.New(zapslog.NewHandler(zapL.Core(), nil))
	slog.SetDefault(logger)

	// create a scheduler
	s, err := scheduler.NewScheduler()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(s)

	// create a broker connection
	brokerClient := alor.New(config.Broker)

	// Telegram bot
	bot, err := tgBot.New(ctx, config.Telegram)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// bot.AddHandler("start", tgBot.NewStartAdapter(startService.New(ctx, repo)))

	// Http server
	httpServer := httpTransport.New(config.Server)
	httpServer.Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/dist/index.html")
	})
	// Static files
	httpFileServer := http.FileServer(http.Dir("./web/dist/assets"))
	httpServer.Mux.Handle("GET /assets/", http.StripPrefix("/assets/", httpFileServer))

	// Portfolios
	httpServer.Mux.HandleFunc("GET /api/v1/portfolios/", func(w http.ResponseWriter, r *http.Request) {})
	httpServer.Mux.HandleFunc("GET /api/v1/portfolios/{portfolio}/", func(w http.ResponseWriter, r *http.Request) {})

	// Depth of market
	httpServer.Mux.HandleFunc("GET /api/v1/order-book/{security}/", func(w http.ResponseWriter, r *http.Request) {})

	// Subscribers
	// Создаём подписчика и подписываем его на необходимые инструменты
	// Все инструкции передаём через тело запроса в формате JSON
	httpServer.Mux.HandleFunc("POST /api/v1/subscribers/subscribe/", func(w http.ResponseWriter, r *http.Request) {})
	// Удаляем подписчика по его ID
	httpServer.Mux.HandleFunc("DELETE /api/v1/subscribers/unsubscribe/{subscriber_id}/", func(w http.ResponseWriter, r *http.Request) {})
	// Получаем список активных подписчиков
	httpServer.Mux.HandleFunc("GET /api/v1/subscribers/", func(w http.ResponseWriter, r *http.Request) {})
	// Получаем информацию об активном подписчике
	httpServer.Mux.HandleFunc("GET /api/v1/subscribers/{subscriber_id}/", func(w http.ResponseWriter, r *http.Request) {})

	// Websocket client
	// Получаем количество необработанных сообщений в очереди
	httpServer.Mux.HandleFunc("GET /api/v1/stream/queue-status/", func(w http.ResponseWriter, r *http.Request) {})

	// Messages
	// Отправить сообщение всем активным клиентам

	// тестовый подписчик
	testSubscriber := alor.NewSyncSubscriber(alor.M1Timeframe, nil, nil) // name?
	// добавляем подписчика в клиент
	brokerClient.Websocket.AddSubscriber(testSubscriber)

	// Merge all components into app
	return &App{
		ctx:        ctx,
		wg:         wg,
		config:     config,
		scheduler:  s,
		broker:     brokerClient,
		httpServer: httpServer,
		zapLogger:  zapL,
		tgBot:      bot,
	}, nil
}

func (a *App) Start() error {
	slog.Info("Hi! I'm Right Decisions Hub! I'm starting...")

	// Подключаемся к брокеру
	a.broker.Start(a.ctx)
	// Начинаем принимать команды от http
	a.httpServer.Start()
	// Начинаем принимать команды от telegram
	err := a.tgBot.Start(a.ctx)
	if err != nil {
		return err
	}
	// Начинаем выполнять задания по расписанию
	a.scheduler.Start()

	slog.Info("The RD-Hub started correctly")

	return nil
}

func (a *App) Stop() error {
	slog.Info("The RD-Hub is shutting down...")

	// Прекращаем выполнять задания по расписанию
	err := a.scheduler.Shutdown()
	if err != nil {
		log.Fatal("Error when stopping scheduler", err)
	}
	// Прекращаем получать команды от telegram
	a.tgBot.Stop()
	// Прекращаем получать команды от http
	a.httpServer.Stop()
	// Отключаемся от брокера
	a.broker.Stop()
	// Выключаем трейсер
	// a.traceProvider.Shutdown(a.ctx)

	slog.Info("The RD-Hub completed correctly")

	_ = a.zapLogger.Sync()

	return nil
}
