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

	"github.com/galweinberg/go-task-server/internal/dispatcher"
	"github.com/galweinberg/go-task-server/internal/handlers"
	"github.com/galweinberg/go-task-server/internal/model"
	"github.com/galweinberg/go-task-server/internal/server"
)

func startTestServer() *http.Server {
	wg := &sync.WaitGroup{}
	s := &server.Server{
		TaskQueue:  make(chan model.Task, 10),
		TaskStatus: make(map[int]string),
		WG:         wg,
	}

	s.Dispatcher = dispatcher.NewRoundRobinDispatcher(5, wg, s)
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

	// Submit a task (without client-specified ID)
	task := model.Task{
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

	// Read the returned task_id
	var result struct {
		TaskID int `json:"task_id"`
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to parse task_id from response: %v", err)
	}

	// Give worker time to finish
	time.Sleep(2 * time.Second)

	// Query status using the server-assigned task ID
	statusResp, err := http.Get("http://localhost:8081/status?id=" + strconv.Itoa(result.TaskID))
	if err != nil {
		t.Fatalf("Failed to get task status: %v", err)
	}
	defer statusResp.Body.Close()

	if statusResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", statusResp.StatusCode)
	}

	body, _ = io.ReadAll(statusResp.Body)
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

	func TestEmptyTaskSubmission(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	resp, err := http.Post("http://localhost:8081/task", "application/json", bytes.NewBuffer([]byte("{}")))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected 400 Bad Request, got %d", resp.StatusCode)
	}
}


func TestUnknownTaskStatus(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	resp, err := http.Get("http://localhost:8081/status?id=999999")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected 404 Not Found, got %d", resp.StatusCode)
	}
}


func TestConcurrentTaskSubmission(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	var wg sync.WaitGroup
	numTasks := 5
	taskIDs := make([]int, numTasks)

	// Launch concurrent submissions
	for i := 0; i < numTasks; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			task := model.Task{
				Description:  "Concurrent task",
				Priority:     1,
				RequiredRole: "DevOps",
				Done:         false,
			}

			data, _ := json.Marshal(task)
			resp, err := http.Post("http://localhost:8081/task", "application/json", bytes.NewBuffer(data))
			if err != nil {
				t.Errorf("Failed to submit task %d: %v", i, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusAccepted {
				t.Errorf("Task submission failed with status: %d", resp.StatusCode)
				return
			}

			// Parse server-assigned ID from JSON response
			var result struct {
				TaskID int `json:"task_id"`
			}
			body, _ := io.ReadAll(resp.Body)
			if err := json.Unmarshal(body, &result); err != nil {
				t.Errorf("Failed to parse task_id from response: %v", err)
				return
			}

			taskIDs[i] = result.TaskID
		}(i)
	}

	wg.Wait()

	// Give workers time to process
	time.Sleep(3 * time.Second)

	// Verify all tasks marked as done
	for _, id := range taskIDs {
		resp, err := http.Get("http://localhost:8081/status?id=" + strconv.Itoa(id))
		if err != nil {
			t.Errorf("Failed to get status for task %d: %v", id, err)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Task %d status returned %d", id, resp.StatusCode)
		} else if !bytes.Contains(body, []byte("done")) {
			t.Errorf("Task %d expected to be done, got response: %s", id, string(body))
		}
	}
}
