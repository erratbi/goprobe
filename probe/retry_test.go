package probe

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()
	
	if config.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries 3, got %d", config.MaxRetries)
	}
	
	if config.InitialDelay != 100*time.Millisecond {
		t.Errorf("Expected InitialDelay 100ms, got %v", config.InitialDelay)
	}
	
	if config.MaxDelay != 5*time.Second {
		t.Errorf("Expected MaxDelay 5s, got %v", config.MaxDelay)
	}
	
	if config.BackoffMultiplier != 2.0 {
		t.Errorf("Expected BackoffMultiplier 2.0, got %f", config.BackoffMultiplier)
	}
	
	if !config.Jitter {
		t.Error("Expected Jitter to be true")
	}
	
	expectedRetryable := []ErrorType{ErrorTypeNetwork, ErrorTypeTimeout}
	if len(config.RetryableErrors) != len(expectedRetryable) {
		t.Errorf("Expected %d retryable errors, got %d", len(expectedRetryable), len(config.RetryableErrors))
	}
}

func TestCircuitBreakerConfig(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	
	if !config.Enabled {
		t.Error("Expected circuit breaker to be enabled")
	}
	
	if config.FailureThreshold != 5 {
		t.Errorf("Expected FailureThreshold 5, got %d", config.FailureThreshold)
	}
	
	if config.ResetTimeout != 30*time.Second {
		t.Errorf("Expected ResetTimeout 30s, got %v", config.ResetTimeout)
	}
	
	if config.HalfOpenMaxRequests != 3 {
		t.Errorf("Expected HalfOpenMaxRequests 3, got %d", config.HalfOpenMaxRequests)
	}
}

func TestCircuitBreakerStates(t *testing.T) {
	config := &CircuitBreakerConfig{
		Enabled:             true,
		FailureThreshold:    2,
		ResetTimeout:        100 * time.Millisecond,
		HalfOpenMaxRequests: 1,
	}
	
	cb := NewCircuitBreaker(config)
	ctx := context.Background()
	
	// Initially closed
	if cb.GetState() != CircuitStateClosed {
		t.Errorf("Expected circuit to be closed initially")
	}
	
	// Fail enough times to open circuit
	networkErr := NewNetworkError("http://test.com", errors.New("connection failed"))
	
	for i := 0; i < config.FailureThreshold; i++ {
		err := cb.Execute(ctx, func() error {
			return networkErr
		})
		if err == nil {
			t.Error("Expected error from failed operation")
		}
	}
	
	// Should be open now
	if cb.GetState() != CircuitStateOpen {
		t.Errorf("Expected circuit to be open after failures")
	}
	
	// Should reject requests when open
	err := cb.Execute(ctx, func() error {
		return nil
	})
	if err == nil {
		t.Error("Expected circuit breaker to reject request when open")
	}
	
	// Wait for reset timeout
	time.Sleep(config.ResetTimeout + 10*time.Millisecond)
	
	// Should transition to half-open
	err = cb.Execute(ctx, func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected successful request in half-open state, got: %v", err)
	}
	
	// Should be closed after successful request
	if cb.GetState() != CircuitStateClosed {
		t.Errorf("Expected circuit to be closed after successful half-open request")
	}
}

func TestRetryExecutorSuccess(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:        2,
		InitialDelay:      1 * time.Millisecond,
		MaxDelay:          10 * time.Millisecond,
		BackoffMultiplier: 2.0,
		Jitter:            false,
		RetryableErrors:   []ErrorType{ErrorTypeNetwork},
	}
	
	executor := NewRetryExecutor(config, nil)
	ctx := context.Background()
	
	attempts := 0
	err := executor.Execute(ctx, func() error {
		attempts++
		if attempts < 2 {
			return NewNetworkError("http://test.com", errors.New("temporary failure"))
		}
		return nil
	})
	
	if err != nil {
		t.Errorf("Expected success after retry, got: %v", err)
	}
	
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetryExecutorMaxRetriesExceeded(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:        2,
		InitialDelay:      1 * time.Millisecond,
		MaxDelay:          10 * time.Millisecond,
		BackoffMultiplier: 2.0,
		Jitter:            false,
		RetryableErrors:   []ErrorType{ErrorTypeNetwork},
	}
	
	executor := NewRetryExecutor(config, nil)
	ctx := context.Background()
	
	attempts := 0
	networkErr := NewNetworkError("http://test.com", errors.New("persistent failure"))
	
	err := executor.Execute(ctx, func() error {
		attempts++
		return networkErr
	})
	
	if err == nil {
		t.Error("Expected error after max retries exceeded")
	}
	
	expectedAttempts := config.MaxRetries + 1 // Initial attempt + retries
	if attempts != expectedAttempts {
		t.Errorf("Expected %d attempts, got %d", expectedAttempts, attempts)
	}
}

