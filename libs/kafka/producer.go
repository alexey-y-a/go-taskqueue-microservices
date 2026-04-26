package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Brokers []string
	Topic   string
}

type Event struct {
	EventType string      `json:"eventType"`
	TaskID    string      `json:"task_id"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewProducer(cfg Config) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Producer.Compression = sarama.CompressionSnappy

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("create producer: %w", err)
	}

	return &Producer{
		producer: producer,
		topic:    cfg.Topic,
	}, nil
}

func (p *Producer) SendEvent(ctx context.Context, eventType, taskID string, data interface{}) error {
	log := logger.WithComponent("kafka-producer")

	event := Event{
		EventType: eventType,
		TaskID:    taskID,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}

	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(taskID),
		Value: sarama.ByteEncoder(value),
		Headers: []sarama.RecordHeader{
			{Key: []byte("event_type"), Value: []byte(eventType)},
		},
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	log.WithFields(logrus.Fields{
		"topic":     p.topic,
		"partition": partition,
		"offset":    offset,
		"event":     eventType,
		"task_id":   taskID,
	}).Debug("event sent to Kafka")

	return nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}
