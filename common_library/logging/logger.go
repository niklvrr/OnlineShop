package logging

import (
	"log/slog"
	"os"
)

func LoadLogger(envLevel string) *slog.Logger {
	logHandler := slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: getLevel(envLevel),
		},
	)

	return slog.New(logHandler)
}

const (
	EnvProd = "prod"
	EnvDev  = "dev"
)

func getLevel(level string) slog.Level {
	switch level {
	case EnvProd:
		return slog.LevelInfo
	case EnvDev:
		return slog.LevelDebug
	}
	return slog.LevelDebug
}
