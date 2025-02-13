package app

import (
	"context"
	appconfig "github.com/MarlyasDad/rd-hub-go/internal/config"
	tgBot "github.com/MarlyasDad/rd-hub-go/internal/infra/telegram"
	httpTransport "github.com/MarlyasDad/rd-hub-go/internal/transport/http"
	"github.com/MarlyasDad/rd-hub-go/pkg/alor"
	"github.com/MarlyasDad/rd-hub-go/pkg/logger"
	"github.com/MarlyasDad/rd-hub-go/pkg/scheduler"
	"go.uber.org/zap"
	"log"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"
)

type (
	App struct {
		ctx          context.Context
		wg           *sync.WaitGroup
		config       appconfig.Config
		scheduler    *scheduler.Scheduler
		brokerClient *alor.Client
		tgBot        tgBot.TgClient
		httpServer   httpTransport.Server
		zapLogger    *zap.Logger
	}
)

func NewApp(ctx context.Context, wg *sync.WaitGroup, config appconfig.Config) (*App, error) {
	slog.Info("Hi! I'm Right Decisions Hub! I'm starting...")

	// Init a zap logger and add it as slog default logger
	zapL := logger.SetupZapLogger(config.Logger.DebugMode)
	logger.SetSlogDefaultFromZap(zapL)
	slog.Info("Logger setup successful")

	//slog.Warn("APP started")
	//slog.Info("This is an info message")
	//
	//err := errors.New("failure")
	//slog.Error("slog", slog.Any("error", err), slog.Int("pid", os.Getpid()))

	// create a scheduler
	s, err := scheduler.NewScheduler()
	if err != nil {
		log.Fatal(err)
	}
	slog.Info("Scheduler setup successful", slog.Any("conn", s))

	// create a broker connection
	brokerClient := alor.New(config.Broker)
	slog.Info("Broker setup successful")

	// Telegram bot
	//bot, err := tgBot.New(ctx, config.Telegram)
	//if err != nil {
	//	log.Println(err)
	//	os.Exit(1)
	//}

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

	// Merge all components into app
	return &App{
		ctx:          ctx,
		wg:           wg,
		config:       config,
		scheduler:    s,
		brokerClient: brokerClient,
		httpServer:   httpServer,
		zapLogger:    zapL,
		// tgBot:        bot,
	}, nil
}

func (a *App) Start() error {
	// Подключаемся к брокеру
	err := a.brokerClient.Connect(a.ctx, true)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	// Начинаем принимать команды от http
	a.httpServer.Start()
	// Начинаем принимать команды от telegram
	// err := a.tgBot.Start(a.ctx)
	//if err != nil {
	//	return err
	//}
	// Начинаем выполнять задания по расписанию
	a.scheduler.Start()

	slog.Info("The RD-Hub started correctly")

	// notifier = bot
	// тестовый подписчик
	testSubscriber := alor.NewSubscriber(
		[]alor.Subscription{
			{
				Exchange: alor.MOEXExchange,
				Code:     "SBER",
				Board:    "TQBR",
				Tf:       alor.M1Timeframe,
				Opcode:   alor.BarsOpcode,
				Format:   alor.SlimResponseFormat,
				From:     int(time.Now().Add(time.Hour * -24).Unix()),
			},
		},
		nil,
		a.brokerClient,
	)
	// добавляем подписчика в клиент
	a.brokerClient.Websocket.AddSubscriber(testSubscriber)

	// Messages
	// Отправить сообщение всем активным клиентам

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
	// a.tgBot.Stop()
	// Прекращаем получать команды от http
	a.httpServer.Stop()
	// Отключаемся от брокера
	a.brokerClient.Stop()

	slog.Info("The RD-Hub completed correctly")

	_ = a.zapLogger.Sync()

	return nil
}
