package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/erratbi/goprobe/probe"
)

func main() {
	// Example 1: Simple usage
	fmt.Println("=== Simple Usage ===")
	output, err := probe.ProbeManifest("https://bitdash-a.akamaihd.net/content/sintel/hls/playlist.m3u8", nil)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Found %d streams:\n", len(output.Streams))
		for _, stream := range output.Streams {
			fmt.Printf("  Stream %s: %s %s\n", stream.StreamID, stream.Type, stream.Codec)
		}
	}

	// Example 2: Advanced usage with retry and circuit breaker
	fmt.Println("\n=== Advanced Usage with Retry ===")
	opts := &probe.ProbeOptions{
		UserAgent:      "GoProbe-Example/1.0",
		TimeoutSeconds: 15,
		RetryConfig: &probe.RetryConfig{
			MaxRetries:        3,
			InitialDelay:      100 * time.Millisecond,
			MaxDelay:          5 * time.Second,
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	output, err = probe.ProbeManifestWithContext(ctx, "https://demo.unified-streaming.com/k8s/features/stable/video/tears-of-steel/tears-of-steel.ism/.m3u8", opts)
	if err != nil {
		log.Printf("Error with advanced options: %v", err)
	} else {
		// Output as JSON (ffprobe compatible)
		jsonData, err := output.OutputJSON()
		if err != nil {
			log.Printf("JSON encoding error: %v", err)
		} else {
			fmt.Printf("JSON Output:\n%s\n", string(jsonData))
		}
	}

	// Example 3: Error handling
	fmt.Println("\n=== Error Handling Example ===")
	_, err = probe.ProbeManifest("https://invalid-url.example.com/manifest.mpd", nil)
	if err != nil {
		var probeErr *probe.ProbeError
		if ok := errors.As(err, &probeErr); ok {
			fmt.Printf("Probe error type: %v\n", probeErr.Type)
			fmt.Printf("Error message: %s\n", probeErr.Message)
			if probeErr.URL != "" {
				fmt.Printf("Failed URL: %s\n", probeErr.URL)
			}
		} else {
			fmt.Printf("Other error: %v\n", err)
		}
	}
}