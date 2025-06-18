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
	OrderProducts(id int64) ([]*entity.ProductOrder, error)
	OrderSetStatus(id int64, status int) error
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

func Products(log *slog.Logger, handler Core) http.HandlerFunc {
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

		data, err := handler.OrderProducts(id)
		if err != nil {
			logger.Error("order products", sl.Err(err))
			render.JSON(w, r, response.Error(fmt.Sprintf("Search failed: %v", err)))
			return
		}
		logger.With(
			slog.Int("count", len(data)),
		).Debug("order products")

		render.JSON(w, r, response.Ok(data))
	}
}

func SearchStatus(log *slog.Logger, handler Core) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mod := sl.Module("http.handlers.order")
		statusId := chi.URLParam(r, "orderStatusId")

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

func ChangeStatus(log *slog.Logger, handler Core) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mod := sl.Module("http.handlers.order")

		logger := log.With(
			mod,
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		if handler == nil {
			logger.Error("order service not available")
			render.JSON(w, r, response.Error("Order service not available"))
			return
		}

		var request Request
		if err := render.Bind(r, &request); err != nil {
			logger.Error("bind request", sl.Err(err))
			render.Status(r, 400)
			render.JSON(w, r, response.Error(fmt.Sprintf("Bind request: %v", err)))
			return
		}

		for _, order := range request.Data {
			logger = logger.With(
				slog.Int64("order_id", order.OrderId),
				slog.Int("order_status_id", order.OrderStatusId),
			)
			err := handler.OrderSetStatus(order.OrderId, order.OrderStatusId)
			if err != nil {
				logger.Error("set status", sl.Err(err))
				render.JSON(w, r, response.Error(fmt.Sprintf("Set status failed: %v", err)))
				return
			}
			logger.Debug("order status changed")
		}

		render.JSON(w, r, response.Ok(nil))
	}
}
