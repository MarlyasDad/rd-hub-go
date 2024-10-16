package main

import (
	"context"
	zaplogger "github.com/MarlyasDad/rd-hub-go/internal/logger/zap"
	jaegertracer "github.com/MarlyasDad/rd-hub-go/internal/tracer/jaeger"
	"log"
	"sync"

	app "github.com/MarlyasDad/rd-hub-go/internal/app/rd_hub"
	appConfig "github.com/MarlyasDad/rd-hub-go/internal/config"
)

func main() {
	conf := appConfig.NewConfig(envVars)

	// Создаём логгер
	sugar, err := zaplogger.NewSugared(conf.Logger)
	if err != nil {
		log.Fatalf("init logger error: %s", err)
	}
	defer func() {
		_ = sugar.Sync()
	}()

	// Запускаем обработчик сигналов
	wg := sync.WaitGroup{}
	ctx := app.RunSignalHandler(context.Background(), &wg, sugar)

	// Создаём tracer
	traceProvider, err := jaegertracer.InitTraceProvider(ctx, conf.Tracer, "RD-Hub Trading Platform")
	if err != nil {
		sugar.Fatal("init trace provider", err)
	}
	defer func() {
		_ = traceProvider.Shutdown(ctx)
	}()

	tracer := traceProvider.Tracer("rd-hub-go tracer")

	// Создаём новое приложение
	service, err := app.NewApp(ctx, &wg, conf, sugar, tracer)
	if err != nil {
		log.Fatal("{FATAL} ", err)
	}

	// Запускаем приложение
	err = service.Start()
	if err != nil {
		sugar.Fatal("{FATAL} ", err)
	}

	wg.Wait()

	err = service.Stop()
	if err != nil {
		sugar.Fatal("{FATAL} ", err)
	}
}
