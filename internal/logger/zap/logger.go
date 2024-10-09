package logger

import (
	"go.uber.org/zap"
)

type Logger struct {
	level  zap.AtomicLevel
	logger *zap.Logger
}

func getZapLogLevel(level int64) zap.AtomicLevel {
	switch level {
	case -1:
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case 0:
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	case 1:
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	case 2:
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	case 3:
		return zap.NewAtomicLevelAt(zap.DPanicLevel)
	case 4:
		return zap.NewAtomicLevelAt(zap.PanicLevel)
	case 5:
		return zap.NewAtomicLevelAt(zap.FatalLevel)
	default:
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	}
}

func New(config Config) Logger {
	loggerConfig := zap.NewProductionConfig()
	loggerLvl := getZapLogLevel(config.Level)

	loggerConfig.OutputPaths = []string{"stdout"}
	loggerConfig.Level = loggerLvl

	logger, err := loggerConfig.Build()
	if err != nil {
		panic(err)
	}

	logger.WithOptions()

	// sugar := logger.Sugar()
	// sugar.Infow("failed to fetch URL",
	// 	// Structured context as loosely typed key-value pairs.
	// 	"url", url,
	// 	"attempt", 3,
	// 	"backoff", time.Second,
	// )
	// sugar.Infof("Failed to fetch URL: %s", url)

	return Logger{
		level:  loggerLvl,
		logger: logger,
	}
}
