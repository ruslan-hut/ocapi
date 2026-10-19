package customer

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
		mod := sl.Module("http.handlers.customer")

		logger := log.With(
			mod,
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		if handler == nil {
			logger.Error("service not available")
			render.JSON(w, r, response.Error("Service not available"))
			return
		}

		var body entity.CustomerSpecialSyncRequest
		if err := render.Bind(r, &body); err != nil {
			logger.Error("bind request data", sl.Err(err))
			render.Status(r, 400)
			render.JSON(w, r, response.Error(fmt.Sprintf("Failed to decode: %v", err)))
			return
		}
		logger = logger.With(slog.Int("size", len(body.Data)))

		start := time.Now()
		result, err := handler.SyncCustomerSpecials(body.Data)
		if err != nil {
			logger.Error("sync customer special", sl.Err(err))
			render.JSON(w, r, response.Error(fmt.Sprintf("Save data failed: %v", err)))
			return
		}
		logger.With(
			slog.Int64("written", result.Written),
			slog.Int64("deleted", result.Deleted),
			slog.Int("skipped_products", len(result.SkippedProducts)),
			slog.Int("skipped_customers", len(result.SkippedCustomers)),
			slog.Duration("duration", time.Since(start)),
		).Info("customer special synced")

		render.JSON(w, r, response.Ok(result))
	}
}

func DeleteSpecial(log *slog.Logger, handler Core) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mod := sl.Module("http.handlers.customer")

		logger := log.With(
			mod,
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		if handler == nil {
			logger.Error("service not available")
			render.JSON(w, r, response.Error("Service not available"))
			return
		}

		var body entity.CustomerIdsRequest
		if err := render.Bind(r, &body); err != nil {
			logger.Error("bind request data", sl.Err(err))
			render.Status(r, 400)
			render.JSON(w, r, response.Error(fmt.Sprintf("Failed to decode: %v", err)))
			return
		}

		result, err := handler.DeleteCustomerSpecials(body.Data)
		if err != nil {
			logger.Error("delete customer special", sl.Err(err))
			render.JSON(w, r, response.Error(fmt.Sprintf("Delete data failed: %v", err)))
			return
		}
		logger.With(
			slog.Int("customers", len(body.Data)),
			slog.Int64("deleted", result.Deleted),
		).Info("customer special deleted")

		render.JSON(w, r, response.Ok(result))
	}
}
