package app

import (
	"context"
	"fmt"
	appconfig "github.com/MarlyasDad/rd-hub-go/internal/config"
	"github.com/go-co-op/gocron/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"log"

	// jaegertracer "github.com/MarlyasDad/rd-hub-go/internal/infra/jaeger"
	httpTransport "github.com/MarlyasDad/rd-hub-go/internal/transport/http"
	tgTransport "github.com/MarlyasDad/rd-hub-go/internal/transport/telegram"
	"github.com/MarlyasDad/rd-hub-go/pkg/alor"
	// ops_counter "github.com/MarlyasDad/rd-hub-go/pkg/ops-counter"
	// "go.uber.org/zap"
	"log/slog"
	"net/http"
	"sync"
)

type (
	App struct {
		ctx        context.Context
		wg         *sync.WaitGroup
		config     appconfig.Config
		scheduler  gocron.Scheduler
		broker     *alor.Client
		tgBot      tgTransport.TgClient
		httpServer httpTransport.Server
		zapLogger  *zap.Logger
		// traceProvider *jaegertracer.TracerProvider
		// repository    repository
		// hub           wsHub
		// websocketServer wsServer
	}
)

func NewApp(ctx context.Context, wg *sync.WaitGroup, config appconfig.Config) (*App, error) {
	zapL := zap.Must(zap.NewProduction())
	logger := slog.New(zapslog.NewHandler(zapL.Core(), nil))
	slog.SetDefault(logger)

	//traceProvider, err := jaegertracer.InitTraceProvider(ctx, config.Tracer, "RD-Hub Trading Platform")
	//if err != nil {
	//	log.Fatal("init trace provider error", err)
	//}

	// tracer := traceProvider.Tracer("rd-hub-go tracer")

	// create a scheduler
	s, err := gocron.NewScheduler()
	if err != nil {
		// handle error
	}

	// Создать задание присылать каждые 9 утра состояние счёта
	fmt.Println(s)

	// create a broker connection
	brokerConn := alor.New(config.Broker)

	// Telegram bot
	//bot, err := telegramTransport.New(ctx, config.Telegram)
	//if err != nil {
	//	slog.Error("init telegram bot: %v", err)
	//	os.Exit(1)
	//}

	// Http server
	httpServer := httpTransport.New(config.Server)
	httpServer.Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/dist/index.html")
	})
	// Static files
	httpFileServer := http.FileServer(http.Dir("./web/dist/assets"))
	httpServer.Mux.Handle("GET /assets/", http.StripPrefix("/assets/", httpFileServer))

	// Statistic
	// опрашивает все системы и выдаёт общую статистику в виде json
	httpServer.Mux.HandleFunc("GET /statistics/", func(w http.ResponseWriter, r *http.Request) {})

	// Portfolios
	// httpServer.Mux.HandleFunc("GET /portfolios/", httpTransport.NewPortfoliosListHandler())
	httpServer.Mux.HandleFunc("GET /portfolios/{portfolio}/", func(w http.ResponseWriter, r *http.Request) {})

	// httpServer.Mux.Handle("GET /counters/", appHttp.NewMetricsHandler())
	// Depth of market
	httpServer.Mux.HandleFunc("GET /order-book/{security}/", func(w http.ResponseWriter, r *http.Request) {})
	// websocket management
	httpServer.Mux.HandleFunc("POST /order-book/streaming/subscribe/", func(w http.ResponseWriter, r *http.Request) {})
	httpServer.Mux.HandleFunc("POST /order-book/streaming/unsubscribe/", func(w http.ResponseWriter, r *http.Request) {})

	// Websocket server handler if exists

	// Merge all components into app
	return &App{
		ctx:        ctx,
		wg:         wg,
		config:     config,
		scheduler:  s,
		broker:     brokerConn,
		httpServer: httpServer,
		zapLogger:  zapL,
		// traceProvider: traceProvider,
		// telegramBot:   bot,
	}, nil
}

func (a *App) Start() error {
	slog.Info("Hi! I'm Right Decisions Hub! I'm starting...")

	// Подключаемся к брокеру
	a.broker.Start(a.ctx)
	// Начинаем принимать команды от http
	a.httpServer.Start()
	// Начинаем принимать команды от telegram
	err := a.tgBot.Start()
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
