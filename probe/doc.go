/*
Package probe provides fast manifest parsing for DASH and HLS streaming protocols.

This package offers a high-performance alternative to ffprobe for analyzing
streaming manifests (DASH MPD and HLS M3U8) to extract stream information
including video codec, resolution, frame rate, audio codec, and more.

Key Features:
  - Fast manifest parsing (sub-second performance vs 9+ seconds with ffprobe)
  - Support for both DASH MPD and HLS M3U8 formats
  - ffprobe-compatible JSON output format
  - Production-grade retry mechanisms with circuit breaker pattern
  - Comprehensive error handling and logging
  - Proxy support and browser-like request camouflage
  - Context cancellation and timeout support

Basic Usage:

	package main

	import (
		"fmt"
		"log"
		"github.com/erratbi/goprobe/probe"
	)

	func main() {
		// Simple usage
		output, err := probe.ProbeManifest("https://example.com/manifest.mpd", nil)
		if err != nil {
			log.Fatal(err)
		}

		for _, stream := range output.Streams {
			fmt.Printf("Stream %s: %s %s\n", stream.StreamID, stream.Type, stream.Codec)
		}
	}

Advanced Usage with Options:

	opts := &probe.ProbeOptions{
		ProxyURL:       "http://proxy:8080",
		UserAgent:      "MyApp/1.0",
		TimeoutSeconds: 10,
		RetryConfig: &probe.RetryConfig{
			MaxRetries:        3,
			InitialDelay:      100 * time.Millisecond,
			BackoffMultiplier: 2.0,
			Jitter:            true,
			RetryableErrors:   []probe.ErrorType{probe.ErrorTypeNetwork, probe.ErrorTypeTimeout},
		},
		CircuitBreakerConfig: &probe.CircuitBreakerConfig{
			Enabled:             true,
			FailureThreshold:    5,
			ResetTimeout:        30 * time.Second,
			HalfOpenMaxRequests: 3,
		},
	}

	output, err := probe.ProbeManifest("https://example.com/manifest.mpd", opts)

The package automatically detects manifest format and returns structured stream
information compatible with ffprobe output format, making it a drop-in
replacement for many use cases.

Error Handling:

The package provides structured error types that can be inspected:

	output, err := probe.ProbeManifest(url, opts)
	if err != nil {
		var probeErr *probe.ProbeError
		if errors.As(err, &probeErr) {
			switch probeErr.Type {
			case probe.ErrorTypeNetwork:
				// Handle network errors
			case probe.ErrorTypeTimeout:
				// Handle timeout errors
			case probe.ErrorTypeParsing:
				// Handle parsing errors
			}
		}
	}
*/
package probe