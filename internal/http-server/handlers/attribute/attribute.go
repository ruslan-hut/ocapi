package attribute

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
	LoadAttributes(attributes []*entity.Attribute) error
}

func Save(log *slog.Logger, handler Core) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mod := sl.Module("http.handlers.attribute")

		logger := log.With(
			mod,
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		if handler == nil {
			logger.Error("attribute service not available")
			render.JSON(w, r, response.Error("Attribute service not available"))
			return
		}

		var body entity.AttributeDataRequest
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

		err := handler.LoadAttributes(body.Data)
		if err != nil {
			logger.Error("load attributes", sl.Err(err))
			render.JSON(w, r, response.Error(fmt.Sprintf("Save data failed: %v", err)))
			return
		}
		logger.Debug("attributes data saved")

		render.JSON(w, r, response.Ok(nil))
	}
}
