// Package payment serves the list of order payments read from the custom order columns
// named in the order_payment config block.
package payment

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"ocapi/entity"
	"ocapi/impl/core"
	"ocapi/internal/lib/api/response"
	"ocapi/internal/lib/sl"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

// Core is the business logic the handler needs.
type Core interface {
	PaymentsList(from time.Time, afterId int64, statuses []string, limit int) ([]*entity.OrderPayment, error)
}

// List handles GET /payments?from_date=RFC3339[&after_id=N][&status=a,b][&limit=N].
// Rows come ordered by payment time, then order id; to read the next page pass the
// time and order id of the last row as from_date and after_id. Returns 501 when the
// order_payment block (with a time column) is not configured.
func List(log *slog.Logger, handler Core) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		logger := log.With(
			sl.Module("http.handlers.payment"),
			slog.String("request_id", middleware.GetReqID(r.Context())),
			slog.String("from_date", q.Get("from_date")),
			slog.String("after_id", q.Get("after_id")),
		)

		if handler == nil {
			logger.Error("payment service not available")
			render.JSON(w, r, response.Error("Payments not available"))
			return
		}

		fromStr := q.Get("from_date")
		if fromStr == "" {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("from_date is required"))
			return
		}
		from, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("Invalid from_date parameter"))
			return
		}

		var afterId int64
		if s := q.Get("after_id"); s != "" {
			if afterId, err = strconv.ParseInt(s, 10, 64); err != nil || afterId < 0 {
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, response.Error("Invalid after_id parameter"))
				return
			}
		}

		limit := 0
		if s := q.Get("limit"); s != "" {
			if limit, err = strconv.Atoi(s); err != nil || limit < 0 {
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, response.Error("Invalid limit parameter"))
				return
			}
		}

		var statuses []string
		for _, st := range strings.Split(q.Get("status"), ",") {
			if st = strings.TrimSpace(st); st != "" {
				statuses = append(statuses, st)
			}
		}

		payments, err := handler.PaymentsList(from, afterId, statuses, limit)
		if errors.Is(err, core.ErrPaymentsNotConfigured) {
			render.Status(r, http.StatusNotImplemented)
			render.JSON(w, r, response.Error("Payments are not configured"))
			return
		}
		if err != nil {
			logger.Error("payments list", sl.Err(err))
			render.JSON(w, r, response.Error(fmt.Sprintf("Payments list failed: %v", err)))
			return
		}
		if payments == nil {
			payments = []*entity.OrderPayment{}
		}
		logger.Debug("payments list", slog.Int("count", len(payments)))
		render.JSON(w, r, response.Ok(payments))
	}
}
