package logging

import (
	"log/slog"
	"os"
)

func Init(service string) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := slog.New(handler).With(
		"service", service,
	)

	slog.SetDefault(logger)
}
