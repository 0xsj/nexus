package health

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"time"
)

// Check is a function that performs a health check.
type Check func(ctx context.Context) *Result

// CheckOptions configures a health check.
type CheckOptions struct {
	// Name is the check name.
	Name string

	// Timeout is the maximum time for the check.
	Timeout time.Duration

	// Critical indicates if failure should mark system as down.
	// Non-critical failures result in degraded status.
	Critical bool

	// Interval is how often to run the check (for background checks).
	Interval time.Duration
}

// DefaultCheckOptions returns default check options.
func DefaultCheckOptions(name string) CheckOptions {
	return CheckOptions{
		Name:     name,
		Timeout:  5 * time.Second,
		Critical: true,
		Interval: 30 * time.Second,
	}
}

// ============================================================================
// Database Checks
// ============================================================================

// SQLPinger is an interface for database ping (satisfied by *sql.DB).
type SQLPinger interface {
	PingContext(ctx context.Context) error
}

// Database creates a database health check.
func Database(db SQLPinger) Check {
	return func(ctx context.Context) *Result {
		start := time.Now()

		if db == nil {
			return Down(fmt.Errorf("database connection is nil")).
				WithDuration(time.Since(start))
		}

		if err := db.PingContext(ctx); err != nil {
			return Down(err).
				WithDuration(time.Since(start))
		}

		return Up().
			WithDuration(time.Since(start)).
			WithMessage("database is reachable")
	}
}

// DatabaseWithStats creates a database health check that includes connection stats.
func DatabaseWithStats(db *sql.DB) Check {
	return func(ctx context.Context) *Result {
		start := time.Now()

		if db == nil {
			return Down(fmt.Errorf("database connection is nil")).
				WithDuration(time.Since(start))
		}

		if err := db.PingContext(ctx); err != nil {
			return Down(err).
				WithDuration(time.Since(start))
		}

		stats := db.Stats()
		return Up().
			WithDuration(time.Since(start)).
			WithMessage("database is reachable").
			WithDetail("open_connections", stats.OpenConnections).
			WithDetail("in_use", stats.InUse).
			WithDetail("idle", stats.Idle).
			WithDetail("max_open", stats.MaxOpenConnections)
	}
}

// ============================================================================
// Redis Checks
// ============================================================================

// RedisPinger is an interface for Redis ping.
type RedisPinger interface {
	Ping(ctx context.Context) error
}

// Redis creates a Redis health check.
func Redis(client RedisPinger) Check {
	return func(ctx context.Context) *Result {
		start := time.Now()

		if client == nil {
			return Down(fmt.Errorf("redis client is nil")).
				WithDuration(time.Since(start))
		}

		if err := client.Ping(ctx); err != nil {
			return Down(err).
				WithDuration(time.Since(start))
		}

		return Up().
			WithDuration(time.Since(start)).
			WithMessage("redis is reachable")
	}
}

// ============================================================================
// HTTP Checks
// ============================================================================

// HTTPEndpoint creates an HTTP endpoint health check.
func HTTPEndpoint(url string) Check {
	return HTTPEndpointWithClient(url, http.DefaultClient)
}

// HTTPEndpointWithClient creates an HTTP endpoint health check with a custom client.
func HTTPEndpointWithClient(url string, client *http.Client) Check {
	return func(ctx context.Context) *Result {
		start := time.Now()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return Down(err).
				WithDuration(time.Since(start))
		}

		resp, err := client.Do(req)
		if err != nil {
			return Down(err).
				WithDuration(time.Since(start))
		}
		defer resp.Body.Close()

		duration := time.Since(start)

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return Up().
				WithDuration(duration).
				WithMessage(fmt.Sprintf("endpoint returned %d", resp.StatusCode)).
				WithDetail("status_code", resp.StatusCode)
		}

		if resp.StatusCode >= 500 {
			return Down(fmt.Errorf("endpoint returned %d", resp.StatusCode)).
				WithDuration(duration).
				WithDetail("status_code", resp.StatusCode)
		}

		return Degraded(fmt.Sprintf("endpoint returned %d", resp.StatusCode)).
			WithDuration(duration).
			WithDetail("status_code", resp.StatusCode)
	}
}

// ============================================================================
// TCP Checks
// ============================================================================

// TCP creates a TCP connectivity health check.
func TCP(host string, port int) Check {
	return TCPWithTimeout(host, port, 5*time.Second)
}

// TCPWithTimeout creates a TCP connectivity health check with custom timeout.
func TCPWithTimeout(host string, port int, timeout time.Duration) Check {
	return func(ctx context.Context) *Result {
		start := time.Now()
		address := fmt.Sprintf("%s:%d", host, port)

		dialer := &net.Dialer{Timeout: timeout}
		conn, err := dialer.DialContext(ctx, "tcp", address)
		if err != nil {
			return Down(err).
				WithDuration(time.Since(start)).
				WithDetail("address", address)
		}
		defer conn.Close()

		return Up().
			WithDuration(time.Since(start)).
			WithMessage(fmt.Sprintf("tcp connection to %s successful", address)).
			WithDetail("address", address)
	}
}

