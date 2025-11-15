package circuitbreaker

import (
	"sync"
	"time"
)

// State represents the circuit breaker state.
type State int

const (
	// StateClosed means requests pass through normally.
	StateClosed State = iota

	// StateOpen means requests are rejected immediately.
	StateOpen

	// StateHalfOpen means a limited number of test requests are allowed.
	StateHalfOpen
)

// String returns the string representation of the state.
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// stateManager manages circuit breaker state transitions.
type stateManager struct {
	mu sync.RWMutex

	currentState State

	// Counters
	consecutiveSuccesses int
	consecutiveFailures  int

	// Timestamps
	lastStateChange time.Time
	lastFailureTime time.Time
}

// newStateManager creates a new state manager.
func newStateManager() *stateManager {
	return &stateManager{
		currentState:    StateClosed,
		lastStateChange: time.Now(),
	}
}

// getState returns the current state.
func (sm *stateManager) getState() State {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.currentState
}

// setState transitions to a new state.
func (sm *stateManager) setState(newState State) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.currentState != newState {
		sm.currentState = newState
		sm.lastStateChange = time.Now()
		sm.consecutiveSuccesses = 0
		sm.consecutiveFailures = 0
	}
}

// recordSuccess increments consecutive successes and resets failures.
func (sm *stateManager) recordSuccess() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.consecutiveSuccesses++
	sm.consecutiveFailures = 0
}

// recordFailure increments consecutive failures and resets successes.
func (sm *stateManager) recordFailure() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.consecutiveFailures++
	sm.consecutiveSuccesses = 0
	sm.lastFailureTime = time.Now()
}

// getConsecutiveSuccesses returns the count of consecutive successes.
func (sm *stateManager) getConsecutiveSuccesses() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.consecutiveSuccesses
}

// getConsecutiveFailures returns the count of consecutive failures.
func (sm *stateManager) getConsecutiveFailures() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.consecutiveFailures
}

// timeSinceStateChange returns duration since last state change.
func (sm *stateManager) timeSinceStateChange() time.Duration {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return time.Since(sm.lastStateChange)
}

// reset resets all counters.
func (sm *stateManager) reset() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.consecutiveSuccesses = 0
	sm.consecutiveFailures = 0
}
