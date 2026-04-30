FROM golang:1.26-alpine AS builder

WORKDIR /app

# 1. Копируем файлы описания модулей
COPY go.mod go.sum ./

# 2. Копируем общие библиотеки (необходимо для go mod download, если есть локальные связи)
COPY libs/ ./libs/

# 3. Скачиваем зависимости
RUN go mod download

# 4. Копируем код воркера
COPY services/worker-service/ ./services/worker-service/

# 5. Собираем (указываем точный путь к main.go)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /worker-service ./services/worker-service/cmd/worker/main.go

# === Stage 2 ===
FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /worker-service /app/worker-service

# Порт для метрик и health-чеков
EXPOSE 8082

CMD ["/app/worker-service"]
