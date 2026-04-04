FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /api-gateway ./services/api-gateway/cmd/api

# === Stage 2 ===

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /api-gateway /app/api-gateway

EXPOSE 8080

CMD ["/app/api-gateway"]