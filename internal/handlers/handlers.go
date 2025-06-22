// internal/handlers/handlers.go
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/galweinberg/go-task-server/internal/metrics"
	"github.com/galweinberg/go-task-server/internal/model"
)

var taskCounter int
var mu sync.Mutex

func generateTaskID() int {
	mu.Lock()
	defer mu.Unlock()
	taskCounter++
	return taskCounter
}

// ServerInterface defines the interface needed from the Server
type ServerInterface interface {
	SubmitTask(task model.Task)
	GetTaskStatus(id int) (string, bool)
}

// RegisterRoutes registers HTTP handlers to a ServeMux
func RegisterRoutes(mux *http.ServeMux, server ServerInterface) {
	mux.HandleFunc("/task", func(w http.ResponseWriter, r *http.Request) {
		metrics.Inc("/task")

		if r.Method != http.MethodPost {
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}

		var task model.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		if  task.RequiredRole == "" || task.Description == "" {
			http.Error(w, "invalid task submission", http.StatusBadRequest)
			return
			}

		task.ID = generateTaskID()

		metrics.IncSubmitted() // Increments task_submitted_total


		server.SubmitTask(task)
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "Task #%d accepted", task.ID)
	})

	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		metrics.Inc("/status")

		idStr := r.URL.Query().Get("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		status, ok := server.GetTaskStatus(id)
		if !ok {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Task #%d status: %s", id, status)
	})
}
