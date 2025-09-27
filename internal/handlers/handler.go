package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"unicode"

	graph "github.com/aashi1008/hamburg-rails/internal/graphs"
	"github.com/aashi1008/hamburg-rails/internal/models"
)

type Handler struct {
	Graph *graph.Graph
}

func NewHandler(g *graph.Graph) *Handler {
	return &Handler{Graph: g}
}

func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) LoadGraph(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	edges := strings.Split(string(data), ",")
	for i := range edges {
		e := strings.TrimFunc(edges[i], func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		})
		edges[i] = strings.TrimSpace(e)
	}
	if err := h.Graph.LoadEdges(edges); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) CurrentEdgeList(w http.ResponseWriter, r *http.Request) {
	type item struct {
		Edges map[string][]graph.Edge `json:"edges"`
		Count int                     `json:"node_count"`
	}
	json.NewEncoder(w).Encode(&item{Edges: h.Graph.Nodes, Count: len(h.Graph.Nodes)})
}

func (h *Handler) FixedDistance(w http.ResponseWriter, r *http.Request) {
	var req models.RouteDistanceRequest
	json.NewDecoder(r.Body).Decode(&req)
	path := make([]string, len(req.Path))
	for i := range req.Path {
		path[i] = strings.ToUpper(req.Path[i])
	}
	dist, err := h.Graph.Distance(path)
	if err != nil {
		http.Error(w, "NO SUCH ROUTE", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(map[string]int{"distance": dist})
}

func (h *Handler) CountByStops(w http.ResponseWriter, r *http.Request) {
	var req models.CountByStopsRequest
	json.NewDecoder(r.Body).Decode(&req)
	from := strings.ToUpper(req.From)
	to := strings.ToUpper(req.To)
	minStops := req.MinStops
	if minStops == 0 {
		minStops = 1
	}
	maxStops := req.MaxStops
	count := h.Graph.CountTripsByStops(from, to, minStops, maxStops)
	json.NewEncoder(w).Encode(map[string]int{"count": count})
}

func (h *Handler) CountByDistance(w http.ResponseWriter, r *http.Request) {
	var req models.CountByDistanceRequest
	json.NewDecoder(r.Body).Decode(&req)
	from := strings.ToUpper(req.From)
	to := strings.ToUpper(req.To)
	count := h.Graph.CountTripsByDistance(from, to, req.MaxDistance)
	json.NewEncoder(w).Encode(map[string]int{"count": count})
}

func (h *Handler) ShortestPath(w http.ResponseWriter, r *http.Request) {
	from := strings.ToUpper(r.URL.Query().Get("from"))
	to := strings.ToUpper(r.URL.Query().Get("to"))
	dist, path := h.Graph.ShortestPath(from, to)
	if dist == -1 {
		http.Error(w, "NO SUCH ROUTE", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"distance": dist, "path": path})
}
