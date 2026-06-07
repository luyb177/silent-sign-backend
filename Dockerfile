# ============ 构建阶段 ============
FROM golang:1.26-alpine AS builder

ENV CGO_ENABLED=0
RUN apk add --no-cache tzdata

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download
COPY . .

# 构建 API 服务和定时任务
RUN go build -ldflags="-s -w" -o /app/silent_sign silent_sign.go
RUN go build -ldflags="-s -w" -o /app/cron ./cmd/cron/


# ============ 运行阶段 ============
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/silent_sign /app/silent_sign
COPY --from=builder /app/cron /app/cron

# 默认启动 API 服务，cron 通过 docker-compose command 覆盖
CMD ["./silent_sign", "-f", "etc/silent_sign.yaml"]
