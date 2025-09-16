package probe

import (
	"fmt"
	"net/url"
)

// ErrorType represents different categories of errors
type ErrorType string

const (
	// ErrorTypeNetwork indicates network-related errors
	ErrorTypeNetwork ErrorType = "network"
	// ErrorTypeParsing indicates manifest parsing errors
	ErrorTypeParsing ErrorType = "parsing"
	// ErrorTypeValidation indicates input validation errors
	ErrorTypeValidation ErrorType = "validation"
	// ErrorTypeTimeout indicates timeout errors
	ErrorTypeTimeout ErrorType = "timeout"
	// ErrorTypeAuth indicates authentication/authorization errors
	ErrorTypeAuth ErrorType = "auth"
)

// ProbeError represents a structured error with context
type ProbeError struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
	URL     string    `json:"url,omitempty"`
	Cause   error     `json:"-"`
}

// Error implements the error interface
func (e *ProbeError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying cause for error wrapping
func (e *ProbeError) Unwrap() error {
	return e.Cause
}

// IsType checks if the error is of a specific type
func (e *ProbeError) IsType(errorType ErrorType) bool {
	return e.Type == errorType
}

// NewNetworkError creates a new network-related error
func NewNetworkError(url string, cause error) *ProbeError {
	return &ProbeError{
		Type:    ErrorTypeNetwork,
		Message: fmt.Sprintf("failed to fetch manifest from %s", url),
		URL:     url,
		Cause:   cause,
	}
}

// NewParsingError creates a new parsing-related error
func NewParsingError(url string, format string, cause error) *ProbeError {
	return &ProbeError{
		Type:    ErrorTypeParsing,
		Message: fmt.Sprintf("failed to parse %s manifest", format),
		URL:     url,
		Cause:   cause,
	}
}

// NewValidationError creates a new validation-related error
func NewValidationError(message string) *ProbeError {
	return &ProbeError{
		Type:    ErrorTypeValidation,
		Message: message,
	}
}

// NewTimeoutError creates a new timeout-related error
func NewTimeoutError(url string, timeoutSeconds int) *ProbeError {
	return &ProbeError{
		Type:    ErrorTypeTimeout,
		Message: fmt.Sprintf("request timed out after %d seconds", timeoutSeconds),
		URL:     url,
	}
}

// NewAuthError creates a new authentication-related error
func NewAuthError(url string, statusCode int) *ProbeError {
	return &ProbeError{
		Type:    ErrorTypeAuth,
		Message: fmt.Sprintf("authentication failed (HTTP %d)", statusCode),
		URL:     url,
	}
}

// validateURL validates and normalizes a URL
func validateURL(rawURL string) (*url.URL, error) {
	if rawURL == "" {
		return nil, NewValidationError("URL cannot be empty")
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, NewValidationError(fmt.Sprintf("invalid URL format: %v", err))
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, NewValidationError(fmt.Sprintf("unsupported URL scheme: %s (only http/https allowed)", parsedURL.Scheme))
	}

	if parsedURL.Host == "" {
		return nil, NewValidationError("URL must have a valid host")
	}

	return parsedURL, nil
}

// validateProbeOptions validates probe options
func validateProbeOptions(opts *ProbeOptions) error {
	if opts == nil {
		return nil
	}

	if opts.ProxyURL != "" {
		if _, err := url.Parse(opts.ProxyURL); err != nil {
			return NewValidationError(fmt.Sprintf("invalid proxy URL: %v", err))
		}
	}

	if opts.TimeoutSeconds < 0 {
		return NewValidationError("timeout cannot be negative")
	}

	if opts.TimeoutSeconds > 300 {
		return NewValidationError("timeout cannot exceed 300 seconds")
	}

	return nil
}