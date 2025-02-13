package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
	"log/slog"
	"os"
	"time"
)

func SetupZapLogger(debugMode bool) *zap.Logger {
	var zapL *zap.Logger

	if debugMode {
		zapL = createDevelopmentLogger()
	} else {
		zapL = createProductionLogger()
	}

	return zapL
}

func SetSlogDefaultFromZap(zapL *zap.Logger) {
	slogL := slog.New(zapslog.NewHandler(zapL.Core(), zapslog.WithCaller(true)))
	slog.SetDefault(slogL)
}

func createProductionLogger() *zap.Logger {
	config := zapcore.EncoderConfig{
		MessageKey:    "msg",
		LevelKey:      "level",
		EncodeLevel:   zapcore.CapitalLevelEncoder,
		TimeKey:       "datetime",
		EncodeTime:    zapcore.ISO8601TimeEncoder,
		CallerKey:     "caller",
		EncodeCaller:  zapcore.ShortCallerEncoder,
		StacktraceKey: "stack_trace",
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(config),
		zapcore.Lock(os.Stdout),
		zap.InfoLevel,
	)

	zapInstance := zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(zap.Field{
			Key:     "timestamp",
			Type:    zapcore.Int64Type,
			Integer: time.Now().UnixMilli(),
		}),
	)

	//config := zap.Config{
	//	Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
	//	Development:       false,
	//	DisableCaller:     false,
	//	DisableStacktrace: false,
	//	Sampling:          nil,
	//	Encoding:          "json",
	//	EncoderConfig:     zapcoreCfg,
	//	OutputPaths: []string{
	//		"stderr",
	//	},
	//	ErrorOutputPaths: []string{
	//		"stderr",
	//	},
	//	InitialFields: map[string]interface{}{
	//		"ts": time.Now().UnixMilli(),
	//	},
	//}

	// return zap.Must(config.Build())

	return zapInstance
}

func createDevelopmentLogger() *zap.Logger {
	config := zapcore.EncoderConfig{
		MessageKey:    "msg",
		LevelKey:      "level",
		EncodeLevel:   zapcore.CapitalLevelEncoder,
		TimeKey:       "datetime",
		EncodeTime:    zapcore.ISO8601TimeEncoder,
		CallerKey:     "caller",
		EncodeCaller:  zapcore.ShortCallerEncoder,
		StacktraceKey: "stack_trace",
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(config),
		zapcore.Lock(os.Stdout),
		zapcore.DebugLevel,
	)

	zapInstance := zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(zap.Field{
			Key:     "timestamp",
			Type:    zapcore.Int64Type,
			Integer: time.Now().UnixMilli(),
		}),
	)

	return zapInstance
}

func getZapLogLevel(level int8) zapcore.Level {
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
