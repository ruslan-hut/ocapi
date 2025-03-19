package main

import (
	"flag"
	"log/slog"
	"ocapi/impl/core"
	"ocapi/impl/telegram"
	"ocapi/internal/config"
	"ocapi/internal/database"
	"ocapi/internal/http-server/api"
	"ocapi/internal/lib/logger"
	"ocapi/internal/lib/sl"
)

func main() {

	configPath := flag.String("conf", "config.yml", "path to config file")
	logPath := flag.String("log", "/var/log/", "path to log file directory")
	flag.Parse()

	conf := config.MustLoad(*configPath)
	lg := logger.SetupLogger(conf.Env, *logPath)

	lg.Info("starting mittere", slog.String("config", *configPath), slog.String("env", conf.Env))
	lg.Debug("debug messages enabled")

	mongo, err := database.NewMongoClient(conf)
	if err != nil {
		lg.Error("mongo client", sl.Err(err))
	}
	if mongo != nil {
		lg.Info("mongo client initialized",
			slog.String("host", conf.Mongo.Host),
			slog.String("port", conf.Mongo.Port),
			slog.String("user", conf.Mongo.User),
			slog.String("database", conf.Mongo.Database),
		)
	}

	handler := core.New(mongo, lg)

	if conf.Telegram.Enabled {
		tg, e := telegram.New(conf.Telegram.ApiKey, lg)
		if e != nil {
			lg.Error("telegram api", sl.Err(e))
		}
		if mongo != nil {
			tg.SetDatabase(mongo)
		}
		tg.Start()
		lg.Info("telegram api initialized")
		handler.SetMessageService(tg)
	}

	// *** blocking start with http server ***
	err = api.New(conf, lg, handler)
	if err != nil {
		lg.Error("server start", sl.Err(err))
		return
	}
	lg.Error("service stopped")
}
