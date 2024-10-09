package main

import (
	"flag"
	"log"

	appConfig "github.com/MarlyasDad/rd-hub-go/internal/config"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

var envVars = appConfig.EnvVars{}

func init() {
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

}
