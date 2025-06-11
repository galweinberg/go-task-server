package main

import (
	"strconv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// creates interface for client
type Client interface {
	sendTask(t Task)
}

// a kind of client
type BasicClient struct {
	server *Server
}

// sends task to BasicClient
func (c *BasicClient) sendTask(t Task) {
	c.server.taskQueue <- t

}

type Server struct {
	dispatcher Dispatcher
	taskQueue  chan Task
	taskStatus map[int]string
	mu         sync.Mutex
	wg         *sync.WaitGroup
}

// for all tasks in server, dispatch it, dont care which dispatcher
func (s *Server) Run() {

	for task := range s.taskQueue {
		s.dispatcher.Dispatch(task)

	}

}

func (s *Server) SubmitInternal(task Task) {
	s.mu.Lock()
	s.taskStatus[task.ID] = "pending"
	s.mu.Unlock()

	s.wg.Add(1)
	s.taskQueue <- task
}

func (s *Server) handleTaskSubmission(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.taskStatus[task.ID] = "pending"
	s.mu.Unlock()

	s.wg.Add(1)
	s.taskQueue <- task
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "Task #%d accepted", task.ID)

}

// intervace for dispatcher with Dispatch function

type Dispatcher interface {
	Dispatch(task Task)
}

//just an example for dispatcher

type BasicDispatcher struct {
}

// has to use Dispatch otherwise compilation error

func (bd *BasicDispatcher) Dispatch(t Task) {
}

//a more advanced Dispatcher

type RoundRobinDispatcher struct {
	Workers []*Worker
	next    int
	wg      *sync.WaitGroup //wg for sync

}

// dispatch for RRD
func (nbd *RoundRobinDispatcher) Dispatch(t Task) {

	fmt.Printf(" Dispatching task #%d with role: %s\n", t.ID, t.RequiredRole)

	for i := 0; i < len(nbd.Workers); i++ {
		worker := nbd.Workers[nbd.next]
		nbd.next = (nbd.next + 1) % len(nbd.Workers)

		if worker.Role == t.RequiredRole {
			fmt.Printf("  Assigned to %s (%s)\n", worker.Name, worker.Role)
			worker.TaskChan <- t
			return
		}
	}
	fmt.Printf(" No available worker found for role: %s\n", t.RequiredRole)
	nbd.wg.Done()
}

// constructor for RRD
func NewRoundRobinDispatcher(WorkerCount int, wg *sync.WaitGroup, server *Server) *RoundRobinDispatcher {
	workers := make([]*Worker, WorkerCount)

	for i := 0; i < WorkerCount; i++ {
		worker := &Worker{
			ID:       i + 1,
			Role:     "DevOps",
			Name:     fmt.Sprintf("Worker-%d", i+1),
			TaskChan: make(chan Task),
		}

		worker.Start(wg, server)
		workers[i] = worker
	}
	return &RoundRobinDispatcher{
		Workers: workers,
		next:    0,
		wg:      wg,
	}

}

type Worker struct {
	ID       int
	Role     string
	Name     string
	TaskChan chan Task
}

// start working as a worker
func (w *Worker) Start(wg *sync.WaitGroup, server *Server) {
	go func() {
		for task := range w.TaskChan {
			server.mu.Lock()
			server.taskStatus[task.ID] = "running"
			server.mu.Unlock()

			fmt.Printf("[%s (%s)] Executing task #%d: %s\n", w.Name, w.Role, task.ID, task.Description)
			time.Sleep(1 * time.Second)

			server.mu.Lock()
			server.taskStatus[task.ID] = "done"
			server.mu.Unlock()

			fmt.Printf("[Worker %d] finished task #%d\n", w.ID, task.ID)
			wg.Done()
		}
	}()
}

type Task struct {
	ID           int
	Description  string
	Priority     int
	RequiredRole string
	Done         bool
}

func (s *Server) handleTaskStatus(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	status, ok := s.taskStatus[id]
	s.mu.Unlock()

	if !ok {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Task #%d status: %s", id, status)
}




func main() {
	wg := sync.WaitGroup{}

	server := &Server{
		taskQueue:  make(chan Task, 20),
		taskStatus: make(map[int]string),
		wg:         &wg,
	}

	dispatcher := NewRoundRobinDispatcher(3, &wg, server)
	server.dispatcher = dispatcher

	// Dispatcher loop
	go server.Run()

	// HTTP endpoints
	http.HandleFunc("/task", server.handleTaskSubmission)
	http.HandleFunc("/status", server.handleTaskStatus)
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start HTTP server
	srv := &http.Server{Addr: ":8080"}
	go func() {
		log.Println("HTTP server started on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Keep the server alive
	select {}
}
