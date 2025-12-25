package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"ocapi/impl/core"
	"ocapi/internal/config"
	"ocapi/internal/database"
	"ocapi/internal/http-server/api"
	"ocapi/internal/lib/logger"
	"ocapi/internal/lib/sl"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	configPath := flag.String("conf", "config.yml", "path to config file")
	logPath := flag.String("log", "/var/log/", "path to log file directory")
	flag.Parse()

	conf := config.MustLoad(*configPath)
	lg := logger.SetupLogger(conf.Env, *logPath)

	lg.Info("starting ocapi", slog.String("config", *configPath), slog.String("env", conf.Env))
	lg.Debug("debug messages enabled")

	handler := core.New(lg)
	handler.SetAuthKey(conf.Listen.ApiKey)
	handler.SetImageParameters(conf.Images.Path, conf.Images.Url)

	db, err := database.NewSQLClient(conf)
	if err != nil {
		lg.Error("mysql client", sl.Err(err))
	}
	if db != nil {
		handler.SetRepository(db)
		lg.Info("mysql client initialized",
			slog.String("host", conf.SQL.HostName),
			slog.String("port", conf.SQL.Port),
			slog.String("user", conf.SQL.UserName),
			slog.String("database", conf.SQL.Database),
		)
		defer db.Close()

		lg.Info("mysql stats", slog.String("connections", db.Stats()))
		go func() {
			ticker := time.NewTicker(30 * time.Minute)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					lg.Info("mysql", slog.String("stats", db.Stats()))
				}
			}
		}()
	}

	//if conf.Telegram.Enabled {
	//	tg, e := telegram.New(conf.Telegram.ApiKey, lg)
	//	if e != nil {
	//		lg.Error("telegram api", sl.Err(e))
	//	}
	//	//if mongo != nil {
	//	//	tg.SetDatabase(mongo)
	//	//}
	//	tg.Start()
	//	lg.Info("telegram api initialized")
	//	handler.SetMessageService(tg)
	//}

	// Create HTTP server
	server, err := api.New(conf, lg, handler)
	if err != nil {
		lg.Error("server create", sl.Err(err))
		return
	}

	// Setup signal handling for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	// Channel to signal server startup failure
	serverErr := make(chan error, 1)

	// Start server in goroutine
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			lg.Error("server error", sl.Err(err))
			serverErr <- err
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case sig := <-stop:
		lg.Info("received shutdown signal", slog.String("signal", sig.String()))
	case err := <-serverErr:
		lg.Error("server failed to start", sl.Err(err))
		return
	}

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		lg.Error("server shutdown error", sl.Err(err))
	}

	lg.Info("service stopped gracefully")
}
