package graphs

import (
	"bufio"
	"container/heap"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
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

	for scan.Scan() {
		text := scan.Text()
		paths := strings.Split(text, ",")
		err = g.LoadEdges(paths)
		if err != nil {
			return fmt.Errorf("error opening graph file: %v", err)
		}
	}
	return nil
}

// LoadEdges replaces the graph data
func (g *Graph) LoadEdges(edges []string) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	newNodes := make(map[string][]Edge)
	for _, e := range edges {
		e = strings.ToUpper(strings.TrimSpace(e))
		if len(e) < 3 {
			return errors.New("invalid edge format")
		}
		from, to := string(e[0]), string(e[1])
		dist := int(e[2] - '0') // simple parse; extend for multi-digit
		// check duplicates
		for _, edge := range newNodes[from] {
			if edge.To == to {
				return errors.New("duplicate edge: " + e)
			}
		}
		newNodes[from] = append(newNodes[from], Edge{To: to, Distance: dist})
	}

	g.Nodes = newNodes
	return nil
}

// Distance calculates distance for a fixed path
func (g *Graph) Distance(path []string) (int, error) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	total := 0
	for i := 0; i < len(path)-1; i++ {
		found := false
		for _, e := range g.Nodes[path[i]] {
			if e.To == path[i+1] {
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
	g.mutex.RLock()
	defer g.mutex.RUnlock()
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
		for _, e := range g.Nodes[n.Node] {
			stack = append(stack, state{e.To, n.Stops + 1})
		}
	}
	return count
}

// CountTripsByDistance counts trips under distance constraint
func (g *Graph) CountTripsByDistance(from, to string, maxDistance int) int {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	count := 0
	type state struct {
		Node     string
		Distance int
	}
	stack := []state{{from, 0}}
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, e := range g.Nodes[n.Node] {
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
	g.mutex.RLock()
	defer g.mutex.RUnlock()

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
		for _, e := range g.Nodes[curr.node] {
			heap.Push(pq, &pqItem{node: e.To, dist: curr.dist + e.Distance, path: append(append([]string{}, curr.path...), e.To)})
		}
	}
	return shortestDist, shortestPath
}

func (g *Graph) FindShortestPath(from, to string) (int, []string) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	var sp []string
	pr := &priorityQueue{}
	heap.Init(pr)
	heap.Push(pr, &pqItem{node: from, dist: 0, path: []string{from}})
	v := make(map[string]int)
	sl := -1

	for pr.Len() > 0 {
		node := heap.Pop(pr).(*pqItem)
		if n, ok := v[node.node]; ok && n <= node.dist {
			continue
		}
		if node.node == to && node.dist > 0 {
			if sl == -1 || sl > node.dist || (sl == node.dist && strings.Join(sp, "") > strings.Join(node.path, "")) {
				sl = node.dist
				sp = node.path
			}
		}
		if node.dist > 0 || node.node != from {
			v[node.node] = node.dist
		}

		for _, p := range g.Nodes[node.node] {
			heap.Push(pr, &pqItem{node: p.To, dist: node.dist + p.Distance, path: append(append([]string{}, node.path...), p.To)})
		}
	}

	return sl, sp
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
