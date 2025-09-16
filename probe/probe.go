// Package probe provides fast manifest parsing for DASH and HLS streams.
// It analyzes MPD and M3U8 manifests to extract stream information including
// video codec, resolution, frame rate, audio codec, and more.
package probe

import (
	"encoding/json"
	"strings"
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
	// Create HTTP client
	httpClient, err := NewHTTPClient(manifestURL, opts)
	if err != nil {
		return nil, err
	}

	// Fetch manifest content
	body, err := httpClient.FetchManifest(manifestURL)
	if err != nil {
		return nil, err
	}

	// Detect format and parse
	if strings.Contains(body, "#EXTM3U") {
		return parseHLSManifest(body)
	}
	return parseMPDManifest(body)
}

// OutputJSON marshals the output to formatted JSON.
// Returns JSON bytes compatible with ffprobe output format.
func (o *Output) OutputJSON() ([]byte, error) {
	return json.MarshalIndent(o, "", "    ")
}