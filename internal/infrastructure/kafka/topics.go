package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

func EnsureTopics(ctx context.Context, brokers []string, topics []string, partitions, replicationFactor int) error {
	dialer := &kafka.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	configs := make([]kafka.TopicConfig, 0, len(topics))
	for _, topic := range topics {
		configs = append(configs, kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     partitions,
			ReplicationFactor: replicationFactor,
		})
	}
	return conn.CreateTopics(configs...)
}
