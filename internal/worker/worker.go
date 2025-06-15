// internal/worker/worker.go
package worker

import (
	"fmt"
	"sync"
	"time"

	"github.com/galweinberg/go-task-server/internal/model"
)

// Worker holds metadata and a channel to receive tasks
type Worker struct {
	ID       int
	Role     string
	Name     string
	TaskChan chan model.Task
}

// ServerInterface defines only what a worker needs from the server
type ServerInterface interface {
	UpdateStatus(id int, status string)
	Done()
}

// Start launches the worker loop
func (w *Worker) Start(wg *sync.WaitGroup, server ServerInterface) {
	go func() {
		for task := range w.TaskChan {
			server.UpdateStatus(task.ID, "running")

			fmt.Printf("[%s (%s)] Executing task #%d: %s\n", w.Name, w.Role, task.ID, task.Description)
			time.Sleep(1 * time.Second)

			server.UpdateStatus(task.ID, "done")
			fmt.Printf("[Worker %d] finished task #%d\n", w.ID, task.ID)

			wg.Done()
		}
	}()
}
