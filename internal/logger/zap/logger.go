package zaplogger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	SugaredLogger struct {
		level  zapcore.Level
		logger *zap.SugaredLogger
	}

	Logger struct {
		level  zapcore.Level
		logger *zap.Logger
	}
)

func getZapLogLevel(level int64) zapcore.Level {
	switch level {
	case -1:
		return zap.DebugLevel
	case 0:
		return zap.InfoLevel
	case 1:
		return zap.WarnLevel
	case 2:
		return zap.ErrorLevel
	case 3:
		return zap.DPanicLevel
	case 4:
		return zap.PanicLevel
	case 5:
		return zap.FatalLevel
	default:
		return zap.DebugLevel
	}
}

func New(config Config) (*zap.Logger, error) {
	loggerLvl := getZapLogLevel(config.Level)

	var (
		loggerDev     = false
		loggerEnc     = "json"
		LoggerEncConf = zap.NewProductionEncoderConfig()
	)

	// Changing configs to debug if loggerLvl == zap.DebugLevel
	if config.Level < 0 {
		loggerDev = true
		loggerEnc = "console"
		LoggerEncConf = zap.NewDevelopmentEncoderConfig()
	}

	loggerConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(loggerLvl),
		Development:      loggerDev,
		Encoding:         loggerEnc,
		EncoderConfig:    LoggerEncConf,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := loggerConfig.Build(zap.AddCallerSkip(0))
	if err != nil {
		return nil, err
	}

	return logger, nil
}

func NewSugared(config Config) (*zap.SugaredLogger, error) {
	logger, err := New(config)
	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}

// sugar.Infow("failed to fetch URL",
// 	// Structured context as loosely typed key-value pairs.
// 	"url", url,
// 	"attempt", 3,
// 	"backoff", time.Second,
// )
// sugar.Infof("Failed to fetch URL: %s", url)
