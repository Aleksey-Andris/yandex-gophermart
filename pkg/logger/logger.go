package logger

import (
	"strings"

	"go.uber.org/zap"
)

type Logger struct {
	logger *zap.SugaredLogger
}

func New(environment string) (*Logger, error) {
	var logger *zap.Logger
	var err error
	switch strings.ToLower(environment) {
	case "prod":
		logger, err = zap.NewProduction()
	case "develop":
		logger, err = zap.NewDevelopment()
	default:
		logger, err = zap.NewDevelopment()
	}
	return &Logger{
		logger: logger.Sugar(),
	}, err
}

func (l *Logger) Debug(message string, args ...interface{}) {
	l.logger.Debug(message, args)
}
func (l *Logger) Info(message string, args ...interface{}) {
	l.logger.Info(message, args)
}
func (l *Logger) Warn(message string, args ...interface{}) {
	l.logger.Warn(message, args)
}
func (l *Logger) Error(message string, args ...interface{}) {
	l.logger.Error(message, args)
}
func (l *Logger) Fatal(message string, args ...interface{}) {
	l.logger.Fatal(message, args)
}
func (l *Logger) Sync() {
	l.logger.Sync()
}
