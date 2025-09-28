package graphs

import (
	"bufio"
	"container/heap"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/aashi1008/hamburg-rails/internal/models"
)

type Edge struct {
	To       string
	Distance int
}

type Graph struct {
	Nodes map[string][]Edge
	mutex sync.RWMutex
}

// NewGraph returns an empty graph
func NewGraph() *Graph {
	return &Graph{
		Nodes: make(map[string][]Edge),
	}
}

// LoadGraphFromFile returns the graph data from file
func (g *Graph) LoadGraphFromFile(graphPath string) error {
	file, err := os.Open(graphPath)
	if err != nil {
		return fmt.Errorf("error opening graph file: %v", err)
	}
	defer file.Close()
	scan := bufio.NewScanner(file)
	var allEdges []string
	for scan.Scan() {
		text := scan.Text()
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		paths := strings.Split(text, ",")
		for _, path := range paths {
			allEdges = append(allEdges, strings.TrimSpace(path))
		}
	}

	err = g.LoadEdges(allEdges)
	if err != nil {
		return fmt.Errorf("error opening graph file: %v", err)
	}
	return nil
}

var tokenRegex = regexp.MustCompile(`^([A-Z]{1,16})([A-Z]{1,16})(\d+)$`)

// LoadEdges replaces the graph data
func (g *Graph) LoadEdges(edges []string) error {

	newNodes := make(map[string][]Edge)
	for _, e := range edges {
		e = strings.ToUpper(strings.TrimSpace(e))
		m := tokenRegex.FindStringSubmatch(e)
		if m == nil {
			return fmt.Errorf("invalid edge token: %q", e)
		}
		from := m[1]
		to := m[2]
		if from == to {
			return fmt.Errorf("self-loop not allowed: %s->%s", from, to)
		}
		dist, err := strconv.Atoi(m[3])
		if err != nil || dist <= 0 {
			return fmt.Errorf("invalid distance for token %q", e)
		}

		for _, edge := range newNodes[from] {
			if edge.To == to {
				return fmt.Errorf("duplicate edge: %s%s%d", from, to, dist)
			}
		}
		newNodes[from] = append(newNodes[from], Edge{To: to, Distance: dist})
	}

	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.Nodes = newNodes
	return nil
}

// snapshotNodes returns a shallow copy reference to the nodes map so
// traversal can proceed without holding lock for the entire operation.
func (g *Graph) snapshotNodes() map[string][]Edge {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	snap := make(map[string][]Edge, len(g.Nodes))
	for k, v := range g.Nodes {
		snap[k] = v
	}
	return snap
}

// Distance calculates distance for a fixed path
func (g *Graph) Distance(path []string) (int, error) {
	snap := g.snapshotNodes()
	total := 0
	for i := 0; i < len(path)-1; i++ {
		found := false
		from := path[i]
		to := path[i+1]
		for _, e := range snap[from] {
			if e.To == to {
				total += e.Distance
				found = true
				break
			}
		}
		if !found {
			return 0, errors.New("NO SUCH ROUTE")
		}
	}
	return total, nil
}

// CountTripsByStops counts trips with stop constraints
func (g *Graph) CountTripsByStops(from, to string, minStops, maxStops int) int {
	if maxStops < 0 || minStops < 0 {
		return 0
	}
	if minStops > maxStops {
		return 0
	}
	nodes := g.snapshotNodes()
	count := 0
	type state struct {
		Node  string
		Stops int
	}
	stack := []state{{from, 0}}
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if n.Stops > maxStops {
			continue
		}
		if n.Stops >= minStops && n.Node == to {
			count++
		}
		for _, e := range nodes[n.Node] {
			stack = append(stack, state{e.To, n.Stops + 1})
		}
	}
	return count
}

