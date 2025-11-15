// Package postgres provides PostgreSQL database connection and utilities.
package postgres

import (
	"context"
	"time"
)

// Health represents the health status of the database connection.
type Health struct {
	Status       HealthStatus  `json:"status"`
	Message      string        `json:"message,omitempty"`
	Timestamp    time.Time     `json:"timestamp"`
	ResponseTime time.Duration `json:"response_time"`
	Stats        PoolStats     `json:"stats"`
}

// HealthStatus represents the health status.
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
)

// PoolStats contains connection pool statistics.
type PoolStats struct {
	MaxConnections      int32 `json:"max_connections"`
	TotalConnections    int32 `json:"total_connections"`
	IdleConnections     int32 `json:"idle_connections"`
	AcquiredConnections int32 `json:"acquired_connections"`
}

// Health performs a health check on the database.
// It verifies the connection and returns statistics about the pool.
func (db *DB) Health(ctx context.Context) Health {
	start := time.Now()

	health := Health{
		Timestamp: start,
	}

	// Ping the database
	if err := db.Ping(ctx); err != nil {
		health.Status = HealthStatusUnhealthy
		health.Message = err.Error()
		health.ResponseTime = time.Since(start)
		return health
	}

	// Get pool statistics
	stat := db.Stats()
	health.Stats = PoolStats{
		MaxConnections:      stat.MaxConns(),
		TotalConnections:    stat.TotalConns(),
		IdleConnections:     stat.IdleConns(),
		AcquiredConnections: stat.AcquiredConns(),
	}

	// Determine health status based on pool usage
	if stat.AcquiredConns() >= stat.MaxConns() {
		health.Status = HealthStatusDegraded
		health.Message = "connection pool exhausted"
	} else {
		health.Status = HealthStatusHealthy
		health.Message = "database is healthy"
	}

	health.ResponseTime = time.Since(start)
	return health
}
