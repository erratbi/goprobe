# goprobe

A fast Go library for analyzing streaming manifests (DASH MPD and HLS M3U8). Extracts stream information including video codec, resolution, frame rate, audio codec, and more - compatible with ffprobe output format but 36x faster.

## Features

- **Fast**: Direct manifest parsing vs ffprobe's binary media analysis
- **Universal**: Supports both DASH (.mpd) and HLS (.m3u8) manifests
- **Smart**: Automatic codec detection and pixel format inference
- **Compatible**: Output format matches ffprobe JSON structure
- **Configurable**: Proxy support, custom headers, timeouts
- **Clean**: Well-structured Go package with comprehensive documentation

## Installation

```bash
go get github.com/erratbi/goprobe
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/erratbi/goprobe/probe"
)

func main() {
    // Basic usage
    output, err := probe.ProbeManifest("https://example.com/manifest.mpd", nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Print stream information
    for _, stream := range output.Streams {
        fmt.Printf("Stream %s: %s %s %s\n", 
            stream.StreamID, stream.Type, stream.Codec, stream.Resolution)
    }
    
    // Get JSON output
    jsonData, _ := output.OutputJSON()
    fmt.Println(string(jsonData))
}
```

## Advanced Usage

```go
// With custom options
opts := &probe.ProbeOptions{
    ProxyURL:       "http://proxy:8080",
    UserAgent:      "MyApp/1.0",
    TimeoutSeconds: 10,
    CustomHeaders: map[string]string{
        "Authorization": "Bearer token123",
    },
}

output, err := probe.ProbeManifest(manifestURL, opts)
```

## CLI Tool

```bash
# Basic usage
go run . https://example.com/manifest.mpd

# With proxy
go run . -proxy http://proxy:8080 https://example.com/manifest.mpd

# With custom options
go run . -proxy http://user:pass@proxy:8080 -ua "MyApp/1.0" -timeout 10 https://example.com/manifest.m3u8

# All options
go run . -h

# Output (JSON)
{
    "streams": [
        {
            "stream_id": "0:0",
            "type": "Video",
            "codec": "h264",
            "pix_fmt": "yuv420p",
            "resolution": "1920x1080",
            "frame_rate": "25"
        },
        {
            "stream_id": "0:1(eng)",
            "type": "Audio",
            "codec": "aac",
            "channels": "stereo",
            "sample_rate": "48000 Hz",
            "language": "eng"
        }
    ]
}
```

## Performance

- **ffprobe**: ~9 seconds (full media analysis)
- **goprobe**: ~0.25 seconds (manifest parsing only)
- **Speedup**: ~36x faster

## Supported Formats

### DASH (MPD)
- Video codecs: H.264, HEVC, VP9, AV1
- Audio codecs: AAC, E-AC-3
- Subtitle formats: STPP, WebVTT
- Pixel formats: Automatic detection based on codec profiles
- DRM: Detection of encrypted streams

### HLS (M3U8)
- Video codecs: H.264, HEVC
- Audio codecs: AAC
- Adaptive bitrate streams
- Multiple quality levels

## API Reference

### Types

```go
type StreamInfo struct {
    StreamID   string `json:"stream_id"`
    Type       string `json:"type"`        // Video, Audio, Subtitle
    Codec      string `json:"codec"`       // h264, hevc, aac, etc.
    PixFmt     string `json:"pix_fmt"`     // yuv420p, yuv420p10le, etc.
    Resolution string `json:"resolution"`  // 1920x1080, etc.
    FrameRate  string `json:"frame_rate"`  // 25, 30, 50, etc.
    BitRate    string `json:"bit_rate"`    // 3000 kb/s, etc.
    Language   string `json:"language"`    // eng, fra, etc.
    // ... more fields
}

type ProbeOptions struct {
    ProxyURL           string
    UserAgent          string
    CustomHeaders      map[string]string
    TimeoutSeconds     int
    DisableCompression bool
    DisableCamouflage  bool
}
```

### Functions

```go
// Main function - analyzes manifest URL
func ProbeManifest(manifestURL string, opts *ProbeOptions) (*Output, error)

// Convert output to JSON
func (o *Output) OutputJSON() ([]byte, error)
```

## License

MIT License - see LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Submit a pull request

## Acknowledgments

Built for fast streaming manifest analysis, inspired by ffprobe but optimized for speed.