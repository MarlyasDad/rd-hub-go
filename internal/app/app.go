package app

import (
	"context"
	"github.com/MarlyasDad/rd-hub-go/internal/app/http"
	appconfig "github.com/MarlyasDad/rd-hub-go/internal/config"
	tgBot "github.com/MarlyasDad/rd-hub-go/internal/infra/telegram"
	"github.com/MarlyasDad/rd-hub-go/pkg/alor"
	"github.com/MarlyasDad/rd-hub-go/pkg/logger"
	"github.com/MarlyasDad/rd-hub-go/pkg/scheduler"
	"go.uber.org/zap"
	"log"
	"log/slog"
	"os"
	"sync"
)

type (
	App struct {
		ctx          context.Context
		wg           *sync.WaitGroup
		config       appconfig.Config
		scheduler    *scheduler.Scheduler
		brokerClient *alor.Client
		tgBot        tgBot.TgClient
		httpServer   http.Server
		zapLogger    *zap.Logger
	}
)

func NewApp(ctx context.Context, wg *sync.WaitGroup, config appconfig.Config) (*App, error) {
	slog.Info("Hi! I'm Right Decisions Hub! I'm starting...")

	// Init a zap logger and add it as slog default logger
	zapLog := logger.SetupZapLogger(config.Logger.DebugMode)
	logger.SetSlogDefaultFromZap(zapLog)
	slog.Info("Logger setup successful")

	//slog.Warn("APP started")
	//slog.Info("This is an info message")

	//err := errors.New("failure")
	//slog.Error("slog", slog.Any("error", err), slog.Int("pid", os.Getpid()))

	// create a scheduler
	sch, err := scheduler.NewScheduler()
	if err != nil {
		log.Fatal(err)
	}
	slog.Info("Scheduler setup successful", slog.Any("conn", sch))

	// create a broker connection
	alorClient := alor.New(config.Broker)
	// brokerClient := broker.New(alorClient)
	slog.Info("Broker setup successful")

	// Telegram bot
	//bot, err := tgBot.New(ctx, config.Telegram)
	//if err != nil {
	//	log.Println(err)
	//	os.Exit(1)
	//}

	// bot.AddHandler("start", tgBot.NewStartAdapter(startService.New(ctx, repo)))

	// Http server
	httpServer := http.New(config.Server)
	http.RegisterHandlers(httpServer.Mux, alorClient)

	// Merge all components into app
	return &App{
		ctx:          ctx,
		wg:           wg,
		config:       config,
		scheduler:    sch,
		brokerClient: alorClient,
		httpServer:   httpServer,
		zapLogger:    zapLog,
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
	// Начинаем выполнять задания по расписанию
	a.scheduler.Start()
	// Начинаем принимать команды от telegram
	// err := a.tgBot.Start(a.ctx)
	//if err != nil {
	//	return err
	//}

	// тестовый подписчик
	//testHandler := barsToFileCommand.New("UWGN.txt")
	//testSubscriber := alor.NewSubscriber(
	//	"Test UWGN subscriber, timeframe M5, sync, HeavyBarsDetailing",
	//	alor.MOEXExchange,
	//	"UWGN",
	//	"TQBR",
	//	alor.M5TF,
	//	alor.WithDeltaData(),
	//	alor.WithMarketProfileData(),
	//	alor.WithOrderFlowData(),
	//	alor.WithAllTradesSubscription(0, false),
	//	alor.WithOrderBookSubscription(10),
	//	alor.WithCustomHandler(testHandler),
	//)
	//
	//err = a.brokerClient.AddSubscriber(testSubscriber)
	//// result, err := a.brokerClient.TestSubscriber(testSubscriber)
	//if err != nil {
	//	return err
	//}

	// Принудительно заменить commands и notifier в customHandlers на заглушки
	// Получить историю через АПИ
	// WithAllTradesHistory(from, to) сначала получить за прошлые сессии, потом за текущую, всё прогнать через subscriber
	// WithBarsHistory(from, to) ---//---
	// Когда вся история скормлена выдать результат

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
	// a.tgBot.Stop()
	// Прекращаем получать команды от http
	a.httpServer.Stop()
	// Отключаемся от брокера
	a.brokerClient.Stop(true)

	slog.Info("The RD-Hub completed correctly")

	_ = a.zapLogger.Sync()

	return nil
}
