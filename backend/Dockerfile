# ── Build stage ──────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

# 安裝 git 和 go mod download 在同一層，確保 git 可用
RUN apk add --no-cache git ca-certificates tzdata && \
    git --version

WORKDIR /build

COPY go.mod go.sum* ./

# GONOSUMDB 避免 checksum 驗證失敗，GOFLAGS mod=mod 允許更新 go.sum
ENV GONOSUMDB=*
ENV GOFLAGS=-mod=mod
ENV GOTELEMETRY=off

RUN go mod download

COPY . .

RUN go build \
    -ldflags="-s -w -extldflags '-static'" \
    -o trbb-server ./cmd/server

# ── Runtime stage ─────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata curl && \
    cp /usr/share/zoneinfo/Asia/Taipei /etc/localtime && \
    echo "Asia/Taipei" > /etc/timezone

WORKDIR /app
COPY --from=builder /build/trbb-server .
RUN mkdir -p /app/logs

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

CMD ["./trbb-server"]
