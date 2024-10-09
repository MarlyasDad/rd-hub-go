package app

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/MarlyasDad/rd-hub-go/internal/clients/alor"
	"github.com/MarlyasDad/rd-hub-go/internal/clients/telegram"
	app_config "github.com/MarlyasDad/rd-hub-go/internal/config"
	"github.com/MarlyasDad/rd-hub-go/internal/tracer/jaeger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type (
	// TgBot interface {
	// 	Start() error
	// 	Stop()
	// }

	App struct {
		ctx           context.Context
		wg            *sync.WaitGroup
		config        app_config.Config
		broker        *alor.AlorConnector
		traceProvider *sdktrace.TracerProvider
		tracer        trace.Tracer
		logger        *zap.SugaredLogger
		bot           telegram.TgClient
		// repository    repository
		// mux           mux
		// server        httpServer
		// hub           wsHub
	}
)

func NewApp(config app_config.Config) (*App, error) {
	ctx := context.Background()
	wg := sync.WaitGroup{}

	loggerConfig := zap.NewProductionConfig()
	loggerLvl := zap.NewAtomicLevelAt(zap.DebugLevel)

	loggerConfig.OutputPaths = []string{"stdout"}
	loggerConfig.Level = loggerLvl

	logger, err := loggerConfig.Build()
	if err != nil {
		panic(err)
	}

	logger.WithOptions()

	sugar := logger.Sugar()
	// sugar.Infow("failed to fetch URL",
	// 	// Structured context as loosely typed key-value pairs.
	// 	"url", url,
	// 	"attempt", 3,
	// 	"backoff", time.Second,
	// )
	// sugar.Infof("Failed to fetch URL: %s", url)

	// Init OTEL
	traceProvider, err := jaeger.InitTraceProvider("localhost:4317", "RD-Hub Trading Platform", 1.0)
	if err != nil {
		sugar.Fatal("init trace provider", err)
	}

	tracer := traceProvider.Tracer("rd-hub-go tracer")

	bot, err := telegram.New(config.Telegram, &wg, sugar)
	if err != nil {
		sugar.Fatalf("init telegram bot: ", err)
	}

	// // Get bot user
	// botUser, err := bot.GetMe()
	// if err != nil {
	// 	sugar.Fatal("get me error", err)
	// }

	// sugar.Infof("Bot user: %+v", botUser)

	// Init Broker conn
	brokerConn := alor.New(config.Broker)

	// Merge components into app
	return &App{
		ctx:           ctx,
		wg:            &wg,
		config:        config,
		broker:        brokerConn,
		traceProvider: traceProvider,
		tracer:        tracer,
		logger:        sugar,
		bot:           bot,
	}, nil
}

func (a *App) Run() error {
	a.logger.Info("HI! I'm Right Decisions Hub!")
	// Завершаем приложение gracefully
	defer a.Shutdown()

	// Запускаем обработчик сигналов
	sigCtx := a.runSignalHandler()

	// Подключаемся к брокеру
	// a.broker.Start(sigCtx)

	a.bot.Start(sigCtx)

	a.logger.Info("RD-Hub: started...")

	// Ждём завершения всех воркеров
	a.wg.Wait()

	return nil
}

func (a *App) Shutdown() error {
	defer a.logger.Sync()

	a.bot.Stop()

	// a.broker.Stop()
	a.traceProvider.Shutdown(a.ctx)

	a.logger.Info("RD-Hub: shutdown...")

	return nil
}

func (a *App) runSignalHandler() context.Context {
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	sigCtx, cancel := context.WithCancel(a.ctx)

	a.wg.Add(1)
	go func() {
		defer a.logger.Info("RD-Hub: [signal] terminate")
		defer signal.Stop(sigterm)
		defer a.wg.Done()
		defer cancel()

		for {
			select {
			case sig, ok := <-sigterm:
				if !ok {
					a.logger.Infof("RD-Hub: [signal] signal chan closed: %s\n", sig.String())
					return
				}

				a.logger.Infof("RD-Hub: [signal] signal recv: %s", sig.String())
				return
			case _, ok := <-sigCtx.Done():
				if !ok {
					a.logger.Info("RD-Hub: [signal] context closed")
					return
				}

				a.logger.Infof("RD-Hub: [signal] ctx done: %n", a.ctx.Err().Error())
				return
			}
		}
	}()

	return sigCtx
}
