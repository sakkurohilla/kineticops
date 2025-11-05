package redpanda

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

var Writer *kafka.Writer

type Producer struct {
	writer *kafka.Writer
}

func InitProducer(brokers []string, topic string) (*Producer, error) {
	// Quick dial to ensure broker is reachable before creating writer
	if len(brokers) == 0 {
		return nil, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return nil, err
	}
	_ = conn.Close()

	Writer = &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	return &Producer{writer: Writer}, nil
}

func (p *Producer) SendMessage(msg []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.writer.WriteMessages(ctx, kafka.Message{Value: msg})
}

func PublishEvent(msg []byte) error {
	if Writer == nil {
		return fmt.Errorf("kafka writer not initialized")
	}
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
