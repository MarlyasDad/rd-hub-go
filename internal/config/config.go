package config

import (
	"github.com/MarlyasDad/rd-hub-go/internal/infra/jaeger"
	"github.com/MarlyasDad/rd-hub-go/internal/transport/http"
	"github.com/MarlyasDad/rd-hub-go/pkg/logger"
	"time"

	"github.com/MarlyasDad/rd-hub-go/internal/infra/telegram"
	"github.com/MarlyasDad/rd-hub-go/pkg/alor"
)

type (
	EnvVars struct {
		ApiHost               string    `envconfig:"rd_server_host"`
		ApiPort               int64     `envconfig:"rd_server_port"`
		BrokerRefreshToken    string    `envconfig:"broker_refresh"`
		BrokerRefreshTokenExp time.Time `envconfig:"broker_refresh_exp"`
		BrokerDevCircuit      bool      `envconfig:"broker_dev_circuit" default:"true"`
		OtelGrpcEndpoint      string    `envconfig:"otel_grpc_endpoint"`
		OtelRatioBased        float64   `envconfig:"otel_ratio_based" default:"0.0"`
		DebugMode             bool      `envconfig:"debug_mode" default:"false"`
		TelegramBotToken      string    `envconfig:"telegram_bot_token"`
	}

	Config struct {
		Server   http.Config
		Broker   alor.Config
		Tracer   jaeger.Config
		Logger   logger.Config
		Telegram telegram.Config
		// repository db_repo.Config
	}
)

func NewConfig(f EnvVars) Config {
	return Config{
		Server: http.Config{
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
			DebugMode: f.DebugMode,
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
