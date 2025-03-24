package fetch

import (
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"ocapi/internal/lib/api/response"
	"ocapi/internal/lib/sl"
)

type Core interface {
	ReadDatabase(table, filter string, limit int, plain bool) (interface{}, error)
}

func ReadData(log *slog.Logger, handler Core) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mod := sl.Module("http.handlers.fetch")

		logger := log.With(
			mod,
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		if handler == nil {
			logger.Error("database service not available")
			render.JSON(w, r, response.Error("Database service not available"))
			return
		}

		var request Request
		if err := render.Bind(r, &request); err != nil {
			logger.Error("bind request", sl.Err(err))
			render.Status(r, 400)
			render.JSON(w, r, response.Error(fmt.Sprintf("Bind request: %v", err)))
			return
		}
		logger = logger.With(
			slog.String("table", request.Table),
			slog.String("filter", request.Filter),
			slog.Int("limit", request.Limit),
		)

		data, err := handler.ReadDatabase(request.Table, request.Filter, request.Limit)
		if err != nil {
			logger.Error("fetch data", sl.Err(err))
			render.JSON(w, r, response.Error(fmt.Sprintf("Fetch data: %v", err)))
			return
		}
		logger.Debug("fetch data")

		render.JSON(w, r, response.Ok(data))
	}
}
