package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Logger struct {
	Level zerolog.Level
}

type LoggerInterface interface {
	Info(msg string, tags map[string]interface{})
	Warn(msg string, tags map[string]interface{})
	Error(msg string, err error, tags map[string]interface{})
	Debug(msg string, tags map[string]interface{})
	Trace(msg string, tags map[string]interface{})
}

func NewLogger(level string) LoggerInterface {
	setup(level)

	return &Logger{
		Level: getLevel(level),
	}
}

func setup(level string) {
	zerolog.SetGlobalLevel(getLevel(level))

	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	})
}

func (l *Logger) Info(msg string, tags map[string]interface{}) {
	l.getLogger().Info().Fields(tags).Msg(msg)
}

func (l *Logger) Warn(msg string, tags map[string]interface{}) {
	l.getLogger().Warn().Fields(tags).Msg(msg)
}

func (l *Logger) Error(msg string, err error, tags map[string]interface{}) {
	l.getLogger().Error().Fields(tags).Err(err).Msg(msg)
}

func (l *Logger) Debug(msg string, tags map[string]interface{}) {
	l.getLogger().Debug().Fields(tags).Msg(msg)
}

func (l *Logger) Trace(msg string, tags map[string]interface{}) {
	l.getLogger().Trace().Fields(tags).Msg(msg)
}

func (l *Logger) getLogger() *zerolog.Logger {
	return &log.Logger
}

func getLevel(level string) zerolog.Level {
	switch level {
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "debug":
		return zerolog.DebugLevel
	case "trace":
		return zerolog.TraceLevel
	default:
		return zerolog.InfoLevel
	}
}
