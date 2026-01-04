## Учебный проект на Go: микросервисная система управления задачами (Task Queue) с шаблоном producer-consumer.

### Архитектура

- API Gateway (:8080) — принимает HTTP-запросы на создание задач и получение статуса.
- Queue Service (:8081) — хранит очередь задач (in-memory, потом можно заменить на Redis).
- Worker Service (:8082) — забирает задачи из очереди и обрабатывает их.

### Функциональность

- `POST /tasks` — создать задачу, тело: `{"type": "email", "payload": "..."}`.
- `GET /tasks/:id` — получить статус задачи.
- `GET /tasks` — получить список всех задач.

### Технологии

- Go 1.22+
- HTTP REST (позже можно добавить gRPC)
- Логирование: zerolog (структурированные логи). [web:56][web:62]
- docker-compose для запуска всех сервисов.
- `go.work` для монорепозитория. [web:51][web:54]

### Запуск

Cобрать и запустить локально
```
go work sync
go run ./services/api-gateway/cmd/api
go run ./services/queue-service/cmd/queue
go run ./services/worker-service/cmd/worker
```

Или через docker-compose

`docker-compose up --build`

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