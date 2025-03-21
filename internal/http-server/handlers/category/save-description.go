package category

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

func SaveDescription(log *slog.Logger, handler Core) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mod := sl.Module("http.handlers.category")

		logger := log.With(
			mod,
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		if handler == nil {
			logger.Error("category service not available")
			render.JSON(w, r, response.Error("Category service not available"))
			return
		}

		var body entity.CategoryDescriptionRequest
		if err := render.Bind(r, &body); err != nil {
			logger.Error("bind request data", sl.Err(err))
			render.Status(r, 400)
			render.JSON(w, r, response.Error(fmt.Sprintf("Failed to decode: %v", err)))
			return
		}

		err := handler.LoadCategoryDescriptions(body.Data)
		if err != nil {
			logger.Error("load descriptions", sl.Err(err))
			render.JSON(w, r, response.Error(fmt.Sprintf("Save data: %v", err)))
			return
		}
		logger.With(slog.Int("size", len(body.Data))).Debug("category data saved")

		render.JSON(w, r, response.Ok(nil))
	}
}
