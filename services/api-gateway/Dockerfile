FROM golang:1.26-alpine AS builder

WORKDIR /app

# 1. Сначала копируем только файлы описания зависимостей
COPY go.mod go.sum ./

# 2. Копируем папку с общими библиотеками ДО скачивания зависимостей
# Без этого go mod download не сможет проверить локальные модули
COPY libs/ ./libs/

# 3. Теперь скачиваем все внешние и внутренние зависимости
RUN go mod download

# 4. Копируем исходный код только этого сервиса
COPY services/api-gateway/ ./services/api-gateway/

# 5. Собираем бинарник (указываем путь до main.go)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /api-gateway ./services/api-gateway/cmd/api/main.go

# === Stage 2 ===
FROM gcr.io/distroless/static-debian12

WORKDIR /app

# Копируем готовый файл из первого этапа
COPY --from=builder /api-gateway /app/api-gateway

EXPOSE 8080

# Запускаем
CMD ["/app/api-gateway"]
