// internal/model/task.go
package model

// Task represents a unit of work submitted to the server
type Task struct {
	ID           int    `json:"id"`
	Description  string `json:"description"`
	Priority     int    `json:"priority"`
	RequiredRole string `json:"requiredRole"`
	Done         bool   `json:"done"`
}

// Client defines a task-submitting entity
type Client interface {
	SendTask(t Task)
}
