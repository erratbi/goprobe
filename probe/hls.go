package probe

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// parseHLSManifest parses an HLS M3U8 manifest and returns stream information
func parseHLSManifest(content string) (*Output, error) {
	var streams []StreamInfo
	streamIndex := 0

	lines := strings.Split(content, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			// Parse stream info line
			bandwidth := extractHLSParam(line, "BANDWIDTH")
			resolution := extractHLSParam(line, "RESOLUTION")
			frameRate := extractHLSParam(line, "FRAME-RATE")
			codecs := extractHLSParam(line, "CODECS")

			// Extract video and audio codecs
			videoCodec, audioCodec := parseHLSCodecs(codecs)

			// Add video stream
			if resolution != "" {
				videoStream := createHLSVideoStream(streamIndex, videoCodec, resolution, frameRate, bandwidth, codecs)
				streams = append(streams, videoStream)
				streamIndex++
			}

			// Add audio stream
			audioStream := createHLSAudioStream(streamIndex, audioCodec)
			streams = append(streams, audioStream)
			streamIndex++
		}
	}

	return &Output{Streams: streams}, nil
}

func createHLSVideoStream(streamIndex int, videoCodec, resolution, frameRate, bandwidth, codecs string) StreamInfo {
	bitRateKbps := ""
	if bandwidth != "" {
		if br, err := strconv.Atoi(bandwidth); err == nil {
			bitRateKbps = fmt.Sprintf("%d kb/s", br/1000)
		}
	}

	frameRateFormatted := frameRate
	if frameRateFormatted == "" {
		frameRateFormatted = "30"
	}

	pixFmt := getPixelFormat(codecs, videoCodec)

	return StreamInfo{
		StreamID:   fmt.Sprintf("0:%d", streamIndex),
		Type:       "Video",
		Codec:      videoCodec,
		PixFmt:     pixFmt,
		Resolution: resolution,
		FrameRate:  frameRateFormatted,
		BitRate:    bitRateKbps,
	}
}

func createHLSAudioStream(streamIndex int, audioCodec string) StreamInfo {
	return StreamInfo{
		StreamID:   fmt.Sprintf("0:%d", streamIndex),
		Type:       "Audio",
		Codec:      audioCodec,
		SampleRate: "48000 Hz",
		Channels:   "stereo",
		SampleFmt:  "fltp",
	}
}

func extractHLSParam(line, param string) string {
	re := regexp.MustCompile(param + `=([^,\s]+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return strings.Trim(matches[1], `"`)
	}
	return ""
}

func parseHLSCodecs(codecs string) (string, string) {
	videoCodec := "h264"
	audioCodec := "aac"

	if strings.Contains(codecs, "avc1") {
		videoCodec = "h264"
	}
	if strings.Contains(codecs, "mp4a") {
		audioCodec = "aac"
	}

	return videoCodec, audioCodec
}