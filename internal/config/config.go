package config

import (
	"time"

	"github.com/MarlyasDad/rd-hub-go/internal/clients/alor"
	"github.com/MarlyasDad/rd-hub-go/internal/clients/telegram"
	"github.com/MarlyasDad/rd-hub-go/internal/logger"
	"github.com/MarlyasDad/rd-hub-go/internal/tracer/jaeger"
)

type (
	EnvVars struct {
		ApiHost               string    `envconfig:"rd_api_host"`
		ApiPort               int64     `envconfig:"rd_api_port"`
		BrokerRefreshToken    string    `envconfig:"broker_refresh"`
		BrokerRefreshTokenExp time.Time `envconfig:"broker_refresh_exp"`
		BrokerDevCircuit      bool      `envconfig:"broker_dev_circuit" default:"true"`
		OtelGrpcEndpoint      string    `envconfig:"otel_grpc_endpoint"`
		OtelRatioBased        float64   `envconfig:"otel_ratio_based" default:"0.0"`
		LogLevel              int64     `envconfig:"log_level" default:"-1"`
		TelegramBotToken      string    `envconfig:"telegram_bot_token"`
	}

	serverConfig struct {
		Host string
		Port int64
	}

	Config struct {
		Server   serverConfig
		Broker   alor.Config
		Tracer   jaeger.Config
		Logger   logger.Config
		Telegram telegram.Config
		// repository db_repo.Config
	}
)

func NewConfig(f EnvVars) Config {
	return Config{
		Server: serverConfig{
			Host: f.ApiHost,
			Port: f.ApiPort,
		},
		Broker: alor.Config{
			RefreshToken:    f.BrokerRefreshToken,
			RefreshTokenExp: f.BrokerRefreshTokenExp,
			DevCircuit:      f.BrokerDevCircuit,
		},
		Tracer: jaeger.Config{
			Endpoint:          f.OtelGrpcEndpoint,
			TraceIDRatioBased: f.OtelRatioBased,
		},
		Logger: logger.Config{
			Level: f.LogLevel,
		},
		Telegram: telegram.Config{
			BotToken: f.TelegramBotToken,
		},
		// repository: db_repo.Config{
		// 	Host:     f.DatabaseHost,
		// 	Port:     f.DatabasePort,
		// 	Name:     f.DatabaseName,
		// 	Username: f.DatabaseUsername,
		// 	Password: f.DatabasePassword,
		// },
	}
}
