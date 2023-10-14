package internal

import "context"

type Logger interface {
	Debug(message string, args ...interface{})
	Info(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message string, args ...interface{})
	Fatal(message string, args ...interface{})
}

type PingStorage interface {
	Ping(ctx context.Context) error
}

type PingUsecase interface {
	Ping(ctx context.Context) error
}
