package routes

import (
	"net/http"

	"github.com/aashi1008/hamburg-rails/internal/handlers"
	"github.com/gorilla/mux"
)

func SetupRoutes(h *handlers.Handler) http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/healthz", h.Healthz).Methods(http.MethodGet)
	r.HandleFunc("/admin/graph", h.LoadGraph).Methods(http.MethodPost)
	r.HandleFunc("/graph", h.CurrentEdgeList).Methods(http.MethodGet)
	r.HandleFunc("/routes/distance", h.FixedDistance).Methods(http.MethodPost)
	r.HandleFunc("/routes/count-by-stops", h.CountByStops).Methods(http.MethodPost)
	r.HandleFunc("/routes/count-by-distance", h.CountByDistance).Methods(http.MethodPost)
	r.HandleFunc("/routes/shortest", h.ShortestPath).Methods(http.MethodGet)
	return r
}
