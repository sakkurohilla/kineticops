// backend/internal/services/aggregation_service.go

package services

import (
	"sort"
)

type Metric struct {
	Timestamp int64
	Value     float64
}

func AggregateMetrics(metrics []Metric) (min, max, avg, p95 float64) {
	if len(metrics) == 0 {
		return
	}
	sum := 0.0
	min, max = metrics[0].Value, metrics[0].Value
	values := make([]float64, len(metrics))
	for i, m := range metrics {
		sum += m.Value
		if m.Value < min {
			min = m.Value
		}
		if m.Value > max {
			max = m.Value
		}
		values[i] = m.Value
	}
	avg = sum / float64(len(metrics))
	sort.Float64s(values)
	p95 = values[int(float64(len(values))*0.95)]
	return
}
