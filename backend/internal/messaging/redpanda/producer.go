package redpanda

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

var Writer *kafka.Writer

type Producer struct {
	writer *kafka.Writer
}

func InitProducer(brokers []string, topic string) *Producer {
	Writer = &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	return &Producer{writer: Writer}
}

func (p *Producer) SendMessage(msg []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.writer.WriteMessages(ctx, kafka.Message{Value: msg})
}

func PublishEvent(msg []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return Writer.WriteMessages(ctx, kafka.Message{Value: msg})
}

// ProducerPing attempts to dial the first broker to verify connectivity.
func ProducerPing(brokers []string, timeout time.Duration) error {
	if len(brokers) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	// Dial the first broker
	conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}
