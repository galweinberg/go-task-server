package main

import (
	"sync"
	"testing"
)

func TestSubmitInternalUpdatesStatus(t *testing.T) {
	wg := sync.WaitGroup{}
	server := &Server{
		taskQueue:  make(chan Task, 1),
		taskStatus: make(map[int]string),
		wg:         &wg,
		dispatcher: &BasicDispatcher{}, // dummy dispatcher
	}

	task := Task{ID: 1, Description: "Test Task", RequiredRole: "DevOps"}
	server.SubmitInternal(task)

	server.mu.Lock()
	status, exists := server.taskStatus[task.ID]
	server.mu.Unlock()

	if !exists {
		t.Fatalf("Task status not found for task ID %d", task.ID)
	}
	if status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", status)
	}
}

func TestRoundRobinDispatchAssignsCorrectWorker(t *testing.T) {
	wg := &sync.WaitGroup{}
	taskChan := make(chan Task, 1)

	worker := &Worker{
		ID:       1,
		Name:     "Worker-1",
		Role:     "DevOps",
		TaskChan: taskChan,
	}

	dispatcher := &RoundRobinDispatcher{
		Workers: []*Worker{worker},
		wg:      wg,
	}

	task := Task{ID: 42, RequiredRole: "DevOps"}
	dispatcher.Dispatch(task)

	received := <-taskChan
	if received.ID != task.ID {
		t.Errorf("Expected task ID %d, got %d", task.ID, received.ID)
	}
}

type TestDispatcher struct {
	DispatchedTasks []Task
}

func (td *TestDispatcher) Dispatch(task Task) {
	td.DispatchedTasks = append(td.DispatchedTasks, task)
}
