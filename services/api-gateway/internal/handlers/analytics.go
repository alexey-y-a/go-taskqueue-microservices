package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/clickhouse"
	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
	"github.com/sirupsen/logrus"
)

type AnalyticsHandler struct {
	chClient *clickhouse.Client
}

func NewAnalyticsHandler(chClient *clickhouse.Client) *AnalyticsHandler {
	return &AnalyticsHandler{chClient: chClient}
}

func (h *AnalyticsHandler) GetDailyStats(w http.ResponseWriter, r *http.Request) {
	log := logger.WithFields("api-gateway", logrus.Fields{
		"method": r.Method,
		"path":   r.URL.Path,
	})

	from, err := time.Parse("2006-01-02", r.URL.Query().Get("from"))
	if err != nil {
		from = time.Now().AddDate(0, 0, -7)
	}

	to, err := time.Parse("2006-01-02", r.URL.Query().Get("to"))
	if err != nil {
		to = time.Now()
	}

	stats, err := h.chClient.GetDailyStats(r.Context(), from, to)
	if err != nil {
		log.WithError(err).Error("Failed to get daily stats")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
