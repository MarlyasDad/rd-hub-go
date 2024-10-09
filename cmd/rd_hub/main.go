package main

import (
	"log"

	app "github.com/MarlyasDad/rd-hub-go/internal/app/rd_hub"
	appConfig "github.com/MarlyasDad/rd-hub-go/internal/config"
)

func main() {
	var conf = appConfig.NewConfig(envVars)

	// Создаём новое приложение
	service, err := app.NewApp(conf)
	if err != nil {
		log.Fatal("{FATAL} ", err)
	}

	// Запускаем приложение
	err = service.Run()
	if err != nil {
		log.Fatal("{FATAL} ", err)
	}
}
