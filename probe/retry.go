package probe

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

// RetryConfig defines retry behavior configuration
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts (default: 3)
	MaxRetries int
	
	// InitialDelay is the initial delay before first retry (default: 100ms)
	InitialDelay time.Duration
	
	// MaxDelay is the maximum delay between retries (default: 5s)
	MaxDelay time.Duration
	
	// BackoffMultiplier for exponential backoff (default: 2.0)
	BackoffMultiplier float64
	
	// Jitter adds randomness to delays to avoid thundering herd (default: true)
	Jitter bool
	
	// RetryableErrors defines which error types should trigger retries
	RetryableErrors []ErrorType
}

// DefaultRetryConfig returns sensible defaults for retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:        3,
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          5 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            true,
		RetryableErrors:   []ErrorType{ErrorTypeNetwork, ErrorTypeTimeout},
	}
}

// CircuitBreakerConfig defines circuit breaker behavior
type CircuitBreakerConfig struct {
	// Enabled controls whether circuit breaker is active
	Enabled bool
	
	// FailureThreshold is the number of failures before opening circuit (default: 5)
	FailureThreshold int
	
	// ResetTimeout is how long to wait before attempting to close circuit (default: 30s)
	ResetTimeout time.Duration
	
	// HalfOpenMaxRequests is max requests allowed in half-open state (default: 3)
	HalfOpenMaxRequests int
}

// DefaultCircuitBreakerConfig returns sensible defaults
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		Enabled:             true,
		FailureThreshold:    5,
		ResetTimeout:        30 * time.Second,
		HalfOpenMaxRequests: 3,
	}
}

// CircuitState represents the current state of the circuit breaker
type CircuitState int

const (
	CircuitStateClosed CircuitState = iota
	CircuitStateOpen
	CircuitStateHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	config    *CircuitBreakerConfig
	state     CircuitState
	failures  int
	requests  int
	lastFailTime time.Time
	mutex     sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}
	
	return &CircuitBreaker{
		config: config,
		state:  CircuitStateClosed,
	}
}

// Execute runs the function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	if !cb.config.Enabled {
		return fn()
	}
	
	if !cb.allowRequest() {
		return &ProbeError{
			Type:    ErrorTypeNetwork,
			Message: "circuit breaker is open",
		}
	}
	
	err := fn()
	cb.recordResult(err)
	return err
}

// allowRequest checks if request should be allowed based on circuit state
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	now := time.Now()
	
	switch cb.state {
	case CircuitStateClosed:
		return true
		
	case CircuitStateOpen:
		if now.Sub(cb.lastFailTime) > cb.config.ResetTimeout {
			cb.state = CircuitStateHalfOpen
			cb.requests = 0
			return true
		}
		return false
		
	case CircuitStateHalfOpen:
		return cb.requests < cb.config.HalfOpenMaxRequests
		
	default:
		return false
	}
}

// recordResult updates circuit breaker state based on request result
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	if cb.state == CircuitStateHalfOpen {
		cb.requests++
	}
	
	if err != nil {
		cb.failures++
		cb.lastFailTime = time.Now()
		
		if cb.state == CircuitStateHalfOpen {
			cb.state = CircuitStateOpen
		} else if cb.failures >= cb.config.FailureThreshold {
			cb.state = CircuitStateOpen
		}
	} else {
		cb.failures = 0
		if cb.state == CircuitStateHalfOpen {
			cb.state = CircuitStateClosed
		}
	}
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// RetryExecutor handles retry logic with exponential backoff
type RetryExecutor struct {
	config         *RetryConfig
	circuitBreaker *CircuitBreaker
}

// NewRetryExecutor creates a new retry executor
func NewRetryExecutor(retryConfig *RetryConfig, cbConfig *CircuitBreakerConfig) *RetryExecutor {
	if retryConfig == nil {
		retryConfig = DefaultRetryConfig()
	}
	
	var cb *CircuitBreaker
	if cbConfig != nil {
		cb = NewCircuitBreaker(cbConfig)
	}
	
	return &RetryExecutor{
		config:         retryConfig,
		circuitBreaker: cb,
	}
}

// Execute runs the function with retry and circuit breaker logic
func (re *RetryExecutor) Execute(ctx context.Context, operation func() error) error {
	if re.circuitBreaker != nil {
		return re.circuitBreaker.Execute(ctx, func() error {
			return re.executeWithRetry(ctx, operation)
		})
	}
	
	return re.executeWithRetry(ctx, operation)
}

// executeWithRetry implements the retry logic with exponential backoff
func (re *RetryExecutor) executeWithRetry(ctx context.Context, operation func() error) error {
	var lastErr error
	
	for attempt := 0; attempt <= re.config.MaxRetries; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		// Execute the operation
		err := operation()
		if err == nil {
			if attempt > 0 {
				logInfo(ctx, "Operation succeeded after retry", map[string]interface{}{
					"attempt": attempt + 1,
				})
			}
			return nil
		}
		
		lastErr = err
		
		// Check if this error type is retryable
		if !re.isRetryable(err) {
			logDebug(ctx, "Error is not retryable", map[string]interface{}{
				"error": err.Error(),
				"attempt": attempt + 1,
			})
			return err
		}
		
		// Don't delay after last attempt
		if attempt == re.config.MaxRetries {
			logError(ctx, "Max retries exceeded", map[string]interface{}{
				"max_retries": re.config.MaxRetries,
				"final_error": err.Error(),
			})
			break
		}
		
		// Calculate delay for next attempt
		delay := re.calculateDelay(attempt)
		
		logWarn(ctx, "Operation failed, retrying", map[string]interface{}{
			"attempt": attempt + 1,
			"error": err.Error(),
			"delay": delay.String(),
		})
		
		// Wait before retry
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	
	return lastErr
}

// isRetryable checks if an error should trigger a retry
func (re *RetryExecutor) isRetryable(err error) bool {
	var probeErr *ProbeError
	if !errors.As(err, &probeErr) {
		return false
	}
	
	for _, retryableType := range re.config.RetryableErrors {
		if probeErr.Type == retryableType {
			return true
		}
	}
	
	return false
}

// calculateDelay computes the delay for the next retry attempt
func (re *RetryExecutor) calculateDelay(attempt int) time.Duration {
	delay := float64(re.config.InitialDelay) * math.Pow(re.config.BackoffMultiplier, float64(attempt))
	
	if re.config.Jitter {
		// Add 25% jitter
		jitter := delay * 0.25 * rand.Float64()
		delay += jitter
	}
	
	maxDelay := float64(re.config.MaxDelay)
	if delay > maxDelay {
		delay = maxDelay
	}
	
	return time.Duration(delay)
}