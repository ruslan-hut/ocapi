package api

import (
	"context"
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
	"ocapi/internal/http-server/handlers/currency"
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
	currency.Core
	fetch.Core
	batch.Core
}

func New(conf *config.Config, log *slog.Logger, handler Handler) (*Server, error) {

	server := &Server{
		conf: conf,
		log:  log.With(sl.Module("api.server")),
	}

	router := chi.NewRouter()
	router.Use(timeout.Timeout(5))
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(render.SetContentType(render.ContentTypeJSON))

	// Health check endpoint (no authentication required)
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	router.NotFound(errors.NotFound(log))
	router.MethodNotAllowed(errors.NotAllowed(log))

	// Authenticated routes
	router.Group(func(r chi.Router) {
		r.Use(authenticate.New(log, handler))
		r.Route("/api/v1", func(v1 chi.Router) {
			v1.Route("/product", func(r chi.Router) {
				r.Get("/{uid}", product.UidSearch(log, handler))
				r.Post("/", product.SaveProduct(log, handler))
				r.Post("/description", product.SaveDescription(log, handler))
				r.Post("/attribute", product.SaveAttribute(log, handler))
				r.Post("/image", product.SaveImage(log, handler))
				r.Post("/images", product.SetImages(log, handler))
				r.Post("/special", product.SaveSpecial(log, handler))
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
				r.Get("/{orderId}/products", order.Products(log, handler))
				r.Post("/", order.ChangeStatus(log, handler))
			})
			v1.Route("/orders", func(r chi.Router) {
				r.Get("/{orderStatusId}", order.SearchStatus(log, handler))
			})
			v1.Route("/fetch", func(r chi.Router) {
				r.Post("/", fetch.ReadData(log, handler))
			})
			v1.Route("/batch", func(r chi.Router) {
				r.Get("/{batchUid}", batch.Result(log, handler))
			})
			v1.Route("/currency", func(r chi.Router) {
				r.Post("/", currency.Update(log, handler))
			})
		})
	})

	httpLog := slog.NewLogLogger(log.Handler(), slog.LevelError)
	server.httpServer = &http.Server{
		Handler:  router,
		ErrorLog: httpLog,
	}

	return server, nil
}

// Start starts the HTTP server (blocking)
func (s *Server) Start() error {
	serverAddress := fmt.Sprintf("%s:%s", s.conf.Listen.BindIP, s.conf.Listen.Port)
	listener, err := net.Listen("tcp", serverAddress)
	if err != nil {
		return err
	}

	s.log.Info("starting api server", slog.String("address", serverAddress))
	return s.httpServer.Serve(listener)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("shutting down api server")
	return s.httpServer.Shutdown(ctx)
}