// ============================================================================
// DNS Checks
// ============================================================================

// DNS creates a DNS resolution health check.
func DNS(host string) Check {
	return func(ctx context.Context) *Result {
		start := time.Now()

		resolver := &net.Resolver{}
		addrs, err := resolver.LookupHost(ctx, host)
		if err != nil {
			return Down(err).
				WithDuration(time.Since(start)).
				WithDetail("host", host)
		}

		return Up().
			WithDuration(time.Since(start)).
			WithMessage(fmt.Sprintf("resolved %s to %d addresses", host, len(addrs))).
			WithDetail("host", host).
			WithDetail("addresses", addrs)
	}
}

// ============================================================================
// Disk Checks
// ============================================================================

// DiskSpaceChecker is an interface for checking disk space.
type DiskSpaceChecker interface {
	Available() (uint64, error)
	Total() (uint64, error)
}

// DiskSpace creates a disk space health check.
func DiskSpace(path string, minFreeBytes uint64) Check {
	return func(ctx context.Context) *Result {
		start := time.Now()

		// Note: Actual implementation would use syscall or a library
		// This is a placeholder that always returns up
		return Up().
			WithDuration(time.Since(start)).
			WithMessage("disk space check placeholder").
			WithDetail("path", path).
			WithDetail("min_free_bytes", minFreeBytes)
	}
}

// ============================================================================
// Memory Checks
// ============================================================================

// MemoryThreshold creates a memory usage health check.
func MemoryThreshold(maxUsagePercent float64) Check {
	return func(ctx context.Context) *Result {
		start := time.Now()

		// Note: Actual implementation would use runtime.MemStats
		// This is a placeholder that always returns up
		return Up().
			WithDuration(time.Since(start)).
			WithMessage("memory check placeholder").
			WithDetail("max_usage_percent", maxUsagePercent)
	}
}

// ============================================================================
// Custom Checks
// ============================================================================

// Custom creates a custom health check from a function.
func Custom(name string, fn func(ctx context.Context) error) Check {
	return func(ctx context.Context) *Result {
		start := time.Now()

		if err := fn(ctx); err != nil {
			return Down(err).
				WithDuration(time.Since(start))
		}

		return Up().
			WithDuration(time.Since(start)).
			WithMessage(name + " check passed")
	}
}

// Always creates a check that always returns the given status.
// Useful for testing.
func Always(status Status) Check {
	return func(ctx context.Context) *Result {
		return NewResult(status)
	}
}

// AlwaysUp creates a check that always returns up.
func AlwaysUp() Check {
	return Always(StatusUp)
}

// AlwaysDown creates a check that always returns down.
func AlwaysDown(err error) Check {
	return func(ctx context.Context) *Result {
		return Down(err)
	}
}

// ============================================================================
// Check Combinators
// ============================================================================

// All creates a check that passes only if all checks pass.
func All(checks ...Check) Check {
	return func(ctx context.Context) *Result {
		start := time.Now()

		for _, check := range checks {
			result := check(ctx)
			if !result.Status.IsHealthy() {
				return result.WithDuration(time.Since(start))
			}
		}

		return Up().
			WithDuration(time.Since(start)).
			WithMessage("all checks passed")
	}
}

// Any creates a check that passes if any check passes.
func Any(checks ...Check) Check {
	return func(ctx context.Context) *Result {
		start := time.Now()
		var lastErr error

		for _, check := range checks {
			result := check(ctx)
			if result.Status.IsHealthy() {
				return result.WithDuration(time.Since(start))
			}
			if result.Error != "" {
				lastErr = fmt.Errorf(result.Error)
			}
		}

		return Down(lastErr).
			WithDuration(time.Since(start)).
			WithMessage("all checks failed")
	}
}

// WithTimeout wraps a check with a timeout.
func WithTimeout(check Check, timeout time.Duration) Check {
	return func(ctx context.Context) *Result {
		start := time.Now()

		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		resultCh := make(chan *Result, 1)
		go func() {
			resultCh <- check(ctx)
		}()

		select {
		case result := <-resultCh:
			return result
		case <-ctx.Done():
			return Down(fmt.Errorf("check timed out after %s", timeout)).
				WithDuration(time.Since(start))
		}
	}
}

// Cached wraps a check with caching.
func Cached(check Check, ttl time.Duration) Check {
	var (
		cachedResult *Result
		cachedAt     time.Time
	)

	return func(ctx context.Context) *Result {
		now := time.Now()

		// Return cached result if still valid
		if cachedResult != nil && now.Sub(cachedAt) < ttl {
			return cachedResult
		}

		// Run check and cache result
		result := check(ctx)
		cachedResult = result
		cachedAt = now

		return result
	}
}
