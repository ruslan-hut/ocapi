package currency

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

type Core interface {
	UpdateCurrency(body *entity.Currency) error
}

func Update(log *slog.Logger, handler Core) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mod := sl.Module("http.handlers.currency")

		logger := log.With(
			mod,
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		if handler == nil {
			logger.Error("service not available")
			render.JSON(w, r, response.Error("Service not available"))
			return
		}

		var body entity.Currency
		if err := render.Bind(r, &body); err != nil {
			logger.Error("bind request data", sl.Err(err))
			render.Status(r, 400)
			render.JSON(w, r, response.Error(fmt.Sprintf("Failed to decode: %v", err)))
			return
		}
		logger = logger.With(
			slog.String("code", body.Code),
			slog.Float64("rate", body.Rate),
		)

		err := handler.UpdateCurrency(&body)
		if err != nil {
			logger.Error("update currency", sl.Err(err))
			render.JSON(w, r, response.Error(fmt.Sprintf("Save data failed: %v", err)))
			return
		}
		logger.Debug("currency rate updated")

		render.JSON(w, r, response.Ok(nil))
	}
}
