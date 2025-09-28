package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	graph "github.com/aashi1008/hamburg-rails/internal/graphs"
	"github.com/aashi1008/hamburg-rails/internal/handlers"
	"github.com/aashi1008/hamburg-rails/internal/server"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	graphPath := flag.String("graph", "", "Path to the graph file")
	flag.Parse()

	g := graph.NewGraph()
	if *graphPath != "" {
		fmt.Println("Graph file path:", *graphPath)
		err := g.LoadGraphFromFile(*graphPath)
		if err != nil {
			log.Fatalf("failed to load graph from file: %v", err)
		}
	}

	h := handlers.NewHandler(g)

	server.StartServer(":8080", h, logger)
}
