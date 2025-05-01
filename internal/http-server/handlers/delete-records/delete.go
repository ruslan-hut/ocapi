package delete_records

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
	DeleteFromTable(table, filter string) (int64, error)
}

func TableRecords(log *slog.Logger, handler Core) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mod := sl.Module("http.handlers.delete")

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
		)

		data, err := handler.DeleteFromTable(request.Table, request.Filter)
		if err != nil {
			logger.Error("delete data", sl.Err(err))
			render.JSON(w, r, response.Error(fmt.Sprintf("Failed: %v", err)))
			return
		}
		logger.With(
			slog.Int64("rows_affected", data),
		).Debug("delete data")

		render.JSON(w, r, response.Ok(data))
	}
}
