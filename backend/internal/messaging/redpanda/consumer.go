package redpanda

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/segmentio/kafka-go"
)

// StartConsumer starts a Kafka reader that processes messages with cb.
// It waits until at least one broker is reachable to avoid noisy errors when
// brokers are not yet ready or advertise loopback addresses. Read errors
// are backoff-retried to reduce log spam.
func StartConsumer(brokers []string, topic string, cb func([]byte)) {
	go func() {
		// wait for reachable broker
		reachable := false
		for !reachable {
			for _, b := range brokers {
				conn, err := net.DialTimeout("tcp", b, 2*time.Second)
				if err == nil {
					_ = conn.Close()
					reachable = true
					break
				}
			}
			if !reachable {
				fmt.Printf("[KAFKA] no reachable brokers yet (%v), retrying in 5s...\n", brokers)
				time.Sleep(5 * time.Second)
			}
		}

		r := kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: "kineticops-group",
		})
		defer r.Close()

		backoff := time.Second
		for {
			m, err := r.ReadMessage(context.Background())
			if err != nil {
				fmt.Println("[ERROR] Kafka consumer:", err)
				// simple backoff to avoid tight error loops on transient DNS or network issues
				time.Sleep(backoff)
				if backoff < 30*time.Second {
					backoff *= 2
				}
				continue
			}
			// reset backoff after a successful read
			backoff = time.Second
			fmt.Println("[DEBUG] Received from Redpanda/Kafka:", string(m.Value))
			cb(m.Value)
		}
	}()
}
