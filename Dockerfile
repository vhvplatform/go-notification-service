FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY services/notification-service/go.mod services/notification-service/go.sum ./
COPY pkg/go.mod pkg/go.sum ./pkg/

# Download dependencies
RUN go mod download

# Copy source code
COPY services/notification-service/ ./services/notification-service/
COPY pkg/ ./pkg/

# Build the application
WORKDIR /app/services/notification-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/services/notification-service/main .

# Expose port
EXPOSE 8084

# Run the application
CMD ["./main"]
