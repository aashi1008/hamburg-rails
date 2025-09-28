package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"unicode"

	graph "github.com/aashi1008/hamburg-rails/internal/graphs"
	"github.com/aashi1008/hamburg-rails/internal/metrics"
	"github.com/aashi1008/hamburg-rails/internal/models"
)

var townRegex = regexp.MustCompile(`^[A-Z]{1,16}$`)

type Handler struct {
	Graph *graph.Graph
}

func NewHandler(g *graph.Graph) *Handler {
	return &Handler{Graph: g}
}

func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func sanitizeEdgesInput(raw string) []string {
	parts := strings.Split(raw, ",")
	edges := make([]string, 0, len(parts))
	for _, p := range parts {
		e := strings.TrimFunc(p, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		})
		e = strings.TrimSpace(e)
		if e != "" {
			edges = append(edges, e)
		}
	}
	return edges
}

func (h *Handler) LoadGraph(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	edges := sanitizeEdgesInput(string(data))
	if err := h.Graph.LoadEdges(edges); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("graph parse error: %v", err))
		return
	}
	nodes := len(h.Graph.Nodes)
	metrics.GraphLoadsTotal.Inc()
	metrics.GraphNodesTotal.Set(float64(nodes))
	writeJSON(w, map[string]string{"status": "ok", "message": "graph loaded"})
}

func (h *Handler) CurrentEdgeList(w http.ResponseWriter, r *http.Request) {
	type item struct {
		Edges map[string][]graph.Edge `json:"edges"`
		Count int                     `json:"node_count"`
	}
	writeJSON(w, &item{Edges: h.Graph.Nodes, Count: len(h.Graph.Nodes)})
}

func validateTown(s string) (string, error) {
	if s == "" {
		return "", fmt.Errorf("empty town name")
	}
	u := strings.ToUpper(strings.TrimSpace(s))
	if !townRegex.MatchString(u) {
		return "", fmt.Errorf("invalid town id: %q", s)
	}
	return u, nil
}

func (h *Handler) FixedDistance(w http.ResponseWriter, r *http.Request) {
	var req models.RouteDistanceRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.Path) < 2 {
		writeError(w, http.StatusUnprocessableEntity, "path must contain at least two towns")
		return
	}
	path := make([]string, len(req.Path))
	for i, p := range req.Path {
		t, err := validateTown(p)
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		path[i] = t
	}
	dist, err := h.Graph.Distance(path)
	if err != nil {
		writeError(w, http.StatusNotFound, "NO SUCH ROUTE")
		return
	}
	writeJSON(w, map[string]int{"distance": dist})
}

func (h *Handler) CountByStops(w http.ResponseWriter, r *http.Request) {
	var req models.CountByStopsRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	from, err := validateTown(req.From)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	to, err := validateTown(req.To)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	minStops := req.MinStops
	if minStops == 0 {
		minStops = 1
	}
	maxStops := req.MaxStops
	if maxStops < 0 {
		writeError(w, http.StatusUnprocessableEntity, "maxStops must be >= 0")
		return
	}
	if minStops > maxStops {
		writeError(w, http.StatusUnprocessableEntity, "minStops cannot be greater than maxStops")
		return
	}
	count := h.Graph.CountTripsByStops(from, to, minStops, maxStops)
	json.NewEncoder(w).Encode(map[string]int{"count": count})
}

func (h *Handler) CountByDistance(w http.ResponseWriter, r *http.Request) {
	var req models.CountByDistanceRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	from, err := validateTown(req.From)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	to, err := validateTown(req.To)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if req.MaxDistance <= 0 {
		writeError(w, http.StatusUnprocessableEntity, "maxDistance must be > 0")
		return
	}
	count := h.Graph.CountTripsByDistance(from, to, req.MaxDistance)
	json.NewEncoder(w).Encode(map[string]int{"count": count})
}

func (h *Handler) ShortestPath(w http.ResponseWriter, r *http.Request) {
	fromRaw := r.URL.Query().Get("from")
	toRaw := r.URL.Query().Get("to")
	from, err := validateTown(fromRaw)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid from: "+err.Error())
		return
	}
	to, err := validateTown(toRaw)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid to: "+err.Error())
		return
	}
	dist, path := h.Graph.ShortestPath(from, to)
	if dist == -1 || len(path) == 0 {
		writeError(w, http.StatusNotFound, "NO SUCH ROUTE")
		return
	}
	writeJSON(w, map[string]interface{}{"distance": dist, "path": path})
}

func (h *Handler) SearchRoutes(w http.ResponseWriter, r *http.Request) {
	var req models.RouteSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	from, err := validateTown(req.From)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	to, err := validateTown(req.To)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	res := h.Graph.SearchRoutes(from, to, req)
	writeJSON(w, res)
}
