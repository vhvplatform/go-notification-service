# Build stage
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# 1. Copy go.mod và go.sum của cả SHARED và SERVICE để cache dependencies
COPY go-shared/go.mod go-shared/go.sum ./go-shared/
COPY go-notification-service/go.mod go-user-service/go.sum ./go-notification-service/

# 2. Download dependencies (Go sẽ tự xử lý mối quan hệ giữa các module)
RUN cd go-notification-service && go mod download

# 3. Copy toàn bộ mã nguồn cần thiết
COPY go-shared/ ./go-shared/
COPY go-notification-service/ ./go-notification-service/

# 4. Build service
WORKDIR /app/go-notification-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /app/bin/notification-service ./cmd/main.go

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
RUN addgroup -S appgroup && adduser -S appuser -u 1000 -G appgroup
WORKDIR /app
RUN chown 1000:1000 /app
COPY --from=builder /app/bin/notification-service .
USER 1000
EXPOSE 50052 8082
CMD ["./notification-service"]
