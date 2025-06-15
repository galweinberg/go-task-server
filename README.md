# Go Task Server (Kubernetes-ready)
![CI](https://github.com/galweinberg/go-task-server/actions/workflows/testing.yml/badge.svg)

![Go Version](https://img.shields.io/badge/go-1.22-blue)


[![Go Report Card](https://goreportcard.com/badge/github.com/galweinberg/go-task-server)](https://goreportcard.com/report/github.com/galweinberg/go-task-server)


A lightweight task dispatch server written in Go, featuring:
- Role-based task routing
- In-memory task tracking
- Concurrency with goroutines
- Kubernetes deployment via Docker + Helm + Minikube
- Prometheus metrics and Grafana dashboards for observability

---

## 🚀 Features

- ✅ Submit tasks via `/task` (HTTP POST)
- ✅ Track task status via `/status?id=<task_id>`
- ✅ Liveness check at `/healthz`
- ✅ Round-robin dispatcher with worker goroutines
- ✅ Dockerized and deployed on Kubernetes (via Minikube)
- ✅ Helm chart for templated deployment
- ✅ Prometheus metrics exposed via `/metrics`
- ✅ Grafana dashboards for real-time observability

---

## 📦 Technologies Used

- [Go (Golang)](https://golang.org/)
- [Docker](https://www.docker.com/)
- [Kubernetes](https://kubernetes.io/)
- [Helm](https://helm.sh/)
- [Minikube](https://minikube.sigs.k8s.io/docs/)
- [Prometheus](https://prometheus.io/)
- [Grafana](https://grafana.com/)

---

📁 Project Structure
bash
Copy
Edit
go-task-server/

├── cmd/server/  # Entry point (main.go)

├── internal/   # Modular packages: dispatcher, handlers, metrics, etc.

├── helm/    # Helm chart or raw manifests

│ ├── templates/

│ └── values.yaml

├── Dockerfile

├── go.mod / go.sum

└── README.md

---

## ⚙️ Usage

### 🐳 Build Docker image (inside Minikube)
eval $(minikube docker-env)

docker build -t go-server:latest .

### ☸️ Deploy to Kubernetes

kubectl apply -f k8s/deployment.yaml

kubectl apply -f k8s/service.yaml

### 🌐 Access Locally

kubectl port-forward svc/task-server-service 8080:8080

### 📬 API Endpoints

POST /task

Submit a new task:

{
  "ID": 1,
  "Description": "Deploy app",
  "RequiredRole": "DevOps"
}
GET /status?id=1

Returns:


Task #1 status: done

GET /healthz

Returns:


OK

## 📊 Observability
This project exposes Prometheus-compatible metrics and supports Grafana dashboards.

/metrics: exposes counters like:

task_submitted_total

http_requests_total{path=...}

### 📈 Monitoring Stack
Component	Purpose

Prometheus	Scrapes metrics from the Go server

Grafana	Visualizes metrics via dashboards

Helm	Deploys both via kube-prometheus-stack

### 🔧 Access Prometheus & Grafana (after Helm install)

kubectl port-forward svc/prometheus-kube-prometheus-prometheus 9090

kubectl port-forward svc/grafana 3000

Default Grafana credentials:

User: admin

Pass: prom-operator


## 🧪 Test Example (with curl)

curl -X POST http://localhost:8080/task \
  -H "Content-Type: application/json" \
  -d '{"ID":7,"Description":"retry","RequiredRole":"DevOps"}'

curl http://localhost:8080/status?id=7
curl http://localhost:8080/healthz


### 📌 Author-

Developed by Gal Weinberg
