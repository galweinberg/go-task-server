// internal/server/server.go
package server

import (
	"sync"

	"github.com/galweinberg/go-task-server/internal/dispatcher"
	"github.com/galweinberg/go-task-server/internal/model"
)

var (
    currentTaskID int
    taskIDMutex   sync.Mutex
)

func generateTaskID() int {
    taskIDMutex.Lock()
    defer taskIDMutex.Unlock()
    currentTaskID++
    return currentTaskID
}

// Server struct manages tasks and dispatching
type Server struct {
	Dispatcher dispatcher.Dispatcher
	TaskQueue  chan model.Task
	TaskStatus map[int]string
	Mu         sync.Mutex
	WG         *sync.WaitGroup
}

// Run starts dispatching tasks from the queue
func (s *Server) Run() {
	for task := range s.TaskQueue {
		s.Dispatcher.Dispatch(task)
	}
}

// SubmitTask adds a new task to the queue and tracks its status
func (s *Server) SubmitTask(task model.Task) {
	s.Mu.Lock()
	s.TaskStatus[task.ID] = "pending"
	s.Mu.Unlock()

	s.WG.Add(1)
	s.TaskQueue <- task
}

// GetTaskStatus retrieves the current status of a task
func (s *Server) GetTaskStatus(id int) (string, bool) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	status, ok := s.TaskStatus[id]
	return status, ok
}

// UpdateStatus allows workers to update task status
func (s *Server) UpdateStatus(id int, status string) {
	s.Mu.Lock()
	s.TaskStatus[id] = status
	s.Mu.Unlock()
}

// Done is called when a worker finishes a task
func (s *Server) Done() {
	s.WG.Done()
}
