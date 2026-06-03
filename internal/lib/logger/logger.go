package logger

import (
	"log"
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func SetupLogger(env, logFile string) *slog.Logger {
	var logger *slog.Logger
	var lf *os.File
	var err error

	if env != envLocal {
		lf, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal("error opening log file: ", err)
		}
		log.Printf("env: %s; log file: %s", env, logFile)
	}

	switch env {
	case envLocal:
		logger = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		logger = slog.New(
			slog.NewTextHandler(lf, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		logger = slog.New(
			slog.NewTextHandler(lf, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		log.Fatal("invalid environment: ", env)
	}

	return logger
}
