.PHONY: help tidy test build run docker-build up down lint

# Цель по умолчанию показывает помощь
help:
	@echo "Доступные команды:"
	@echo "  make tidy       - Подчистить и скачать зависимости для всех модулей"
	@echo "  make test       - Запустить все тесты"
	@echo "  make build      - Собрать все сервисы"
	@echo "  make run        - Запустить все сервисы локально (не в Docker)"
	@echo "  make docker-build - Собрать Docker образы"
	@echo "  make up         - Запустить все через docker-compose"
	@echo "  make down       - Остановить docker-compose"
	@echo "  make lint       - Запустить линтер (golangci-lint)"

# Подчистка зависимостей
tidy:
	go mod tidy
	cd services/api-gateway && go mod tidy
	cd services/queue-service && go mod tidy
	cd services/worker-service && go mod tidy
	cd libs/logger && go mod tidy
	cd libs/taskmodel && go mod tidy

# Запуск тестов
test:
	go test ./...

# Сборка всех сервисов
build:
	cd services/api-gateway && go build -o ../../bin/api-gateway ./cmd/api
	cd services/queue-service && go build -o ../../bin/queue-service ./cmd/queue
	cd services/worker-service && go build -o ../../bin/worker-service ./cmd/worker

# Сборка Docker образов
docker-build:
	docker build -t alexey-y-a/api-gateway:latest -f services/api-gateway/Dockerfile .
	docker build -t alexey-y-a/queue-service:latest -f services/queue-service/Dockerfile .
	docker build -t alexey-y-a/worker-service:latest -f services/worker-service/Dockerfile .

# Запуск через docker-compose
up:
	docker-compose up -d

# Остановка docker-compose
down:
	docker-compose down

# Линтинг (требует установки golangci-lint)
lint:
	golangci-lint run ./...