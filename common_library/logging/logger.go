package logging

import (
	"log/slog"
	"os"
)

func NewLogger(envLevel string) *slog.Logger {
	logHandler := slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: getLevel(envLevel),
		},
	)

	return slog.New(logHandler)
}

const (
	envProd = "prod"
	envDev  = "dev"
)

func getLevel(level string) slog.Level {
	switch level {
	case envProd:
		return slog.LevelInfo
	case envDev:
		return slog.LevelDebug
	}
	return slog.LevelDebug
}
