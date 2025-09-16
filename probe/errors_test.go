package probe

import (
	"errors"
	"testing"
)

func TestProbeError(t *testing.T) {
	tests := []struct {
		name     string
		error    *ProbeError
		expected string
		isType   ErrorType
	}{
		{
			name: "network error with cause",
			error: NewNetworkError("https://example.com", errors.New("connection refused")),
			expected: "network: failed to fetch manifest from https://example.com (caused by: connection refused)",
			isType: ErrorTypeNetwork,
		},
		{
			name: "parsing error",
			error: NewParsingError("https://example.com/manifest.mpd", "MPD", errors.New("invalid XML")),
			expected: "parsing: failed to parse MPD manifest (caused by: invalid XML)",
			isType: ErrorTypeParsing,
		},
		{
			name: "validation error",
			error: NewValidationError("URL cannot be empty"),
			expected: "validation: URL cannot be empty",
			isType: ErrorTypeValidation,
		},
		{
			name: "timeout error",
			error: NewTimeoutError("https://example.com", 30),
			expected: "timeout: request timed out after 30 seconds",
			isType: ErrorTypeTimeout,
		},
		{
			name: "auth error",
			error: NewAuthError("https://example.com", 401),
			expected: "auth: authentication failed (HTTP 401)",
			isType: ErrorTypeAuth,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.error.Error() != tt.expected {
				t.Errorf("Expected error message %q, got %q", tt.expected, tt.error.Error())
			}
			
			if !tt.error.IsType(tt.isType) {
				t.Errorf("Expected error type %v, got %v", tt.isType, tt.error.Type)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectError bool
		errorType   ErrorType
	}{
		{
			name:        "valid HTTP URL",
			url:         "http://example.com/manifest.mpd",
			expectError: false,
		},
		{
			name:        "valid HTTPS URL",
			url:         "https://example.com/manifest.m3u8",
			expectError: false,
		},
		{
			name:        "empty URL",
			url:         "",
			expectError: true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "invalid scheme",
			url:         "ftp://example.com/manifest.mpd",
			expectError: true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "no host",
			url:         "https:///manifest.mpd",
			expectError: true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "malformed URL",
			url:         "not-a-url",
			expectError: true,
			errorType:   ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateURL(tt.url)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				
				var probeErr *ProbeError
				if !errors.As(err, &probeErr) {
					t.Errorf("Expected ProbeError, got %T", err)
					return
				}
				
				if probeErr.Type != tt.errorType {
					t.Errorf("Expected error type %v, got %v", tt.errorType, probeErr.Type)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestValidateProbeOptions(t *testing.T) {
	tests := []struct {
		name        string
		opts        *ProbeOptions
		expectError bool
		errorType   ErrorType
	}{
		{
			name:        "nil options",
			opts:        nil,
			expectError: false,
		},
		{
			name: "valid options",
			opts: &ProbeOptions{
				ProxyURL:       "http://proxy:8080",
				TimeoutSeconds: 30,
			},
			expectError: false,
		},
		{
			name: "invalid proxy URL",
			opts: &ProbeOptions{
				ProxyURL: "://invalid-url",
			},
			expectError: true,
			errorType:   ErrorTypeValidation,
		},
		{
			name: "negative timeout",
			opts: &ProbeOptions{
				TimeoutSeconds: -1,
			},
			expectError: true,
			errorType:   ErrorTypeValidation,
		},
		{
			name: "timeout too large",
			opts: &ProbeOptions{
				TimeoutSeconds: 400,
			},
			expectError: true,
			errorType:   ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProbeOptions(tt.opts)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				
				var probeErr *ProbeError
				if !errors.As(err, &probeErr) {
					t.Errorf("Expected ProbeError, got %T", err)
					return
				}
				
				if probeErr.Type != tt.errorType {
					t.Errorf("Expected error type %v, got %v", tt.errorType, probeErr.Type)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}