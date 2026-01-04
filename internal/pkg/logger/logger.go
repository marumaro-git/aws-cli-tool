package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/morikuni/failure"
)

type Logger interface {
	Info(ctx context.Context, msg string)
	Warn(ctx context.Context, msg string)
	Error(ctx context.Context, err error)
}

type SlogLogger struct {
	logger *slog.Logger
}

func NewSlogLogger() *SlogLogger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	return &SlogLogger{
		logger: logger,
	}
}

func (l *SlogLogger) Info(ctx context.Context, msg string) {
	l.logger.InfoContext(ctx, msg)
}

func (l *SlogLogger) Warn(ctx context.Context, msg string) {
	l.logger.WarnContext(ctx, msg)
}

func (l *SlogLogger) Error(ctx context.Context, err error) {

	code, ok := failure.CodeOf(err)
	if !ok {
		code = failure.StringCode("Unknown")
	}

	msg, ok := failure.MessageOf(err)
	if !ok {
		msg = err.Error()
	}

	trace, ok := failure.CallStackOf(err)
	if !ok {
		trace = nil
	}

	errGroup := slog.Group(
		"error",
		slog.String("code", code.ErrorCode()),
		slog.String("message", msg),
		slog.String("stacktrace", fmt.Sprintf("%+v", trace)),
	)

	l.logger.ErrorContext(ctx, "An error occurred", errGroup)

}
