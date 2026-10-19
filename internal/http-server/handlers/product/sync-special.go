package product

import (
	"fmt"
	"log/slog"
	"net/http"
	"ocapi/entity"
	"ocapi/internal/lib/api/response"
	"ocapi/internal/lib/sl"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func SyncSpecial(log *slog.Logger, handler Core) http.HandlerFunc {
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

		var body entity.ProductSpecialSyncRequest
		if err := render.Bind(r, &body); err != nil {
			logger.Error("bind request data", sl.Err(err))
			render.Status(r, 400)
			render.JSON(w, r, response.Error(fmt.Sprintf("Failed to decode: %v", err)))
			return
		}
		logger = logger.With(
			slog.Int("page", body.Page),
			slog.Int("total", body.Total),
			slog.Int("size", len(body.Data)),
		)

		start := time.Now()
		result, err := handler.SyncProductSpecials(body.Groups, body.Data)
		if err != nil {
			logger.Error("sync special", sl.Err(err))
			render.JSON(w, r, response.Error(fmt.Sprintf("Save data failed: %v", err)))
			return
		}
		logger.With(
			slog.Int64("written", result.Written),
			slog.Int64("deleted", result.Deleted),
			slog.Int("skipped", len(result.SkippedProducts)),
			slog.Duration("duration", time.Since(start)),
		).Info("product special synced")

		render.JSON(w, r, response.Ok(result))
	}
}
