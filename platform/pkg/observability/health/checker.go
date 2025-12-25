package health

import (
	"context"
	"sync"
	"time"
)

// Checker manages health checks for the application.
type Checker struct {
	mu sync.RWMutex

	// checks is the registry of health checks.
	checks map[string]registeredCheck

	// version is the application version.
	version string

	// startTime is when the checker was created.
	startTime time.Time

	// started indicates if the application has fully started.
	started bool

	// defaultTimeout is the default timeout for checks.
	defaultTimeout time.Duration
}

// registeredCheck holds a check and its options.
type registeredCheck struct {
	check    Check
	options  CheckOptions
	lastRun  time.Time
	lastResult *Result
}

// NewChecker creates a new health checker.
func NewChecker() *Checker {
	return &Checker{
		checks:         make(map[string]registeredCheck),
		startTime:      time.Now(),
		defaultTimeout: 5 * time.Second,
	}
}

// WithVersion sets the application version.
func (c *Checker) WithVersion(version string) *Checker {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.version = version
	return c
}

// WithDefaultTimeout sets the default timeout for checks.
func (c *Checker) WithDefaultTimeout(timeout time.Duration) *Checker {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.defaultTimeout = timeout
	return c
}

// ============================================================================
// Registration
// ============================================================================

// Register adds a health check with default options.
func (c *Checker) Register(name string, check Check) {
	c.RegisterWithOptions(name, check, DefaultCheckOptions(name))
}

// RegisterWithOptions adds a health check with custom options.
func (c *Checker) RegisterWithOptions(name string, check Check, opts CheckOptions) {
	c.mu.Lock()
	defer c.mu.Unlock()

	opts.Name = name
	if opts.Timeout == 0 {
		opts.Timeout = c.defaultTimeout
	}

	c.checks[name] = registeredCheck{
		check:   check,
		options: opts,
	}
}

// RegisterNonCritical adds a non-critical health check.
// Failures will result in degraded status instead of down.
func (c *Checker) RegisterNonCritical(name string, check Check) {
	opts := DefaultCheckOptions(name)
	opts.Critical = false
	c.RegisterWithOptions(name, check, opts)
}

// Unregister removes a health check.
func (c *Checker) Unregister(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.checks, name)
}

// ============================================================================
// Startup Management
// ============================================================================

// MarkStarted marks the application as fully started.
func (c *Checker) MarkStarted() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.started = true
}

// MarkNotStarted marks the application as not started.
func (c *Checker) MarkNotStarted() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.started = false
}

// IsStarted returns whether the application has fully started.
func (c *Checker) IsStarted() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.started
}

// ============================================================================
// Check Execution
// ============================================================================

// Check runs all registered health checks and returns the overall status.
func (c *Checker) Check(ctx context.Context) *Response {
	c.mu.RLock()
	checks := make(map[string]registeredCheck, len(c.checks))
	for k, v := range c.checks {
		checks[k] = v
	}
	version := c.version
	startTime := c.startTime
	c.mu.RUnlock()

	// Run checks in parallel
	results := c.runChecksParallel(ctx, checks)

	// Determine overall status
	overallStatus := StatusUp
	for name, result := range results {
		if result.Status == StatusDown {
			if checks[name].options.Critical {
				overallStatus = StatusDown
				break
			} else if overallStatus != StatusDown {
				overallStatus = StatusDegraded
			}
		} else if result.Status == StatusDegraded && overallStatus == StatusUp {
			overallStatus = StatusDegraded
		}
	}

	response := NewResponse(overallStatus).
		WithVersion(version).
		WithUptime(time.Since(startTime))

	for name, result := range results {
		response.WithCheck(name, result)
	}

	return response
}

// CheckComponent runs a single health check.
func (c *Checker) CheckComponent(ctx context.Context, name string) *Result {
	c.mu.RLock()
	registered, exists := c.checks[name]
	c.mu.RUnlock()

	if !exists {
		return Down(nil).WithMessage("check not found: " + name)
	}

	return c.runCheck(ctx, registered)
}

