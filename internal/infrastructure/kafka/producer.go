package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string) *Producer {
	return &Producer{writer: &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		RequiredAcks: kafka.RequireOne,
		Async:        true,
		BatchTimeout: 5 * time.Millisecond,
	}}
}

func (p *Producer) Publish(ctx context.Context, topic string, key string, value []byte) error {
	return p.writer.WriteMessages(ctx, kafka.Message{Topic: topic, Key: []byte(key), Value: value})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
