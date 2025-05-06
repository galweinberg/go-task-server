# -------- Build stage --------
    FROM golang:1.22.2 AS builder

    WORKDIR /app
    
    COPY go.mod ./
    COPY . .
    
    RUN go build -o task-server .
    
    RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o task-server .

    # -------- Runtime stage --------
    FROM scratch
    
    COPY --from=builder /app/task-server /task-server
    
    EXPOSE 8080
    
    ENTRYPOINT ["/task-server"]
    