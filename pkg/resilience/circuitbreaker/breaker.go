package circuitbreaker

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
)

var (
	// ErrCircuitOpen is returned when the circuit breaker is open.
	ErrCircuitOpen = errors.New("circuit breaker is open")

	// ErrTooManyRequests is returned when too many requests in half-open state.
	ErrTooManyRequests = errors.New("too many requests in half-open state")
)

// CircuitBreaker implements the circuit breaker pattern.
type CircuitBreaker struct {
	name         string
	config       Config
	stateManager *stateManager

	// Half-open state tracking
	halfOpenRequests atomic.Int32
}

// New creates a new circuit breaker.
func New(name string, config Config) (*CircuitBreaker, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &CircuitBreaker{
		name:         name,
		config:       config,
		stateManager: newStateManager(),
	}, nil
}

// Execute runs the given function with circuit breaker protection.
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	// Check if we can execute
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	// Execute the function
	err := fn()

	// Record result
	cb.afterRequest(err)

	return err
}

// ExecuteWithData runs the given function returning data with circuit breaker protection.
func ExecuteWithData[T any](cb *CircuitBreaker, ctx context.Context, fn func() (T, error)) (T, error) {
	var result T

	// Check if we can execute
	if err := cb.beforeRequest(); err != nil {
		return result, err
	}

	// Execute the function
	res, err := fn()

	// Record result
	cb.afterRequest(err)

	return res, err
}

// beforeRequest checks if the request should be allowed.
func (cb *CircuitBreaker) beforeRequest() error {
	state := cb.stateManager.getState()

	switch state {
	case StateClosed:
		// Allow request
		return nil

	case StateOpen:
		// Check if timeout has elapsed
		if cb.stateManager.timeSinceStateChange() >= cb.config.Timeout {
			// Transition to half-open
			cb.transitionTo(StateHalfOpen)
			return nil
		}
		// Circuit is open, reject request
		return ErrCircuitOpen

	case StateHalfOpen:
		// Check if we can allow more test requests
		current := cb.halfOpenRequests.Load()
		if current >= int32(cb.config.MaxHalfOpenRequests) {
			return ErrTooManyRequests
		}
		// Increment and allow
		cb.halfOpenRequests.Add(1)
		return nil

	default:
		return ErrCircuitOpen
	}
}

// afterRequest processes the request result.
func (cb *CircuitBreaker) afterRequest(err error) {
	state := cb.stateManager.getState()

	if err == nil {
		// Success
		cb.onSuccess(state)
	} else {
		// Failure
		cb.onFailure(state)
	}
}

// onSuccess handles a successful request.
func (cb *CircuitBreaker) onSuccess(state State) {
	switch state {
	case StateClosed:
		// Reset failure counter
		cb.stateManager.reset()

	case StateHalfOpen:
		// Record success
		cb.stateManager.recordSuccess()

		// Decrement half-open counter
		cb.halfOpenRequests.Add(-1)

		// Check if we can close
		if cb.stateManager.getConsecutiveSuccesses() >= cb.config.SuccessThreshold {
			cb.transitionTo(StateClosed)
			cb.halfOpenRequests.Store(0)
		}
	}
}

// onFailure handles a failed request.
func (cb *CircuitBreaker) onFailure(state State) {
	switch state {
	case StateClosed:
		// Record failure
		cb.stateManager.recordFailure()

		// Check if we should open
		if cb.stateManager.getConsecutiveFailures() >= cb.config.MaxFailures {
			cb.transitionTo(StateOpen)
		}

	case StateHalfOpen:
		// Any failure in half-open means back to open
		cb.halfOpenRequests.Store(0)
		cb.transitionTo(StateOpen)
	}
}

// transitionTo transitions to a new state.
func (cb *CircuitBreaker) transitionTo(newState State) {
	oldState := cb.stateManager.getState()

	if oldState == newState {
		return
	}

	cb.stateManager.setState(newState)

	// Call state change callback
	if cb.config.OnStateChange != nil {
		cb.config.OnStateChange(oldState, newState)
	}
}

// State returns the current circuit breaker state.
func (cb *CircuitBreaker) State() State {
	return cb.stateManager.getState()
}

// Name returns the circuit breaker name.
func (cb *CircuitBreaker) Name() string {
	return cb.name
}

// Reset manually resets the circuit breaker to closed state.
func (cb *CircuitBreaker) Reset() {
	cb.stateManager.setState(StateClosed)
	cb.stateManager.reset()
	cb.halfOpenRequests.Store(0)
}
