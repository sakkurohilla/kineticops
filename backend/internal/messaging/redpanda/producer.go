package redpanda

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

var Writer *kafka.Writer

func InitProducer(brokers []string, topic string) {
	Writer = &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}

func PublishEvent(msg []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return Writer.WriteMessages(ctx, kafka.Message{Value: msg})
}
