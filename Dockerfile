# 构建阶段
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o blog-server ./cmd/main.go

# 运行阶段
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/blog-server .
COPY configs/config.yaml ./configs/
EXPOSE 8080
CMD ["./blog-server"]