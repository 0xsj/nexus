// pkg/grpc/health/health.go

package health

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/0xsj/nexus-go/pkg/observability/logger"
)

// ============================================================================
// Health Checker Interface
// ============================================================================

// Checker is the interface for health checks.
type Checker interface {
	// Name returns the name of the component being checked.
	Name() string

	// Check performs the health check and returns an error if unhealthy.
	Check(ctx context.Context) error
}

// CheckerFunc is a function adapter for Checker.
type CheckerFunc struct {
	name  string
	check func(ctx context.Context) error
}

// NewCheckerFunc creates a new CheckerFunc.
func NewCheckerFunc(name string, check func(ctx context.Context) error) *CheckerFunc {
	return &CheckerFunc{
		name:  name,
		check: check,
	}
}

// Name returns the checker name.
func (c *CheckerFunc) Name() string {
	return c.name
}

// Check performs the health check.
func (c *CheckerFunc) Check(ctx context.Context) error {
	return c.check(ctx)
}

// ============================================================================
// Health Status
// ============================================================================

// Status represents the health status of a component.
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
	StatusUnknown   Status = "unknown"
)

// ComponentStatus represents the health status of a single component.
type ComponentStatus struct {
	Name      string        `json:"name"`
	Status    Status        `json:"status"`
	Message   string        `json:"message,omitempty"`
	Latency   time.Duration `json:"latency_ms"`
	CheckedAt time.Time     `json:"checked_at"`
}

// OverallStatus represents the overall health status.
type OverallStatus struct {
	Status     Status                     `json:"status"`
	Components map[string]ComponentStatus `json:"components"`
	CheckedAt  time.Time                  `json:"checked_at"`
}

// ============================================================================
// Health Service
// ============================================================================

// Service manages health checks and integrates with gRPC health.
type Service struct {
	grpcHealth *health.Server
	logger     logger.Logger

	checkers []Checker
	mu       sync.RWMutex

	// Background check configuration
	checkInterval time.Duration
	checkTimeout  time.Duration
	stopCh        chan struct{}
}

// ServiceOption configures the health service.
type ServiceOption func(*Service)

// WithCheckInterval sets the interval for background health checks.
func WithCheckInterval(interval time.Duration) ServiceOption {
	return func(s *Service) {
		s.checkInterval = interval
	}
}

// WithCheckTimeout sets the timeout for individual health checks.
func WithCheckTimeout(timeout time.Duration) ServiceOption {
	return func(s *Service) {
		s.checkTimeout = timeout
	}
}

// WithLogger sets the logger for the health service.
func WithLogger(log logger.Logger) ServiceOption {
	return func(s *Service) {
		s.logger = log
	}
}

