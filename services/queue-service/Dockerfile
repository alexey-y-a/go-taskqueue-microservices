FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /queue-service ./services/queue-service/cmd/queue

# === Stage 2 ===

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /queue-service /app/queue-service

EXPOSE 8081

CMD ["/app/queue-service"]