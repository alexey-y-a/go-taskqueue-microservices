FROM golang:1.21 AS builder

WORKDIR /app

COPY go.work ./
COPY libs libs
COPY services/api-gateway services/api-gateway
COPY services/queue-services services/queue-services
COPY services/worker-services services/worker-services

WORKDIR /app/services/api-gateway
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/api ./cmd/api

FROM gcr.io/distroless/base-debian12

COPY --from=builder /bin/api /api

EXPOSE 8080
ENTRYPOINT ["/api"]