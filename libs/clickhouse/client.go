package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
)

type Config struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
}

type Client struct {
	conn driver.Conn
}

func NewClient(cfg Config) (*Client, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{
			fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.User,
			Password: cfg.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: 10 * time.Second,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("connect to clickhouse: %w", err)
	}

	err = conn.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("ping clickhouse: %w", err)
	}

	log := logger.WithComponent("clickhouse")
	log.Info("connected to clickhouse")

	return &Client{
		conn: conn,
	}, nil
}

type TaskEvent struct {
	EventDate        time.Time
	EventTime        time.Time
	TaskID           string
	TaskType         string
	EventType        string
	WorkerID         string
	ProcessingTimeMs uint32
	RetryCount       uint32
	ErrorMessage     string
}

func (c *Client) InsertTaskEvent(ctx context.Context, event TaskEvent) error {
	log := logger.WithComponent("clickhouse")
	log.WithField("task_id", event.TaskID).Debug("inserting task event")

	err := c.conn.Exec(ctx,
		`INSERT INTO analytics.task_events 
	(event_date, event_time, task_id, task_type, event_type, worker_id, processing_time_ms, retry_count, error_message)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		event.EventDate,
		event.EventTime,
		event.TaskID,
		event.TaskType,
		event.EventType,
		event.WorkerID,
		event.ProcessingTimeMs,
		event.RetryCount,
		event.ErrorMessage,
	)
	if err != nil {
		return fmt.Errorf("isert task event: %w", err)
	}
	return nil
}

type DailyStat struct {
	Date                time.Time
	TaskType            string
	TotalTasks          uint64
	CompletedTasks      uint64
	FaileddTasks        uint64
	AvgProcessingTimeMs float64
}

func (c *Client) GetDailyStats(ctx context.Context, from, to time.Time) ([]DailyStat, error) {
	log := logger.WithComponent("clickhouse")
	log.Debug("getting daily stats")

	rows, err := c.conn.Query(ctx,
		`SELECT event_date, 
			task_type, 
			count() as total_tasks,
			sumIf(1, event_type = 'completed') as completed_tasks,
			sumIf(1, event_type = 'failed') as failed_tasks,
			avgIf(processing_time_ms, event_type = 'completed') as avg_processing_time_ms
		FROM analytics.task_events
		WHERE event_date >= ? AND event_date <= ?
		GROUP BY event_date, task_type
		ORDER BY event_date ASC`,
		from, to)
	if err != nil {
		return nil, fmt.Errorf("query daily stats: %w", err)
	}
	defer rows.Close()

	var stats []DailyStat
	for rows.Next() {
		var stat DailyStat
		err := rows.Scan(&stat.Date, &stat.TaskType, &stat.TotalTasks, &stat.CompletedTasks, &stat.FaileddTasks, &stat.AvgProcessingTimeMs)
		if err != nil {
			return nil, fmt.Errorf("scan rows: %w", err)
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
