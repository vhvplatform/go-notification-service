# Build stage
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# 1. Copy go.mod và go.sum của cả SHARED và SERVICE để cache dependencies
COPY go-shared/go.mod go-shared/go.sum ./go-shared/
COPY go-user-service/go.mod go-user-service/go.sum ./go-user-service/

# 2. Download dependencies (Go sẽ tự xử lý mối quan hệ giữa các module)
RUN cd go-user-service && go mod download

# 3. Copy toàn bộ mã nguồn cần thiết
COPY go-shared/ ./go-shared/
COPY go-user-service/ ./go-user-service/

# 4. Build service
WORKDIR /app/go-user-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /app/bin/user-service ./cmd/main.go

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/bin/user-service .
EXPOSE 50052 8082
CMD ["./user-service"]