// NewService creates a new health service.
func NewService(grpcHealth *health.Server, opts ...ServiceOption) *Service {
	s := &Service{
		grpcHealth:    grpcHealth,
		checkInterval: 30 * time.Second,
		checkTimeout:  5 * time.Second,
		stopCh:        make(chan struct{}),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// RegisterChecker registers a health checker.
func (s *Service) RegisterChecker(checker Checker) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.checkers = append(s.checkers, checker)

	// Set initial status as unknown
	if s.grpcHealth != nil {
		s.grpcHealth.SetServingStatus(
			checker.Name(),
			grpc_health_v1.HealthCheckResponse_UNKNOWN,
		)
	}
}

// RegisterCheckerFunc registers a health check function.
func (s *Service) RegisterCheckerFunc(name string, check func(ctx context.Context) error) {
	s.RegisterChecker(NewCheckerFunc(name, check))
}

// ============================================================================
// Health Checks
// ============================================================================

// Check performs all health checks and returns the overall status.
func (s *Service) Check(ctx context.Context) *OverallStatus {
	s.mu.RLock()
	checkers := make([]Checker, len(s.checkers))
	copy(checkers, s.checkers)
	s.mu.RUnlock()

	overall := &OverallStatus{
		Status:     StatusHealthy,
		Components: make(map[string]ComponentStatus),
		CheckedAt:  time.Now().UTC(),
	}

	// Run checks concurrently
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, checker := range checkers {
		wg.Add(1)
		go func(c Checker) {
			defer wg.Done()

			status := s.checkComponent(ctx, c)

			mu.Lock()
			overall.Components[c.Name()] = status

			// Update overall status
			if status.Status == StatusUnhealthy && overall.Status == StatusHealthy {
				overall.Status = StatusUnhealthy
			} else if status.Status == StatusDegraded && overall.Status == StatusHealthy {
				overall.Status = StatusDegraded
			}
			mu.Unlock()
		}(checker)
	}

	wg.Wait()

	return overall
}

// checkComponent performs a health check for a single component.
func (s *Service) checkComponent(ctx context.Context, checker Checker) ComponentStatus {
	// Create timeout context
	checkCtx, cancel := context.WithTimeout(ctx, s.checkTimeout)
	defer cancel()

	start := time.Now()
	err := checker.Check(checkCtx)
	latency := time.Since(start)

	status := ComponentStatus{
		Name:      checker.Name(),
		Latency:   latency,
		CheckedAt: time.Now().UTC(),
	}

	if err != nil {
		status.Status = StatusUnhealthy
		status.Message = err.Error()

		// Update gRPC health status
		if s.grpcHealth != nil {
			s.grpcHealth.SetServingStatus(
				checker.Name(),
				grpc_health_v1.HealthCheckResponse_NOT_SERVING,
			)
		}
	} else {
		status.Status = StatusHealthy

		// Update gRPC health status
		if s.grpcHealth != nil {
			s.grpcHealth.SetServingStatus(
				checker.Name(),
				grpc_health_v1.HealthCheckResponse_SERVING,
			)
		}
	}

	return status
}

// CheckComponent performs a health check for a specific component.
func (s *Service) CheckComponent(ctx context.Context, name string) (*ComponentStatus, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, checker := range s.checkers {
		if checker.Name() == name {
			status := s.checkComponent(ctx, checker)
			return &status, true
		}
	}

	return nil, false
}

// ============================================================================
// Background Checks
// ============================================================================

// StartBackgroundChecks starts periodic background health checks.
func (s *Service) StartBackgroundChecks(ctx context.Context) {
	if s.logger != nil {
		s.logger.Info("starting background health checks",
			logger.Duration("interval", s.checkInterval),
		)
	}

	// Run initial check
	s.Check(ctx)

	go func() {
		ticker := time.NewTicker(s.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.Check(ctx)
			case <-s.stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// StopBackgroundChecks stops periodic background health checks.
func (s *Service) StopBackgroundChecks() {
	close(s.stopCh)
}

// ============================================================================
// gRPC Health Integration
// ============================================================================

// SetServingStatus sets the serving status for a service.
func (s *Service) SetServingStatus(service string, serving bool) {
	if s.grpcHealth == nil {
		return
	}

	status := grpc_health_v1.HealthCheckResponse_SERVING
	if !serving {
		status = grpc_health_v1.HealthCheckResponse_NOT_SERVING
	}

	s.grpcHealth.SetServingStatus(service, status)
}

// SetOverallServingStatus sets the overall serving status.
func (s *Service) SetOverallServingStatus(serving bool) {
	s.SetServingStatus("", serving)
}

// ============================================================================
// Common Checkers
// ============================================================================

// DatabaseChecker creates a health checker for database connections.
func DatabaseChecker(name string, pingFunc func(ctx context.Context) error) Checker {
	return NewCheckerFunc(name, pingFunc)
}

// CacheChecker creates a health checker for cache connections.
func CacheChecker(name string, pingFunc func(ctx context.Context) error) Checker {
	return NewCheckerFunc(name, pingFunc)
}

// ExternalServiceChecker creates a health checker for external services.
func ExternalServiceChecker(name string, checkFunc func(ctx context.Context) error) Checker {
	return NewCheckerFunc(name, checkFunc)
}

// ============================================================================
// Liveness vs Readiness
// ============================================================================

// LivenessStatus returns a simple liveness status.
// Liveness indicates the application is running (not deadlocked).
func (s *Service) LivenessStatus() Status {
	// If we can respond, we're alive
	return StatusHealthy
}

// ReadinessStatus returns whether the service is ready to accept traffic.
// Readiness indicates all dependencies are available.
func (s *Service) ReadinessStatus(ctx context.Context) Status {
	overall := s.Check(ctx)
	return overall.Status
}
