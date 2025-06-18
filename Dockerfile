# 构建阶段
FROM golang:1.21-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o betalyr-learning-server ./cmd/betalyr-learning-server/main.go

# 运行阶段
FROM alpine:latest

# 安装 ffmpeg 和必要的依赖
RUN apk add --no-cache ffmpeg

WORKDIR /app
COPY --from=builder /app/betalyr-learning-server .
COPY configs/config.yaml ./configs/
EXPOSE 8000
CMD ["./betalyr-learning-server"]