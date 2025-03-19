package main

import (
	"flag"
	"log/slog"
	"ocapi/impl/core"
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

	lg.Info("starting ocapi", slog.String("config", *configPath), slog.String("env", conf.Env))
	lg.Debug("debug messages enabled")

	db, err := database.NewSQLClient(conf)
	if err != nil {
		lg.Error("mysql client", sl.Err(err))
	}
	if db != nil {
		lg.Info("mysql client initialized",
			slog.String("host", conf.SQL.HostName),
			slog.String("port", conf.SQL.Port),
			slog.String("user", conf.SQL.UserName),
			slog.String("database", conf.SQL.Database),
		)
		defer db.Close()
	}

	handler := core.New(db, conf.Listen.ApiKey, lg)

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

	// *** blocking start with http server ***
	err = api.New(conf, lg, handler)
	if err != nil {
		lg.Error("server start", sl.Err(err))
		return
	}
	lg.Error("service stopped")
}
