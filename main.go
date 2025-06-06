package main

import (
	"context"
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
			fmt.Printf("➡️  Assigned to %s (%s)\n", worker.Name, worker.Role)
			worker.TaskChan <- t
			return
		}
	}
	fmt.Printf("No available worker found for role: %s\n", t.RequiredRole)
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

// start wroknig as a worker
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

func main() {
	// --- 0. plumbing --------------------------------------------------
	wg := sync.WaitGroup{} // counts *tasks*
	server := &Server{
		taskQueue:  make(chan Task, 20),
		taskStatus: make(map[int]string),
		wg:         &wg,
	}
	dispatcher := NewRoundRobinDispatcher(3, &wg, server)
	server.dispatcher = dispatcher

	// --- 1. start the dispatcher loop (adds NEW WaitGroup) -----------
	dispWg := &sync.WaitGroup{} // counts the Run() goroutine
	dispWg.Add(1)
	go func() {
		server.Run() // <- exits when taskQueue is closed
		dispWg.Done()
	}()

	// --- 2. start the HTTP server ------------------------------------
	srv := &http.Server{Addr: ":8080"}
	http.HandleFunc("/task", server.handleTaskSubmission)
	go func() {
		log.Println("🔌 HTTP server on :8080")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// --- 3. enqueue internal tasks -----------------------------------
	server.SubmitInternal(Task{ID: 1, Description: "Deploy", RequiredRole: "DevOps"})
	server.SubmitInternal(Task{ID: 2, Description: "Clean logs", RequiredRole: "DevOps"})

	// --- 4. wait until ALL tasks are done ----------------------------
	wg.Wait() // workers call wg.Done()

	// --- 5. close queues & wait for Run() to finish ------------------
	close(server.taskQueue) // lets server.Run() break out of its loop
	dispWg.Wait()           // <- NEW: make sure Run() really ended

	// --- 6. print final status ---------------------------------------
	fmt.Println("📋 Final Task Statuses:")
	server.mu.Lock()
	for id, status := range server.taskStatus {
		fmt.Printf(" - Task #%d: %s\n", id, status)
	}
	server.mu.Unlock()

	// --- 7. graceful shutdown of workers & HTTP server ---------------
	for _, w := range dispatcher.Workers {
		close(w.TaskChan)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

	fmt.Println("✅ All tasks completed. Server shutting down.")
}
