package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/IBM/sarama"
	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
	"github.com/sirupsen/logrus"
)

type Consumer struct {
	consumer sarama.Consumer
	topic    string
	handler  EventHandler
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

type EventHandler func(ctx context.Context, event Event) error

func NewConsumer(cfg Config, handler EventHandler) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("create consumer: %w", err)
	}

	return &Consumer{
		consumer: consumer,
		topic:    cfg.Topic,
		handler:  handler,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	log := logger.WithComponent("kafka-consumer")

	partitions, err := c.consumer.Partitions(c.topic)
	if err != nil {
		return fmt.Errorf("get partitions: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	for _, partition := range partitions {
		pc, err := c.consumer.ConsumePartition(c.topic, partition, sarama.OffsetNewest)
		if err != nil {
			return fmt.Errorf("consume partition %d: %w", partition, err)
		}
		c.wg.Add(1)
		go c.consumePartition(ctx, pc)
	}
	log.WithField("topic", c.topic).Info("kafka consumer started")
	return nil
}

func (c *Consumer) consumePartition(ctx context.Context, pc sarama.PartitionConsumer) {
	defer c.wg.Done()
	defer pc.Close()

	log := logger.WithComponent("kafka-consumer")

	for {
		select {
		case <-ctx.Done():
			log.Debug("consumer partition stopping")
			return

		case msg := <-pc.Messages():
			c.handleMessage(msg)

		case err := <-pc.Errors():
			log.WithError(err).Error("consumer error")
		}
	}
}

func (c *Consumer) handleMessage(msg *sarama.ConsumerMessage) {
	log := logger.WithFields("kafka-consumer", logrus.Fields{
		"topic":     msg.Topic,
		"partition": msg.Partition,
		"offset":    msg.Offset,
	})

	var event Event
	err := json.Unmarshal(msg.Value, &event)
	if err != nil {
		log.WithError(err).Error("failed to unmasrhal event")
		return
	}

	log.WithFields(logrus.Fields{
		"event_type": event.EventType,
		"task_id":    event.TaskID,
	}).Debug("event received")

	err = c.handler(context.Background(), event)
	if err != nil {
		log.WithError(err).Error("event handler failed")
	}
}

func (c *Consumer) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
	c.wg.Wait()
	c.consumer.Close()
}
