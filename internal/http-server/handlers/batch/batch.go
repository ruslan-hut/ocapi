package batch

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"ocapi/internal/lib/api/response"
	"ocapi/internal/lib/sl"
)

func Result(log *slog.Logger, handler Core) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mod := sl.Module("http.handlers.batch")
		uid := chi.URLParam(r, "batchUid")

		logger := log.With(
			mod,
			slog.String("request_id", middleware.GetReqID(r.Context())),
			slog.String("batch_uid", uid),
		)

		if handler == nil {
			logger.Error("service not available")
			render.JSON(w, r, response.Error("Service not available"))
			return
		}

		result, err := handler.FinishBatch(uid)
		if err != nil {
			logger.Error("finish batch", sl.Err(err))
			render.Status(r, 400)
			render.JSON(w, r, response.Error(fmt.Sprintf("Batch result: %v", err)))
			return
		}
		logger.With(
			slog.Bool("success", result.Success),
			slog.Int("products", result.Products),
			slog.Int("categories", result.Categories),
			slog.String("message", result.Message),
		).Debug("batch result")

		render.JSON(w, r, response.Ok(result))
	}
}
