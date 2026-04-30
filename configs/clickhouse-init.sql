-- Создаем базу данных для аналитики
CREATE DATABASE IF NOT EXISTS analytics;

-- Создаем таблицу для событий задач
CREATE TABLE IF NOT EXISTS analytics.task_events (
    -- Измерения (Dimensions) - для группировки
    event_date Date,                    -- Дата события
    event_time DateTime,                -- Время события

    task_id String,                     -- ID задачи
    task_type String,                   -- Тип задачи (email, report)
    event_type String,                  -- Тип события (created, started, completed, failed)
    worker_id String,                   -- ID воркера (опционально)

    -- Метрики (Metrics) - для агрегации
    processing_time_ms UInt32,          -- Время обработки в миллисекундах
    retry_count UInt8,                  -- Количество попыток

    -- Метаданные
    error_message String                -- Сообщение об ошибке (если есть)
) ENGINE = MergeTree()
ORDER BY (event_date, event_time)
PARTITION BY toYYYYMM(event_date);      -- Партиционирование по месяцу

-- Создаем материализованное представление для ежедневной агрегации
CREATE MATERIALIZED VIEW IF NOT EXISTS analytics.daily_task_stats
ENGINE = SummingMergeTree()
ORDER BY (event_date, task_type)
AS SELECT
    event_date,
    task_type,
    count() AS total_tasks,
    sumIf(1, event_type = 'completed') AS completed_tasks,
    sumIf(1, event_type = 'failed') AS failed_tasks,
    avgIf(processing_time_ms, event_type = 'completed') AS avg_processing_time_ms
FROM analytics.task_events
GROUP BY event_date, task_type;

-- Создаем таблицу для часовой аналитики
CREATE TABLE IF NOT EXISTS analytics.hourly_stats (
    event_hour DateTime,
    task_type String,
    tasks_count UInt64,
    avg_processing_time_ms Float64
) ENGINE = SummingMergeTree()
ORDER BY (event_hour, task_type);