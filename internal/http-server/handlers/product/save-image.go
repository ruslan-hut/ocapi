package product

import (
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"ocapi/entity"
	"ocapi/internal/lib/api/response"
	"ocapi/internal/lib/sl"
)

func SaveImage(log *slog.Logger, handler Core) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mod := sl.Module("http.handlers.product")

		logger := log.With(
			mod,
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		if handler == nil {
			logger.Error("product service not available")
			render.JSON(w, r, response.Error("Product service not available"))
			return
		}

		var body entity.ProductImageRequest
		if err := render.Bind(r, &body); err != nil {
			logger.Error("bind request data", sl.Err(err))
			render.Status(r, 400)
			render.JSON(w, r, response.Error(fmt.Sprintf("Failed to decode: %v", err)))
			return
		}
		logger = logger.With(slog.Int("size", len(body.Data)))

		err := handler.LoadProductImages(body.Data)
		if err != nil {
			render.JSON(w, r, response.Error(fmt.Sprintf("Save image: %v", err)))
			return
		}
		logger.Debug("product images saved")

		render.JSON(w, r, response.Ok(nil))
	}
}
