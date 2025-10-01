package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"quasarflow-api/pkg/errors"
)

type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
	WithContext(ctx context.Context) Logger
}

type Field = zapcore.Field

type zapLogger struct {
	logger *zap.Logger
}

func New(level string) Logger {
	config := zap.NewProductionConfig()

	// Configurar o nível de log baseado no parâmetro
	switch level {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	// Customizar o formato de saída
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	logger, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		os.Exit(1)
	}

	return &zapLogger{logger: logger}
}

func (l *zapLogger) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, fields...)
}

func (l *zapLogger) Info(msg string, fields ...Field) {
	l.logger.Info(msg, fields...)
}

func (l *zapLogger) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, fields...)
}

func (l *zapLogger) Error(msg string, fields ...Field) {
	l.logger.Error(msg, fields...)
}

func (l *zapLogger) Fatal(msg string, fields ...Field) {
	l.logger.Fatal(msg, fields...)
}

func (l *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{logger: l.logger.With(fields...)}
}

func (l *zapLogger) WithContext(ctx context.Context) Logger {
	// Adicionar campos do contexto ao logger
	return l
}

// Funções auxiliares para criar campos
func String(key, value string) Field {
	return zap.String(key, value)
}

func Int(key string, value int) Field {
	return zap.Int(key, value)
}

func Error(err error) Field {
	if appErr, ok := err.(*errors.AppError); ok {
		return zap.Object("error", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
			enc.AddString("type", string(appErr.Type))
			enc.AddString("message", appErr.Message)
			if appErr.Detail != "" {
				enc.AddString("detail", appErr.Detail)
			}
			if appErr.Err != nil {
				enc.AddString("cause", appErr.Err.Error())
			}
			return nil
		}))
	}
	return zap.Error(err)
}

func Any(key string, value interface{}) Field {
	return zap.Any(key, value)
}
