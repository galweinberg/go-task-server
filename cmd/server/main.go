// cmd/server/main.go
package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"

    "github.com/galweinberg/go-task-server/internal/dispatcher"
    "github.com/galweinberg/go-task-server/internal/handlers"
    "github.com/galweinberg/go-task-server/internal/metrics"
    "github.com/galweinberg/go-task-server/internal/model"
    "github.com/galweinberg/go-task-server/internal/server"
)

func main() {
	wg := &sync.WaitGroup{}

	s := &server.Server{
		TaskQueue:  make(chan model.Task, 20),
		TaskStatus: make(map[int]string),
		WG:         wg,
	}

	d := dispatcher.NewRoundRobinDispatcher(3, wg, s)
	s.Dispatcher = d
	go s.Run()

	// Register metrics
	metrics.Register()

	// HTTP routing
	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux, s)

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		metrics.Inc("/healthz")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	log.Println(" Registered /metrics and starting server")

	log.Println("HTTP server started on :8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
}
