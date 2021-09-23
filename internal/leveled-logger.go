package internal

import (
	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog"
)

type LeveledLogger struct {
	retryablehttp.LeveledLogger
	logger *zerolog.Logger
}

func NewLeveledLogger(logger *zerolog.Logger) *LeveledLogger {
	return &LeveledLogger{logger: logger}
}

func (e *LeveledLogger) Error(msg string, keysAndValues ...interface{}) {
	e.logger.Error().Fields(keysAndValues).Msg(msg)
}
func (e *LeveledLogger) Info(msg string, keysAndValues ...interface{}) {
	e.logger.Info().Fields(keysAndValues).Msg(msg)
}
func (e *LeveledLogger) Debug(msg string, keysAndValues ...interface{}) {
	e.logger.Debug().Fields(keysAndValues).Msg(msg)
}
func (e *LeveledLogger) Warn(msg string, keysAndValues ...interface{}) {
	e.logger.Warn().Fields(keysAndValues).Msg(msg)
}
