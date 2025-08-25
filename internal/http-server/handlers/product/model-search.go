package product

import (
	"fmt"
	"log/slog"
	"net/http"
	"ocapi/internal/lib/api/response"
	"ocapi/internal/lib/sl"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func UidSearch(log *slog.Logger, handler Core) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mod := sl.Module("http.handlers.product")
		uid := chi.URLParam(r, "uid")

		logger := log.With(
			mod,
			slog.String("request_id", middleware.GetReqID(r.Context())),
			slog.String("uid", uid),
		)

		if handler == nil {
			logger.Error("product service not available")
			render.JSON(w, r, response.Error("Product search not available"))
			return
		}

		product, err := handler.FindProduct(uid)
		if err != nil {
			logger.Error("product search", sl.Err(err))
			render.JSON(w, r, response.Error(fmt.Sprintf("Search failed: %v", err)))
			return
		}
		logger.Debug("product search")

		render.JSON(w, r, response.Ok(product))
	}
}
