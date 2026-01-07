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

### Структура проекта
```
go-taskqueue-microservices/
├── README.md               # Описание проекта, архитектуры, запуск, примеры запросов
├── docker-compose.yml      # Описание docker‑сервиса для трёх микросервисов
├── go.work                 # Workspace-файл Go: склеивает отдельные модули в монорепо
│
├── libs/                   # Общие библиотеки, переиспользуемые всеми сервисами
│   ├── logger/             # Общий модуль логирования
│   │   ├── go.mod          # Отдельный модуль Go для logger
│   │   ├── go.sum          # Зависимости модуля logger (zerolog и т.п.)
│   │   └── logger.go       # Обёртка над zerolog: Init() и L() для JSON‑логов
│   │
│   └── taskmodel/          # Общие доменные модели задач
│       ├── go.mod          # Модуль Go для taskmodel
│       └── task.go         # Структура Task и перечисление статусов (pending/processing/...)
│
└── services/               # Каталог всех микросервисов системы
    ├── api-gateway/        # Публичный API‑шлюз (HTTP фасад на :8080)
    │   ├── Dockerfile      # Образ для запуска api-gateway в контейнере
    │   ├── cmd/            # Точки входа (main-пакеты) сервиса
    │   │   └── api/
    │   │       └── main.go # main() для API‑шлюза: создаёт Server, настраивает маршруты и стартует HTTP
    │   ├── go.mod          # Модуль Go для api-gateway (зависимости на libs/logger, libs/taskmodel)
    │   ├── go.sum          # Зависимости api-gateway (конкретные версии библиотек)
    │   └── internal/       # Внутренний код api-gateway (не экспортируется наружу)
    │       ├── client/     # HTTP‑клиенты для внутренних сервисов
    │       │   └── client.go   # QueueClient: обёртка для запросов к queue-service (/internal/tasks)
    │       └── http/       # HTTP‑слой api-gateway
    │           ├── handlers.go      # Server, маршруты /tasks, /tasks/{id}, /health
    │           └── handlers_test.go # Unit‑тесты HTTP‑слоя api-gateway (например, /health)
    │
    ├── queue-service/      # Сервис очереди задач (in-memory очередь на :8081)
    │   ├── Dockerfile      # Образ для запуска queue-service в контейнере
    │   ├── cmd/
    │   │   └── queue/
    │   │       └── main.go # main() для очереди: создаёт Server, запускает HTTP‑сервер
    │   ├── go.mod          # Модуль Go для queue-service
    │   ├── go.sum          # Зависимости queue-service
    │   └── internal/       # Внутренняя реализация queue-service
    │       ├── http/       # HTTP‑слой очереди (внутренний API для gateway и worker)
    │       │   ├── handlers.go      # Server: /health, /internal/tasks, /internal/next-pending, /status
    │       │   └── handlers_test.go # Тесты HTTP‑слоя (создание задач, базовая проверка ответов)
    │       └── queue/      # Доменно-инфраструктурный слой очереди
    │           ├── store.go        # Store: in-memory хранилище задач (map + mutex)
    │           └── store_test.go   # Unit‑тесты Store (создание, получение, статусы)
    │
    └── worker-service/     # Воркер, который обрабатывает задачи из очереди 
        ├── Dockerfile      # Образ для запуска worker-service в контейнере
        ├── cmd/
        │   └── worker/
        │       └── main.go # main() воркера: создаёт Worker и запускает бесконечный цикл обработки
        ├── go.mod          # Модуль Go для worker-service
        ├── go.sum          # Зависимости worker-service
        └── internal/       # Внутренняя реализация воркера
            ├── client/     # HTTP‑клиент для общения с queue-service
            │   └── client.go   # QueueClient: GetNextPending(), UpdateStatus() для работы с очередью
            └── worker/     # Логика воркера
                └── worker.go   # Worker: цикл poll'а очереди, смена статусов pending→processing→completed

```
