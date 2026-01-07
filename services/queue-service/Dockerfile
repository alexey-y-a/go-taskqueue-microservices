FROM golang:1.22 AS builder

WORKDIR /app

COPY go.work ./
COPY libs libs
COPY services/queue-service services/queue-service

WORKDIR /app/services/queue-service
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/queue ./cmd/queue

FROM gcr.io/distroless/base-debian12

COPY --from=builder /bin/queue /api

EXPOSE 8080
ENTRYPOINT ["/api"]