func TestRetryExecutorNonRetryableError(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:        3,
		InitialDelay:      1 * time.Millisecond,
		MaxDelay:          10 * time.Millisecond,
		BackoffMultiplier: 2.0,
		Jitter:            false,
		RetryableErrors:   []ErrorType{ErrorTypeNetwork},
	}
	
	executor := NewRetryExecutor(config, nil)
	ctx := context.Background()
	
	attempts := 0
	authErr := NewAuthError("http://test.com", 401)
	
	err := executor.Execute(ctx, func() error {
		attempts++
		return authErr
	})
	
	if err == nil {
		t.Error("Expected error from non-retryable failure")
	}
	
	if attempts != 1 {
		t.Errorf("Expected 1 attempt for non-retryable error, got %d", attempts)
	}
	
	var probeErr *ProbeError
	if !errors.As(err, &probeErr) {
		t.Error("Expected ProbeError")
	}
	
	if probeErr.Type != ErrorTypeAuth {
		t.Errorf("Expected auth error, got %v", probeErr.Type)
	}
}

func TestRetryExecutorContextCancellation(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:        5,
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          1 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            false,
		RetryableErrors:   []ErrorType{ErrorTypeNetwork},
	}
	
	executor := NewRetryExecutor(config, nil)
	
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	
	attempts := 0
	err := executor.Execute(ctx, func() error {
		attempts++
		return NewNetworkError("http://test.com", errors.New("slow failure"))
	})
	
	if err == nil {
		t.Error("Expected context cancellation error")
	}
	
	if err != context.DeadlineExceeded {
		t.Errorf("Expected context deadline exceeded, got: %v", err)
	}
	
	// Should have attempted at least once before timeout
	if attempts < 1 {
		t.Errorf("Expected at least 1 attempt, got %d", attempts)
	}
}

func TestRetryExecutorWithCircuitBreaker(t *testing.T) {
	retryConfig := &RetryConfig{
		MaxRetries:        2,
		InitialDelay:      1 * time.Millisecond,
		MaxDelay:          10 * time.Millisecond,
		BackoffMultiplier: 2.0,
		Jitter:            false,
		RetryableErrors:   []ErrorType{ErrorTypeNetwork},
	}
	
	cbConfig := &CircuitBreakerConfig{
		Enabled:             true,
		FailureThreshold:    3,
		ResetTimeout:        100 * time.Millisecond,
		HalfOpenMaxRequests: 1,
	}
	
	executor := NewRetryExecutor(retryConfig, cbConfig)
	ctx := context.Background()
	
	// Fail enough times to open circuit breaker
	networkErr := NewNetworkError("http://test.com", errors.New("persistent failure"))
	
	for i := 0; i < cbConfig.FailureThreshold+1; i++ {
		err := executor.Execute(ctx, func() error {
			return networkErr
		})
		if err == nil {
			t.Errorf("Expected error on attempt %d", i+1)
		}
	}
	
	// Next request should be rejected by circuit breaker
	err := executor.Execute(ctx, func() error {
		t.Error("This function should not be called when circuit is open")
		return nil
	})
	
	if err == nil {
		t.Error("Expected circuit breaker to reject request")
	}
	
	var probeErr *ProbeError
	if errors.As(err, &probeErr) {
		if probeErr.Type != ErrorTypeNetwork || probeErr.Message != "circuit breaker is open" {
			t.Errorf("Expected circuit breaker error, got: %v", err)
		}
	} else {
		t.Errorf("Expected ProbeError from circuit breaker, got: %T", err)
	}
}

func TestCalculateDelay(t *testing.T) {
	config := &RetryConfig{
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          1 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            false,
	}
	
	executor := NewRetryExecutor(config, nil)
	
	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 100 * time.Millisecond},
		{1, 200 * time.Millisecond},
		{2, 400 * time.Millisecond},
		{3, 800 * time.Millisecond},
		{4, 1 * time.Second}, // Capped at MaxDelay
		{5, 1 * time.Second}, // Still capped
	}
	
	for _, tt := range tests {
		t.Run(fmt.Sprintf("attempt_%d", tt.attempt), func(t *testing.T) {
			delay := executor.calculateDelay(tt.attempt)
			if delay != tt.expected {
				t.Errorf("Expected delay %v for attempt %d, got %v", tt.expected, tt.attempt, delay)
			}
		})
	}
}

func TestCalculateDelayWithJitter(t *testing.T) {
	config := &RetryConfig{
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          1 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            true,
	}
	
	executor := NewRetryExecutor(config, nil)
	
	// Test multiple times to ensure jitter varies
	delays := make([]time.Duration, 10)
	for i := 0; i < 10; i++ {
		delays[i] = executor.calculateDelay(0)
	}
	
	// Check that delays are within expected range (75ms - 125ms with 25% jitter)
	for i, delay := range delays {
		if delay < 75*time.Millisecond || delay > 125*time.Millisecond {
			t.Errorf("Delay %d (%v) outside expected jitter range", i, delay)
		}
	}
	
	// Check that not all delays are identical (jitter should add randomness)
	allSame := true
	for i := 1; i < len(delays); i++ {
		if delays[i] != delays[0] {
			allSame = false
			break
		}
	}
	
	if allSame {
		t.Error("Expected jitter to create different delays, but all were identical")
	}
}