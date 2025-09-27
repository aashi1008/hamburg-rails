package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	graph "github.com/aashi1008/hamburg-rails/internal/graphs"
	"github.com/aashi1008/hamburg-rails/internal/handlers"
	"github.com/aashi1008/hamburg-rails/internal/routes"
)

func main() {
	graphPath := flag.String("graph", "", "Path to the graph file")
	flag.Parse()

	g := graph.NewGraph()
	if *graphPath != "" {
		fmt.Println("Graph file path:", *graphPath)
		g.LoadGraphFromFile(*graphPath)
	}

	h := handlers.NewHandler(g)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: routes.SetupRoutes(h),
	}

	go func() {
		log.Println("Server starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown: %v", err)
	}
	log.Println("Server gracefully stopped")
}