// runChecksParallel runs all checks in parallel.
func (c *Checker) runChecksParallel(ctx context.Context, checks map[string]registeredCheck) map[string]*Result {
	results := make(map[string]*Result, len(checks))
	resultsMu := sync.Mutex{}
	wg := sync.WaitGroup{}

	for name, registered := range checks {
		wg.Add(1)
		go func(name string, registered registeredCheck) {
			defer wg.Done()

			result := c.runCheck(ctx, registered)

			resultsMu.Lock()
			results[name] = result
			resultsMu.Unlock()
		}(name, registered)
	}

	wg.Wait()
	return results
}

// runCheck runs a single check with timeout.
func (c *Checker) runCheck(ctx context.Context, registered registeredCheck) *Result {
	// Apply timeout
	checkCtx, cancel := context.WithTimeout(ctx, registered.options.Timeout)
	defer cancel()

	start := time.Now()

	// Run check in goroutine to handle timeout
	resultCh := make(chan *Result, 1)
	go func() {
		resultCh <- registered.check(checkCtx)
	}()

	select {
	case result := <-resultCh:
		// Update last run info
		c.mu.Lock()
		if reg, exists := c.checks[registered.options.Name]; exists {
			reg.lastRun = time.Now()
			reg.lastResult = result
			c.checks[registered.options.Name] = reg
		}
		c.mu.Unlock()
		return result

	case <-checkCtx.Done():
		return Down(ctx.Err()).
			WithDuration(time.Since(start)).
			WithMessage("check timed out")
	}
}

// ============================================================================
// Kubernetes Probes
// ============================================================================

// Liveness returns the liveness probe result.
// Returns up if the application is alive (not deadlocked/stuck).
func (c *Checker) Liveness(ctx context.Context) *LivenessResponse {
	// Liveness is simple - just confirm the app is running
	// Don't check dependencies here - that's for readiness
	return &LivenessResponse{
		Status: StatusUp,
	}
}

// Readiness returns the readiness probe result.
// Returns up if the application can serve traffic.
func (c *Checker) Readiness(ctx context.Context) *ReadinessResponse {
	c.mu.RLock()
	started := c.started
	checks := make(map[string]registeredCheck, len(c.checks))
	for k, v := range c.checks {
		checks[k] = v
	}
	c.mu.RUnlock()

	// If not started, not ready
	if !started {
		return &ReadinessResponse{
			Status: StatusDown,
		}
	}

	// Run all checks
	results := c.runChecksParallel(ctx, checks)

	// Determine readiness based on critical checks
	status := StatusUp
	checkStatuses := make(map[string]Status, len(results))

	for name, result := range results {
		checkStatuses[name] = result.Status

		if result.Status == StatusDown && checks[name].options.Critical {
			status = StatusDown
		}
	}

	return &ReadinessResponse{
		Status: status,
		Checks: checkStatuses,
	}
}

// Startup returns the startup probe result.
// Returns up once the application has completed initialization.
func (c *Checker) Startup(ctx context.Context) *StartupResponse {
	c.mu.RLock()
	started := c.started
	c.mu.RUnlock()

	status := StatusDown
	if started {
		status = StatusUp
	}

	return &StartupResponse{
		Status:  status,
		Started: started,
	}
}

// ============================================================================
// Info
// ============================================================================

// Uptime returns how long the checker has been running.
func (c *Checker) Uptime() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Since(c.startTime)
}

// Version returns the application version.
func (c *Checker) Version() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.version
}

// RegisteredChecks returns the names of all registered checks.
func (c *Checker) RegisteredChecks() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make([]string, 0, len(c.checks))
	for name := range c.checks {
		names = append(names, name)
	}
	return names
}

// LastResult returns the last result for a check.
func (c *Checker) LastResult(name string) *Result {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if registered, exists := c.checks[name]; exists {
		return registered.lastResult
	}
	return nil
}