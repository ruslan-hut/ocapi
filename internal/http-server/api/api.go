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
	"ocapi/internal/http-server/handlers/attribute"
	"ocapi/internal/http-server/handlers/batch"
	"ocapi/internal/http-server/handlers/category"
	"ocapi/internal/http-server/handlers/delete-records"
	"ocapi/internal/http-server/handlers/errors"
	"ocapi/internal/http-server/handlers/fetch"
	"ocapi/internal/http-server/handlers/order"
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
	attribute.Core
	category.Core
	order.Core
	fetch.Core
	batch.Core
	delete_records.Core
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

	router.NotFound(errors.NotFound(log))
	router.MethodNotAllowed(errors.NotAllowed(log))

	router.Route("/api/v1", func(v1 chi.Router) {
		v1.Route("/product", func(r chi.Router) {
			r.Get("/{uid}", product.UidSearch(log, handler))
			r.Post("/", product.SaveProduct(log, handler))
			r.Post("/description", product.SaveDescription(log, handler))
			r.Post("/attribute", product.SaveAttribute(log, handler))
			r.Post("/image", product.SaveImage(log, handler))
		})
		v1.Route("/attribute", func(r chi.Router) {
			r.Post("/", attribute.Save(log, handler))
		})
		v1.Route("/category", func(r chi.Router) {
			r.Post("/", category.SaveCategory(log, handler))
			r.Post("/description", category.SaveDescription(log, handler))
		})
		v1.Route("/order", func(r chi.Router) {
			r.Get("/{orderId}", order.SearchId(log, handler))
		})
		v1.Route("/orders", func(r chi.Router) {
			r.Get("/{orderStatusId}", category.SaveCategory(log, handler))
		})
		v1.Route("/fetch", func(r chi.Router) {
			r.Post("/", fetch.ReadData(log, handler))
		})
		v1.Route("/delete", func(r chi.Router) {
			r.Post("/", delete_records.TableRecords(log, handler))
		})
		v1.Route("/batch", func(r chi.Router) {
			r.Get("/{batchUid}", batch.Result(log, handler))
		})
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
