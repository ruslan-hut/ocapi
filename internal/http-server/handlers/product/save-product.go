package product

import (
	"fmt"
	"log/slog"
	"net/http"
	"ocapi/entity"
	"ocapi/internal/lib/api/response"
	"ocapi/internal/lib/sl"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

// MaxBatchSize limits the number of products that can be processed in a single request
const MaxBatchSize = 500

func SaveProduct(log *slog.Logger, handler Core) http.HandlerFunc {
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

		var body entity.ProductDataRequest
		if err := render.Bind(r, &body); err != nil {
			logger.Error("bind request data", sl.Err(err))
			render.Status(r, 400)
			render.JSON(w, r, response.Error(fmt.Sprintf("Failed to decode: %v", err)))
			return
		}
		logger = logger.With(
			slog.Bool("full", body.Full),
			slog.Int("page", body.Page),
			slog.Int("total", body.Total),
			slog.Int("size", len(body.Data)),
		)

		// Validate batch size
		if len(body.Data) > MaxBatchSize {
			logger.Warn("batch size exceeds limit", slog.Int("limit", MaxBatchSize))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error(fmt.Sprintf("Batch size %d exceeds limit of %d", len(body.Data), MaxBatchSize)))
			return
		}

		err := handler.LoadProducts(body.Data)
		if err != nil {
			logger.Error("load products", sl.Err(err))
			render.JSON(w, r, response.Error(fmt.Sprintf("Save data failed: %v", err)))
			return
		}
		logger.Debug("product data saved")

		render.JSON(w, r, response.Ok(nil))
	}
}
