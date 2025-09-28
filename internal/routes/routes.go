package routes

import (
	"log/slog"
	"net/http"

	"github.com/aashi1008/hamburg-rails/internal/handlers"
	"github.com/aashi1008/hamburg-rails/internal/metrics"
	middleware "github.com/aashi1008/hamburg-rails/internal/middleware/logging"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupRoutes(h *handlers.Handler, logger *slog.Logger) *mux.Router {
	r := mux.NewRouter()
	r.Use(middleware.LoggingMiddleware(logger))
	r.HandleFunc("/healthz", h.Healthz).Methods(http.MethodGet)
	r.HandleFunc("/admin/graph", h.LoadGraph).Methods(http.MethodPost)
	r.HandleFunc("/graph", h.CurrentEdgeList).Methods(http.MethodGet)
	r.HandleFunc("/routes/distance", h.FixedDistance).Methods(http.MethodPost)
	r.HandleFunc("/routes/count-by-stops", h.CountByStops).Methods(http.MethodPost)
	r.HandleFunc("/routes/count-by-distance", h.CountByDistance).Methods(http.MethodPost)
	r.HandleFunc("/routes/shortest", h.ShortestPath).Methods(http.MethodGet)
	r.Handle("/metrics", promhttp.HandlerFor(metrics.CustomRegistry, promhttp.HandlerOpts{}))
	r.HandleFunc("/routes/search", h.SearchRoutes).Methods(http.MethodPost)
	return r
}
