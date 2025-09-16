package probe

import "strings"

// parseVideoCodec determines video codec from codec string
func parseVideoCodec(codecString string) string {
	if strings.Contains(codecString, "avc1") {
		return "h264"
	}
	if strings.Contains(codecString, "hev1") || strings.Contains(codecString, "hvc1") {
		return "hevc"
	}
	if strings.Contains(codecString, "vp09") {
		return "vp9"
	}
	if strings.Contains(codecString, "av01") {
		return "av1"
	}
	return "h264" // default
}

// parseAudioCodec determines audio codec from codec string
func parseAudioCodec(codecString string) string {
	if strings.Contains(codecString, "ec-3") {
		return "eac3"
	}
	if strings.Contains(codecString, "mp4a") {
		return "aac"
	}
	return "aac" // default
}

// getPixelFormat determines pixel format based on codec profile information
func getPixelFormat(codecString string, videoCodec string) string {
	// Parse codec profile information for pixel format
	if strings.Contains(codecString, "avc1") {
		// H.264 codec profiles
		if strings.Contains(codecString, "avc1.640028") || strings.Contains(codecString, "avc1.640032") {
			return "yuv420p10le" // High 10 profile
		}
		return "yuv420p" // Most common for H.264
	}

	if strings.Contains(codecString, "hev1") || strings.Contains(codecString, "hvc1") {
		// HEVC codec profiles
		if strings.Contains(codecString, "hev1.2.4") || strings.Contains(codecString, "hvc1.2.4") {
			return "yuv420p10le" // Main 10 profile
		}
		return "yuv420p" // Main profile
	}

	if strings.Contains(codecString, "vp09") {
		// VP9 codec
		if strings.Contains(codecString, "vp09.02") {
			return "yuv420p10le" // Profile 2
		}
		return "yuv420p" // Profile 0
	}

	if strings.Contains(codecString, "av01") {
		// AV1 codec
		return "yuv420p" // Most common
	}

	// Default based on codec
	switch videoCodec {
	case "hevc":
		return "yuv420p"
	case "vp9":
		return "yuv420p"
	case "av1":
		return "yuv420p"
	default:
		return "yuv420p" // H.264 default
	}
}