// CountTripsByDistance counts trips under distance constraint
func (g *Graph) CountTripsByDistance(from, to string, maxDistance int) int {
	if maxDistance <= 0 {
		return 0
	}
	nodes := g.snapshotNodes()
	count := 0
	type state struct {
		Node     string
		Distance int
	}
	stack := []state{{from, 0}}
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, e := range nodes[n.Node] {
			d := n.Distance + e.Distance
			if d >= maxDistance {
				continue
			}
			if e.To == to {
				count++
			}
			stack = append(stack, state{e.To, d})
		}
	}
	return count
}

// ShortestPath returns shortest distance and path using Dijkstra
func (g *Graph) ShortestPath(from, to string) (int, []string) {
	if from == "" || to == "" {
		return -1, nil
	}
	nodes := g.snapshotNodes()
	visited := make(map[string]int)
	pq := &priorityQueue{}
	heap.Init(pq)
	heap.Push(pq, &pqItem{node: from, dist: 0, path: []string{from}})

	shortestDist := -1
	var shortestPath []string

	for pq.Len() > 0 {
		curr := heap.Pop(pq).(*pqItem)
		if val, ok := visited[curr.node]; ok && val <= curr.dist {
			continue
		}
		if curr.dist > 0 || curr.node != from {
			visited[curr.node] = curr.dist
		}

		if curr.node == to && curr.dist > 0 {
			if shortestDist == -1 || curr.dist < shortestDist ||
				(curr.dist == shortestDist && strings.Join(curr.path, "") < strings.Join(shortestPath, "")) {
				shortestDist = curr.dist
				shortestPath = curr.path
			}
		}
		for _, e := range nodes[curr.node] {
			heap.Push(pq, &pqItem{node: e.To, dist: curr.dist + e.Distance, path: append(append([]string{}, curr.path...), e.To)})
		}
	}
	return shortestDist, shortestPath
}

type pqItem struct {
	node string
	dist int
	path []string
}
type priorityQueue []*pqItem

func (pq priorityQueue) Len() int { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].dist < pq[j].dist
}
func (pq priorityQueue) Swap(i, j int)       { pq[i], pq[j] = pq[j], pq[i] }
func (pq *priorityQueue) Push(x interface{}) { *pq = append(*pq, x.(*pqItem)) }
func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[:n-1]
	return item
}

func (g *Graph) SearchRoutes(from, to string, req models.RouteSearchRequest) models.RouteSearchResponse {
	// DFS to find routes with constraints
	type state struct {
		Path     []string
		Distance int
	}
	stack := []state{{Path: []string{from}, Distance: 0}}
	results := []state{}
	nodes := g.snapshotNodes()
	for len(stack) > 0 {
		curr := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		last := curr.Path[len(curr.Path)-1]
		if last == to && len(curr.Path) > 1 {
			results = append(results, curr)
		}

		for _, e := range nodes[last] {
			newDist := curr.Distance + e.Distance
			newPath := append(append([]string{}, curr.Path...), e.To)

			if req.Constraints.MaxStops > 0 && len(newPath)-1 > req.Constraints.MaxStops {
				continue
			}
			if req.Constraints.MaxDistance > 0 && newDist > req.Constraints.MaxDistance {
				continue
			}
			if req.Constraints.DistinctNodes && containsDuplicate(newPath) {
				continue
			}
			stack = append(stack, state{Path: newPath, Distance: newDist})
		}
	}

	// Sort by distance then lexicographically
	sort.Slice(results, func(i, j int) bool {
		if results[i].Distance != results[j].Distance {
			return results[i].Distance < results[j].Distance
		}
		return strings.Join(results[i].Path, "") < strings.Join(results[j].Path, "")
	})

	limit := req.Limit
	if limit <= 0 || limit > len(results) {
		limit = len(results)
	}

	res := models.RouteSearchResponse{}
	for _, r := range results[:limit] {
		res.Routes = append(res.Routes, struct {
			Path     []string `json:"path"`
			Distance int      `json:"distance"`
		}{Path: r.Path, Distance: r.Distance})
	}
	return res
}

func containsDuplicate(path []string) bool {
	set := make(map[string]struct{})
	for _, n := range path {
		if _, exists := set[n]; exists {
			return true
		}
		set[n] = struct{}{}
	}
	return false
}
