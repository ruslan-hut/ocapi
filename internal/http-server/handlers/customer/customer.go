package customer

import (
	"fmt"
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
	UpdateCustomers(data []*entity.Customer) error
	CustomerList(limit, offset int) ([]*entity.CustomerInfo, error)
}

func Update(log *slog.Logger, handler Core) http.HandlerFunc {
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

		var body entity.CustomerData
		if err := render.Bind(r, &body); err != nil {
			logger.Error("bind request data", sl.Err(err))
			render.Status(r, 400)
			render.JSON(w, r, response.Error(fmt.Sprintf("Failed to decode: %v", err)))
			return
		}

		err := handler.UpdateCustomers(body.Data)
		if err != nil {
			logger.Error("update customers", sl.Err(err))
			render.JSON(w, r, response.Error(fmt.Sprintf("Save data failed: %v", err)))
			return
		}
		logger.Debug("customers updated", slog.Int("count", len(body.Data)))

		render.JSON(w, r, response.Ok(nil))
	}
}

func List(log *slog.Logger, handler Core) http.HandlerFunc {
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

		limit, err := intParam(r, "limit")
		if err != nil {
			render.Status(r, 400)
			render.JSON(w, r, response.Error(fmt.Sprintf("Invalid limit: %v", err)))
			return
		}
		offset, err := intParam(r, "offset")
		if err != nil {
			render.Status(r, 400)
			render.JSON(w, r, response.Error(fmt.Sprintf("Invalid offset: %v", err)))
			return
		}

		customers, err := handler.CustomerList(limit, offset)
		if err != nil {
			logger.Error("read customers", sl.Err(err))
			render.JSON(w, r, response.Error(fmt.Sprintf("Read data failed: %v", err)))
			return
		}
		logger.Debug("customers listed", slog.Int("count", len(customers)))

		render.JSON(w, r, response.Ok(customers))
	}
}

func intParam(r *http.Request, name string) (int, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}
