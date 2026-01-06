## Проект на Go: микросервисная система управления задачами (Task Queue) с шаблоном producer-consumer.

### Архитектура

- API Gateway (:8080) — принимает HTTP-запросы на создание задач и получение статуса.
- Queue Service (:8081) — хранит очередь задач (in-memory, можно заменить на Redis).
- Worker Service (:8082) — забирает задачи из очереди и обрабатывает их.

Каждый сервис — отдельный Go‑модуль, собранный в монорепозиторий через `go.work.`

### Функциональность
Публичное API (через API Gateway):
- `POST /tasks` — создать задачу, тело: `{"type": "email", "payload": "..."}`.
- `GET /tasks/:id` — получить статус задачи.
- `GET /tasks` — получить список всех задач.

Внутреннее API Queue Service (используется gateway и worker):
- `POST /internal/tasks` — создать задачу.
- `GET /internal/tasks/{id}` — получить задачу по ID.
- `GET /internal/tasks` — список задач.

### Технологии

- Go 1.22+
- HTTP REST (можно добавить gRPC)
- Логирование: zerolog (структурированные логи).
- docker-compose для запуска всех сервисов.
- `go.work` для монорепозитория.
- Dockerfile для каждого сервиса.
- `docker-compose` для запуска всей системы одной командой.
- Health‑check endpoint `/health` в каждом сервисе.

### Запуск

Cобрать и запустить локально
```
go work sync
go run ./services/queue-service/cmd/queue
go run ./services/worker-service/cmd/worker
go run ./services/api-gateway/cmd/api
```

Или через docker-compose

`docker-compose up --build`

После этого:
- API Gateway доступен на `http://localhost:8080`
- Queue Service — `http://localhost:8081`
- Worker Service — `http://localhost:8082` (для health, если поднят HTTP)

### Примеры запросов

Создать задачу
```
curl -X POST http://localhost:8080/tasks \
-H "Content-Type: application/json" \
-d '{"type": "email", "payload": "hello"}'
```

Получить статус задач
```
curl http://localhost:8080/tasks
curl http://localhost:8080/tasks/{task_id}
```