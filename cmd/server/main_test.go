package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"goProj/internal/dispatcher"
	"goProj/internal/handlers"
	"goProj/internal/model"
	"goProj/internal/server"
)

func startTestServer() *http.Server {
	wg := &sync.WaitGroup{}
	s := &server.Server{
		TaskQueue:  make(chan model.Task, 10),
		TaskStatus: make(map[int]string),
		WG:         wg,
	}

	s.Dispatcher = dispatcher.NewRoundRobinDispatcher(1, wg, s)
	go s.Run()

	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux, s)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	srv := &http.Server{Addr: ":8081", Handler: mux}
	go srv.ListenAndServe()

	// Give server a moment to start
	time.Sleep(200 * time.Millisecond)

	return srv
}

func TestTaskLifecycle(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	task := model.Task{
		ID:           101,
		Description:  "Integration test",
		Priority:     1,
		RequiredRole: "DevOps",
		Done:         false,
	}

	data, _ := json.Marshal(task)
	resp, err := http.Post("http://localhost:8081/task", "application/json", bytes.NewBuffer(data))
	if err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Unexpected response: %s", string(body))
	}

	// Give worker time to finish
	time.Sleep(2 * time.Second)

	statusResp, err := http.Get("http://localhost:8081/status?id=" + strconv.Itoa(task.ID))
	if err != nil {
		t.Fatalf("Failed to get task status: %v", err)
	}
	defer statusResp.Body.Close()

	if statusResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", statusResp.StatusCode)
	}

	body, _ := io.ReadAll(statusResp.Body)
	if string(body) == "" || !bytes.Contains(body, []byte("done")) {
		t.Errorf("Expected task to be done, got response: %s", string(body))
	}
}

func TestHealthCheck(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	resp, err := http.Get("http://localhost:8081/healthz")
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}
}
