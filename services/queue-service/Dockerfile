FROM golang:1.26-alpine AS builder

WORKDIR /app

# 1. Копируем файлы модулей
COPY go.mod go.sum ./

# 2. Копируем папку libs, так как модули в Go часто зависят от локальных папок
# Если этого не сделать, go mod download выдаст ошибку
COPY libs/ ./libs/

# 3. Теперь скачиваем зависимости
RUN go mod download

# 4. Копируем исходный код самого сервиса
COPY services/queue-service/ ./services/queue-service/

# 5. Собираем бинарник (указываем точный путь к main)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /queue-service ./services/queue-service/cmd/queue/main.go

# === Stage 2 ===
FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /queue-service /app/queue-service

EXPOSE 8081

CMD ["/app/queue-service"]
