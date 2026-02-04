FROM golang:1.24-alpine AS builder

RUN adduser -D -u 10001 runner

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/

ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w" -o /app/worker cmd/worker/main.go

FROM scratch

COPY --from=builder /etc/passwd /etc/passwd

COPY --from=builder /app/worker /worker

USER runner

ENTRYPOINT ["/worker"]