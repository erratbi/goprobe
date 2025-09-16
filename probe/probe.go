// Package probe provides fast manifest parsing for DASH and HLS streams.
// It analyzes MPD and M3U8 manifests to extract stream information including
// video codec, resolution, frame rate, audio codec, and more.
package probe

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// StreamInfo represents information about a media stream
type StreamInfo struct {
	StreamID   string `json:"stream_id"`
	Type       string `json:"type"`
	Codec      string `json:"codec"`
	PixFmt     string `json:"pix_fmt,omitempty"`
	Resolution string `json:"resolution,omitempty"`
	FrameRate  string `json:"frame_rate,omitempty"`
	BitRate    string `json:"bit_rate,omitempty"`
	Channels   string `json:"channels,omitempty"`
	SampleFmt  string `json:"sample_fmt,omitempty"`
	SampleRate string `json:"sample_rate,omitempty"`
	Language   string `json:"language,omitempty"`
}

// Output represents the complete probe output
type Output struct {
	Streams []StreamInfo `json:"streams"`
}

// ProbeOptions contains configuration for probing manifests
type ProbeOptions struct {
	// ProxyURL is the proxy server URL (e.g., "http://proxy:8080")
	ProxyURL string
	
	// UserAgent to use for requests (defaults to Chrome user agent)
	UserAgent string
	
	// CustomHeaders to add to requests
	CustomHeaders map[string]string
	
	// Timeout for HTTP requests in seconds (defaults to 30)
	TimeoutSeconds int
	
	// DisableCompression disables gzip/deflate compression
	DisableCompression bool
	
	// DisableCamouflage disables browser-like headers (origin, referer, etc.)
	DisableCamouflage bool
}

// ProbeManifest fetches and analyzes a streaming manifest URL.
// It automatically detects the format (DASH MPD or HLS M3U8) and returns
// structured stream information compatible with ffprobe output.
//
// Example:
//   output, err := probe.ProbeManifest("https://example.com/manifest.mpd", nil)
//   if err != nil {
//       log.Fatal(err)
//   }
//   
//   for _, stream := range output.Streams {
//       fmt.Printf("Stream %s: %s %s\n", stream.StreamID, stream.Type, stream.Codec)
//   }
func ProbeManifest(manifestURL string, opts *ProbeOptions) (*Output, error) {
	return ProbeManifestWithContext(context.Background(), manifestURL, opts)
}

// ProbeManifestWithContext fetches and analyzes a streaming manifest URL with context support.
// This version supports cancellation and timeout through the context parameter.
func ProbeManifestWithContext(ctx context.Context, manifestURL string, opts *ProbeOptions) (*Output, error) {
	start := time.Now()
	
	logInfo(ctx, "Starting manifest probe", map[string]interface{}{
		"url": manifestURL,
	})

	// Validate URL
	parsedURL, err := validateURL(manifestURL)
	if err != nil {
		logError(ctx, "URL validation failed", map[string]interface{}{
			"url": manifestURL,
			"error": err.Error(),
		})
		return nil, err
	}

	// Validate options
	if err := validateProbeOptions(opts); err != nil {
		logError(ctx, "Options validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	// Create HTTP client
	httpClient, err := NewHTTPClient(parsedURL.String(), opts)
	if err != nil {
		logError(ctx, "HTTP client creation failed", map[string]interface{}{
			"url": parsedURL.String(),
			"error": err.Error(),
		})
		return nil, err
	}

	// Fetch manifest content
	fetchStart := time.Now()
	body, err := httpClient.FetchManifest(parsedURL.String())
	if err != nil {
		logError(ctx, "Manifest fetch failed", map[string]interface{}{
			"url": parsedURL.String(),
			"duration": time.Since(fetchStart),
			"error": err.Error(),
		})
		return nil, err
	}

	logDebug(ctx, "Manifest fetched successfully", map[string]interface{}{
		"url": parsedURL.String(),
		"size": len(body),
		"fetch_duration": time.Since(fetchStart),
	})

	// Validate manifest content
	if len(body) == 0 {
		err := NewParsingError(parsedURL.String(), "unknown", fmt.Errorf("empty manifest content"))
		logError(ctx, "Empty manifest content", map[string]interface{}{
			"url": parsedURL.String(),
		})
		return nil, err
	}

	if len(body) > 50*1024*1024 { // 50MB limit
		err := NewParsingError(parsedURL.String(), "unknown", fmt.Errorf("manifest too large (%d bytes)", len(body)))
		logError(ctx, "Manifest too large", map[string]interface{}{
			"url": parsedURL.String(),
			"size": len(body),
		})
		return nil, err
	}

	// Detect format and parse
	parseStart := time.Now()
	var output *Output
	if strings.Contains(body, "#EXTM3U") {
		logDebug(ctx, "Detected HLS manifest", map[string]interface{}{
			"url": parsedURL.String(),
		})
		output, err = parseHLSManifest(body, parsedURL.String())
	} else {
		logDebug(ctx, "Detected MPD manifest", map[string]interface{}{
			"url": parsedURL.String(),
		})
		output, err = parseMPDManifest(body, parsedURL.String())
	}

	if err != nil {
		logError(ctx, "Manifest parsing failed", map[string]interface{}{
			"url": parsedURL.String(),
			"parse_duration": time.Since(parseStart),
			"error": err.Error(),
		})
		return nil, err
	}

	totalDuration := time.Since(start)
	logInfo(ctx, "Manifest probe completed successfully", map[string]interface{}{
		"url": parsedURL.String(),
		"streams_found": len(output.Streams),
		"total_duration": totalDuration,
		"fetch_duration": time.Since(fetchStart),
		"parse_duration": time.Since(parseStart),
	})

	return output, nil
}

// OutputJSON marshals the output to formatted JSON.
// Returns JSON bytes compatible with ffprobe output format.
func (o *Output) OutputJSON() ([]byte, error) {
	return json.MarshalIndent(o, "", "    ")
}