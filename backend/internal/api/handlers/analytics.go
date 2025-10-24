// backend/internal/api/handlers/analytics.go

package handlers

import (
	"github.com/gin-gonic/gin"
	// ...other imports
)

func GetAggregatedMetrics(c *gin.Context) {
	// Parse params for from, to, interval, metric
	// Query data from DB for that interval
	// Call services.AggregateMetrics on each bucket
	// Return as JSON
}
