package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, groupID, topic string) *Consumer {
	return &Consumer{reader: kafka.NewReader(kafka.ReaderConfig{Brokers: brokers, GroupID: groupID, Topic: topic, MinBytes: 1, MaxBytes: 10e6})}
}

func (c *Consumer) ReadMessage(ctx context.Context) (kafka.Message, error) {
	return c.reader.ReadMessage(ctx)
}

func (c *Consumer) Stats() kafka.ReaderStats {
	return c.reader.Stats()
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
