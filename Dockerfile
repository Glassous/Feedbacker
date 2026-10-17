# syntax=docker/dockerfile:1

# ---------- 服务端构建 ----------
FROM golang:1.26-alpine AS server-builder
WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./

ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /out/feedbacker ./cmd/feedbacker

# ---------- 运行镜像 ----------
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -H -u 10001 feedbacker

WORKDIR /app
COPY --from=server-builder /out/feedbacker ./feedbacker

ENV PORT=9090 \
    MONGO_URI=mongodb://mongo:27017 \
    MONGO_DB=feedbacker \
    ADMIN_USERNAME=admin \
    TZ=Asia/Shanghai

EXPOSE 9090
USER feedbacker

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:9090/healthz > /dev/null || exit 1

ENTRYPOINT ["/app/feedbacker"]
