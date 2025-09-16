package probe

import "testing"

func TestParseVideoCodec(t *testing.T) {
	tests := []struct {
		name        string
		codecString string
		expected    string
	}{
		{
			name:        "H.264 AVC1",
			codecString: "avc1.640028",
			expected:    "h264",
		},
		{
			name:        "HEVC HEV1",
			codecString: "hev1.2.4.L120.B0",
			expected:    "hevc",
		},
		{
			name:        "HEVC HVC1",
			codecString: "hvc1.2.4.L120.B0",
			expected:    "hevc",
		},
		{
			name:        "VP9",
			codecString: "vp09.00.10.08",
			expected:    "vp9",
		},
		{
			name:        "AV1",
			codecString: "av01.0.04M.08",
			expected:    "av1",
		},
		{
			name:        "Unknown codec",
			codecString: "unknown.codec",
			expected:    "h264", // default
		},
		{
			name:        "Empty codec string",
			codecString: "",
			expected:    "h264", // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseVideoCodec(tt.codecString)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestParseAudioCodec(t *testing.T) {
	tests := []struct {
		name        string
		codecString string
		expected    string
	}{
		{
			name:        "AAC MP4A",
			codecString: "mp4a.40.2",
			expected:    "aac",
		},
		{
			name:        "E-AC-3",
			codecString: "ec-3",
			expected:    "eac3",
		},
		{
			name:        "Unknown codec",
			codecString: "unknown.codec",
			expected:    "aac", // default
		},
		{
			name:        "Empty codec string",
			codecString: "",
			expected:    "aac", // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAudioCodec(tt.codecString)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetPixelFormat(t *testing.T) {
	tests := []struct {
		name        string
		codecString string
		videoCodec  string
		expected    string
	}{
		{
			name:        "H.264 standard profile",
			codecString: "avc1.640020",
			videoCodec:  "h264",
			expected:    "yuv420p",
		},
		{
			name:        "H.264 High 10 profile",
			codecString: "avc1.640028",
			videoCodec:  "h264",
			expected:    "yuv420p10le",
		},
		{
			name:        "HEVC Main profile",
			codecString: "hev1.1.6.L120.B0",
			videoCodec:  "hevc",
			expected:    "yuv420p",
		},
		{
			name:        "HEVC Main 10 profile",
			codecString: "hev1.2.4.L120.B0",
			videoCodec:  "hevc",
			expected:    "yuv420p10le",
		},
		{
			name:        "VP9 Profile 0",
			codecString: "vp09.00.10.08",
			videoCodec:  "vp9",
			expected:    "yuv420p",
		},
		{
			name:        "VP9 Profile 2",
			codecString: "vp09.02.10.10",
			videoCodec:  "vp9",
			expected:    "yuv420p10le",
		},
		{
			name:        "AV1",
			codecString: "av01.0.04M.08",
			videoCodec:  "av1",
			expected:    "yuv420p",
		},
		{
			name:        "Unknown codec",
			codecString: "unknown",
			videoCodec:  "unknown",
			expected:    "yuv420p", // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPixelFormat(tt.codecString, tt.videoCodec)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}