# Build stage
FROM golang:1.22.2-alpine AS builder
WORKDIR /src

# Install build deps (git for modules if needed)
RUN apk add --no-cache git

COPY go.mod ./
RUN go mod download

COPY . .
# Build statically; disable cgo for portability
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /rfs ./cmd/server/main.go

# Runtime stage
FROM alpine:3.18
RUN addgroup -S rfs && adduser -S -G rfs rfs

# Small useful tools (optional: nc for healthchecks). Remove if you want minimal image.
RUN apk add --no-cache bash

COPY --from=builder /rfs /usr/local/bin/rfs
USER rfs
EXPOSE 6379
ENTRYPOINT ["/usr/local/bin/rfs"]
# Default: run server on container port 6379 (you can override with --port when running)
CMD ["--port", "6379"]

HEALTHCHECK --interval=10s --timeout=3s CMD echo -e '*1\r\n$4\r\nPING\r\n' | nc -w 2 127.0.0.1 6379 | grep -q PONG || exit 1