package app

import (
	"context"
	"github.com/MarlyasDad/rd-hub-go/internal/clients/alor"
	appconfig "github.com/MarlyasDad/rd-hub-go/internal/config"
	httptransport "github.com/MarlyasDad/rd-hub-go/internal/transport/http"
	telegramtransport "github.com/MarlyasDad/rd-hub-go/internal/transport/telegram"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"net/http"
	"sync"
)

type (
	// TgBot interface {
	// 	Start() error
	// 	Stop()
	// }

	App struct {
		ctx         context.Context
		wg          *sync.WaitGroup
		config      appconfig.Config
		broker      *alor.AlorConnector
		tracer      trace.Tracer
		logger      *zap.SugaredLogger
		telegramBot telegramtransport.TgClient
		httpServer  httptransport.Server
		// repository    repository
		// hub           wsHub
		// websocketServer wsServer
	}
)

func NewApp(ctx context.Context, wg *sync.WaitGroup, config appconfig.Config, sugar *zap.SugaredLogger, tracer trace.Tracer) (*App, error) {
	// Init http server
	httpServer := httptransport.New(config.Server)
	httpServer.Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/dist/index.html")
	})

	httpFileServer := http.FileServer(http.Dir("./web/dist/assets"))
	httpServer.Mux.Handle("/assets/", http.StripPrefix("/assets/", httpFileServer))

	// Init telegram bot
	bot, err := telegramtransport.New(ctx, config.Telegram, sugar)
	if err != nil {
		sugar.Fatalf("init telegram bot: %v", err)
	}

	// Init broker connection
	brokerConn := alor.New(config.Broker)

	// Merge all components into app
	return &App{
		ctx:         ctx,
		wg:          wg,
		config:      config,
		broker:      brokerConn,
		tracer:      tracer,
		logger:      sugar,
		telegramBot: bot,
		httpServer:  httpServer,
	}, nil
}

func (a *App) Start() error {
	a.logger.Info("HI! I'm Right Decisions Hub!")

	a.httpServer.Start()

	err := a.telegramBot.Start()
	if err != nil {
		return err
	}

	// Подключаемся к брокеру
	// a.broker.Start(a.ctx)

	a.logger.Info("RD-Hub: started...")

	return nil
}

func (a *App) Stop() error {
	defer a.logger.Sync()

	a.httpServer.Stop()
	a.telegramBot.Stop()
	// a.broker.Stop()

	a.logger.Info("RD-Hub: shutdown...")

	return nil
}
