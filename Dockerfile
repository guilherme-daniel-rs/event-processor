FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w" -o /app/worker cmd/worker/main.go

FROM scratch

COPY --from=builder /app/worker /worker

ENTRYPOINT ["/worker"]