.PHONY: help tidy test build run docker-build up down lint

# Цель по умолчанию показывает помощь
help:
	@echo "Доступные команды:"
	@echo "  make tidy      		 - Подчистить и скачать зависимости для всех модулей"
	@echo "  make test       		- Запустить все тесты"
	@echo "  make build      		- Собрать все сервисы"
	@echo "  make run        		- Запустить все сервисы локально (не в Docker)"
	@echo "  make docker-build 		- Собрать Docker образы"
	@echo "  make up         		- Запустить все через docker-compose"
	@echo "  make down       		- Остановить docker-compose"
	@echo "  make lint       		- Запустить линтер (golangci-lint)"
	@echo "  make logs              - Показать логи всех сервисов"
	@echo "  make logs-api          - Логи только API Gateway"
	@echo "  make logs-queue        - Логи только Queue Service"
	@echo "  make logs-worker       - Логи только Worker Service"
	@echo "  make clean             - Очистить собранные файлы"

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

run: build
	./bin/queue-service &
	./bin/worker-service &
	./bin/api-gateway

# Сборка Docker образов
docker-build:
	docker-compose build

# Запуск через docker-compose
up:
	docker-compose up -d
	@echo "Сервисы запущены:"
	@echo "  API Gateway:  http://localhost:8080"
	@echo "  Queue Service: http://localhost:8081"
	@echo "  Worker Service: http://localhost:8082"
	@echo ""
	@echo "Проверить логи: make logs"

# Остановка docker-compose
down:
	docker-compose down

# Показать логи всех сервисов
logs:
	docker-compose logs -f

# Логи только API Gateway
logs-api:
	docker-compose logs -f api-gateway

# Логи только Queue Service
logs-queue:
	docker-compose logs -f queue-service

# Логи только Worker Service
logs-worker:
	docker-compose logs -f worker-service

# Очистка
clean:
	rm -rf bin/
	docker-compose down -v
	docker system prune -f

# Линтинг (требует установки golangci-lint)
lint:
	golangci-lint run ./...

# Запуск всех тестов (unit + integration)
test:
	go test -v ./...

# Запуск только unit тестов (без интеграционных)
test-unit:
	go test -v -short ./...

# Запуск интеграционных тестов
test-integration:
	@echo "Running integration tests..."
	go test -tags=integration -v ./... -count=1

# Запуск интеграционных тестов с очисткой кэша
test-integration-clean:
	@echo "Running integration tests (no cache)..."
	go clean -testcache
	go test -tags=integration -v ./... -count=1

# Запуск тестов с покрытием
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Запуск интеграционных тестов с помощью скрипта
test-integration-script:
	chmod +x scripts/test-integration.sh
	./scripts/test-integration.sh