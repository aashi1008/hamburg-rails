package graphs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcurrentReadersAndReload(t *testing.T) {
	g := seedGraph()
	done := make(chan struct{})
	// readers
	for i := 0; i < 50; i++ {
		go func() {
			for j := 0; j < 50; j++ {
				_, _ = g.ShortestPath("A", "C")
				_ = g.CountTripsByDistance("C", "C", 30)
			}
			done <- struct{}{}
		}()
	}
	// reloaders
	for i := 0; i < 5; i++ {
		go func() {
			_ = g.LoadEdges([]string{"AB5", "BC4", "CD8"})
		}()
	}
	// wait for readers
	for i := 0; i < 50; i++ {
		<-done
	}
}


func seedGraph() *Graph {
	g := NewGraph()
	edges := strings.Split("AB5, BC4, CD8, DC8, DE6, AD5, CE2, EB3, AE7", ",")
	for i := range edges {
		edges[i] = strings.TrimSpace(edges[i])
	}
	if err := g.LoadEdges(edges); err != nil {
		panic(err)
	}
	return g
}

func TestFixedRouteDistance(t *testing.T) {
	g := seedGraph()

	tests := []struct {
		path     []string
		expected int
		err      bool
	}{
		{[]string{"A", "B", "C"}, 9, false},
		{[]string{"A", "D"}, 5, false},
		{[]string{"A", "D", "C"}, 13, false},
		{[]string{"A", "E", "B", "C", "D"}, 22, false},
		{[]string{"A", "E", "D"}, 0, true},
	}

	for _, tt := range tests {
		dist, err := g.Distance(tt.path)
		if tt.err {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, dist)
		}
	}
}

func TestTripsByStops(t *testing.T) {
	g := seedGraph()

	// C->C max 3 stops = 2 (C-D-C, C-E-B-C)
	assert.Equal(t, 2, g.CountTripsByStops("C", "C", 1, 3))

	// A->C exactly 4 stops = 3 (A-B-C-D-C, A-D-C-D-C, A-D-E-B-C)
	count := g.CountTripsByStops("A", "C", 4, 4)
	assert.Equal(t, 3, count)
}

func TestTripsByDistance(t *testing.T) {
	g := seedGraph()

	// C->C distance < 30 = 7
	count := g.CountTripsByDistance("C", "C", 30)
	assert.Equal(t, 7, count)
}

func TestShortestPath(t *testing.T) {
	g := seedGraph()

	dist, path := g.ShortestPath("A", "C")
	assert.Equal(t, 9, dist)
	assert.Equal(t, []string{"A", "B", "C"}, path)

	dist, path = g.ShortestPath("B", "B")
	assert.Equal(t, 9, dist)
	// could be B-C-D-C-B or B-C-E-B; it returns one lex min
	assert.NotEmpty(t, path)
}

func TestGraphLoadErrors(t *testing.T) {
	g := NewGraph()

	// duplicate edge
	err := g.LoadEdges([]string{"AB5", "AB5"})
	assert.Error(t, err)

	// malformed
	err = g.LoadEdges([]string{"A5"})
	assert.Error(t, err)

	// empty
	err = g.LoadEdges([]string{})
	assert.NoError(t, err)

	err = g.LoadEdges([]string{"ABCDEF50"})
	assert.NoError(t, err)
}

func TestGraphLoadFromFile(t *testing.T) {
	g := NewGraph()

	err := g.LoadGraphFromFile("../../graph.txt")
	assert.NoError(t, err)

	err = g.LoadGraphFromFile("../../graph123.txt")
	assert.Error(t, err)

}

func FuzzCountTripsByStops_MinGreaterThanMax(f *testing.F) {
	// Seed corpus
	f.Add("A", "B", 5, 3) // minStops > maxStops

	f.Fuzz(func(t *testing.T, from, to string, minStops, maxStops int) {
		g := NewGraph()
		_ = g.LoadEdges([]string{"AB1", "BC1", "CA1"})

		if minStops > maxStops {
			count := g.CountTripsByStops(from, to, minStops, maxStops)
			if count != 0 {
				t.Errorf("expected 0 trips when minStops > maxStops, got %d", count)
			}
		}
	})
}
