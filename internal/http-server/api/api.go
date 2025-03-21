package api

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net"
	"net/http"
	"ocapi/internal/config"
	"ocapi/internal/http-server/handlers/category"
	"ocapi/internal/http-server/handlers/fetch"
	"ocapi/internal/http-server/handlers/product"
	"ocapi/internal/http-server/handlers/service"
	"ocapi/internal/http-server/middleware/authenticate"
	"ocapi/internal/http-server/middleware/timeout"
	"ocapi/internal/lib/sl"
)

type Server struct {
	conf       *config.Config
	httpServer *http.Server
	log        *slog.Logger
}

type Handler interface {
	authenticate.Authenticate
	service.Service
	product.Core
	category.Core
	fetch.Core
}

func New(conf *config.Config, log *slog.Logger, handler Handler) error {

	server := Server{
		conf: conf,
		log:  log.With(sl.Module("api.server")),
	}

	router := chi.NewRouter()
	router.Use(timeout.Timeout(5))
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Use(authenticate.New(log, handler))

	router.Route("/product", func(r chi.Router) {
		r.Get("/{model}", product.ModelSearch(log, handler))
		r.Post("/", product.SaveProduct(log, handler))
		r.Post("/description", product.SaveDescription(log, handler))
	})
	router.Route("/category", func(r chi.Router) {
		r.Post("/", category.SaveCategory(log, handler))
		r.Post("/description", category.SaveDescription(log, handler))
	})
	router.Route("/fetch", func(r chi.Router) {
		r.Post("/", fetch.ReadData(log, handler))
	})

	httpLog := slog.NewLogLogger(log.Handler(), slog.LevelError)
	server.httpServer = &http.Server{
		Handler:  router,
		ErrorLog: httpLog,
	}

	serverAddress := fmt.Sprintf("%s:%s", conf.Listen.BindIP, conf.Listen.Port)
	listener, err := net.Listen("tcp", serverAddress)
	if err != nil {
		return err
	}

	server.log.Info("starting api server", slog.String("address", serverAddress))

	return server.httpServer.Serve(listener)
}
