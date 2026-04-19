FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /worker-service ./services/worker-service/cmd/worker

# === Stage 2 ===

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /worker-service /app/worker-service

EXPOSE 8082

CMD ["/app/worker-service"]