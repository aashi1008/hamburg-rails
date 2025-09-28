package routes

import (
	"net/http"

	"github.com/aashi1008/hamburg-rails/internal/handlers"
	"github.com/aashi1008/hamburg-rails/internal/metrics"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupRoutes(h *handlers.Handler) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/healthz", h.Healthz).Methods(http.MethodGet)
	r.HandleFunc("/admin/graph", h.LoadGraph).Methods(http.MethodPost)
	r.HandleFunc("/graph", h.CurrentEdgeList).Methods(http.MethodGet)
	r.HandleFunc("/routes/distance", h.FixedDistance).Methods(http.MethodPost)
	r.HandleFunc("/routes/count-by-stops", h.CountByStops).Methods(http.MethodPost)
	r.HandleFunc("/routes/count-by-distance", h.CountByDistance).Methods(http.MethodPost)
	r.HandleFunc("/routes/shortest", h.ShortestPath).Methods(http.MethodGet)
	r.Handle("/metrics", promhttp.HandlerFor(metrics.CustomRegistry, promhttp.HandlerOpts{}))
	return r
}
