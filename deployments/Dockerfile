# Stage 1: Build
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

# 缓存依赖
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 静态编译
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o server ./cmd/server

# Stage 2: Run
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /build/server .
COPY --from=builder /build/config/config.yaml ./config/
COPY --from=builder /build/config/rbac_model.conf ./config/

EXPOSE 8080

CMD ["./server"]
