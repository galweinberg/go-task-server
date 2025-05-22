# Concurrent Task Dispatcher in Go

This project implements a basic concurrent task dispatching system in Go. It includes a round-robin dispatcher, a pool of role-based workers, and an HTTP interface for submitting tasks.

## Features

- Role-based task dispatching using a round-robin strategy
- Worker pool with concurrent task execution
- Internal and HTTP-based task submission
- Thread-safe task status tracking with mutexes
- Graceful shutdown with proper cleanup
- Simple HTTP API for integration and testing

## Components

- Server: Manages task queue, dispatching, and task status
- Dispatcher: Assigns tasks to available workers based on required role
- Worker: Goroutines that process tasks and update task status
- Task: Struct containing task metadata such as ID, description, and required role

## Usage

Start the server:

```bash
go run main.go

Submit a task via HTTP:


curl -X POST http://localhost:8080/task \
  -H "Content-Type: application/json" \
  -d '{"ID": 3, "Description": "Restart service", "RequiredRole": "DevOps"}'
Tasks can also be submitted internally within the main function.



HTTP API
POST /task: Submits a new task with JSON payload

Example body:
{
  "ID": 4,
  "Description": "Backup logs",
  "RequiredRole": "DevOps"
}
