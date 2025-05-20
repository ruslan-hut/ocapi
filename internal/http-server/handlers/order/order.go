package order

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"ocapi/entity"
	"ocapi/internal/lib/api/response"
	"ocapi/internal/lib/sl"
	"strconv"
)

type Core interface {
	OrderSearch(id int64) (*entity.Order, error)
	OrderSearchStatus(id int64) ([]int64, error)
}

func SearchId(log *slog.Logger, handler Core) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mod := sl.Module("http.handlers.order")
		orderId := chi.URLParam(r, "orderId")

		logger := log.With(
			mod,
			slog.String("request_id", middleware.GetReqID(r.Context())),
			slog.String("orderId", orderId),
		)

		if handler == nil {
			logger.Error("order service not available")
			render.JSON(w, r, response.Error("Order search not available"))
			return
		}

		id, err := strconv.ParseInt(orderId, 10, 64)
		if err != nil {
			logger.Warn("invalid order id")
			render.Status(r, 400)
			render.JSON(w, r, response.Error("Invalid order id"))
			return
		}

		order, err := handler.OrderSearch(id)
		if err != nil {
			logger.Error("order search", sl.Err(err))
			render.JSON(w, r, response.Error(fmt.Sprintf("Search failed: %v", err)))
			return
		}
		logger.Debug("order id search")

		render.JSON(w, r, response.Ok(order))
	}
}

func SearchStatus(log *slog.Logger, handler Core) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mod := sl.Module("http.handlers.order")
		statusId := chi.URLParam(r, "statusId")

		logger := log.With(
			mod,
			slog.String("request_id", middleware.GetReqID(r.Context())),
			slog.String("statusId", statusId),
		)

		if handler == nil {
			logger.Error("order service not available")
			render.JSON(w, r, response.Error("Order search not available"))
			return
		}

		id, err := strconv.ParseInt(statusId, 10, 64)
		if err != nil {
			logger.Warn("invalid status id")
			render.Status(r, 400)
			render.JSON(w, r, response.Error("Invalid status id"))
			return
		}

		orders, err := handler.OrderSearchStatus(id)
		if err != nil {
			logger.Error("order search", sl.Err(err))
			render.JSON(w, r, response.Error(fmt.Sprintf("Search failed: %v", err)))
			return
		}
		logger.With(
			slog.Int("count", len(orders)),
		).Debug("order status search")

		render.JSON(w, r, response.Ok(orders))
	}
}
