package redpanda

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

func StartConsumer(brokers []string, topic string, cb func([]byte)) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: "kineticops-group",
	})
	go func() {
		for {
			m, err := r.ReadMessage(context.Background())
			if err != nil {
				fmt.Println("[ERROR] Kafka consumer:", err)
				continue
			}
			fmt.Println("[DEBUG] Received from Redpanda/Kafka:", string(m.Value)) // <-- Add this line
			cb(m.Value)
		}
	}()
}
