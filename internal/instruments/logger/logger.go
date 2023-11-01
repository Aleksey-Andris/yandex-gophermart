package logger

import (
	"strings"

	"go.uber.org/zap"
)

type Logger struct {
	*zap.SugaredLogger
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
	return &Logger{logger.Sugar()}, err
}
