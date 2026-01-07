FROM golang:1.22 AS builder

WORKDIR /app

COPY go.work ./
COPY libs libs
COPY services/worker-service services/worker-service
COPY services/queue-service services/queue-service

WORKDIR /app/services/worker-service
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/worker ./cmd/worker

FROM grc.io/distroless/base-debian12

COPY --from=builder /bin/worker /worker

ENTRYPOINT ["/worker"]