package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	"github.com/segmentio/kafka-go"
)

// StartReingestConsumer listens to a Redpanda topic for failed metric batches and re-inserts them.
func StartReingestConsumer(brokers []string, topic string, groupID string) {
	go func() {
		// Wait until at least one broker is reachable to avoid noisy repeated
		// kafka-go connection errors when brokers advertise loopback addresses
		// or are still starting. This reduces log spam and allows the rest of
		// the service to operate while Redpanda boots.
		reachable := false
		for !reachable {
			for _, b := range brokers {
				conn, err := net.DialTimeout("tcp", b, 2*time.Second)
				if err == nil {
					conn.Close()
					reachable = true
					break
				}
			}
			if !reachable {
				fmt.Printf("[REINGEST] no reachable Redpanda brokers yet, retrying in 5s...\n")
				time.Sleep(5 * time.Second)
			}
		}

		r := kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		})
		defer r.Close()
		fmt.Printf("[REINGEST] started consumer for topic=%s\n", topic)
		for {
			m, err := r.ReadMessage(context.Background())
			if err != nil {
				fmt.Printf("[REINGEST] read error: %v\n", err)
				time.Sleep(1 * time.Second)
				continue
			}
			var batch []*models.Metric
			if err := json.Unmarshal(m.Value, &batch); err != nil {
				fmt.Printf("[REINGEST] json unmarshal failed: %v\n", err)
				continue
			}
			if len(batch) == 0 {
				continue
			}
			// Attempt to save batch - if fails, log and continue (message remains consumed)
			if err := postgres.SaveMetricsBatch(postgres.DB, batch); err != nil {
				fmt.Printf("[REINGEST] failed to save batch of %d: %v\n", len(batch), err)
			} else {
				fmt.Printf("[REINGEST] reingested %d metrics\n", len(batch))
			}
		}
	}()
}
