package main

import (
	"context"
	"flag"
	"github.com/MarlyasDad/rd-hub-go/internal/app"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"log"
	"log/slog"
	"os"
	"sync"

	appConfig "github.com/MarlyasDad/rd-hub-go/internal/config"
)

func main() {
	envVars := loadEnv()
	conf := appConfig.NewConfig(envVars)

	wg := sync.WaitGroup{}
	ctx := RunSignalHandler(context.Background(), &wg)

	application, err := app.NewApp(ctx, &wg, conf)
	if err != nil {
		// slog.Error("Failed to create app", "err", err)
		log.Println("Failed to create app", "err", err)
		os.Exit(1)
	}

	err = application.Start()
	if err != nil {
		slog.Error("Failed to run server", "err", err)
		os.Exit(1)
	}

	wg.Wait()

	err = application.Stop()
	if err != nil {
		slog.Error("Failed to stop server", "err", err)
		os.Exit(1)
	}
}

func loadEnv() appConfig.EnvVars {
	var envVars = appConfig.EnvVars{}

	var configPath string
	flag.StringVar(&configPath, "config", ".env", "path to .env file")

	err := godotenv.Load(configPath)
	if err != nil {
		log.Println("local .env file not found")
	}

	err = envconfig.Process("rd", &envVars)
	if err != nil {
		log.Fatal(err.Error())
	}

	return envVars
}
