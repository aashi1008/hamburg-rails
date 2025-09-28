package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	CustomRegistry = prometheus.NewRegistry()

	GraphNodesTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "graph_nodes_total",
			Help: "Total number of graph nodes",
		},
	)

	GraphLoadsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "graph_loads_total",
			Help: "Total number of graph loads",
		},
	)
)

func init() {
	CustomRegistry.MustRegister(GraphNodesTotal)
	CustomRegistry.MustRegister(GraphLoadsTotal)
}
