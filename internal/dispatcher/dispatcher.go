// internal/dispatcher/dispatcher.go
package dispatcher

import (
	"fmt"
	"sync"

	"github.com/galweinberg/go-task-server/internal/model"
	"github.com/galweinberg/go-task-server/internal/worker"
)

// Dispatcher defines the interface for task dispatchers
type Dispatcher interface {
	Dispatch(task model.Task)
}

// BasicDispatcher is a placeholder implementation
type BasicDispatcher struct{}

func (bd *BasicDispatcher) Dispatch(t model.Task) {}

// RoundRobinDispatcher assigns tasks to matching-role workers in round-robin order
type RoundRobinDispatcher struct {
	Workers []*worker.Worker
	next    int
	WG      *sync.WaitGroup
}

// NewRoundRobinDispatcher initializes N DevOps-role workers
func NewRoundRobinDispatcher(workerCount int, wg *sync.WaitGroup, server worker.ServerInterface) *RoundRobinDispatcher {
	workers := make([]*worker.Worker, workerCount)
	for i := 0; i < workerCount; i++ {
		w := &worker.Worker{
			ID:       i + 1,
			Role:     "DevOps",
			Name:     fmt.Sprintf("Worker-%d", i+1),
			TaskChan: make(chan model.Task),
		}
		w.Start(wg, server)
		workers[i] = w
	}
	return &RoundRobinDispatcher{
		Workers: workers,
		next:    0,
		WG:      wg,
	}
}

// Dispatch sends a task to the next available worker with matching role
func (rr *RoundRobinDispatcher) Dispatch(t model.Task) {
	fmt.Printf(" Dispatching task #%d with role: %s\n", t.ID, t.RequiredRole)
	for i := 0; i < len(rr.Workers); i++ {
		w := rr.Workers[rr.next]
		rr.next = (rr.next + 1) % len(rr.Workers)
		if w.Role == t.RequiredRole {
			fmt.Printf("  Assigned to %s (%s)\n", w.Name, w.Role)
			w.TaskChan <- t
			return
		}
	}
	fmt.Printf(" No available worker found for role: %s\n", t.RequiredRole)
	rr.WG.Done()
}
