package probe

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

// MPD XML structures
type MPD struct {
	XMLName                xml.Name `xml:"MPD"`
	Type                   string   `xml:"type,attr"`
	AvailabilityStartTime  string   `xml:"availabilityStartTime,attr"`
	PublishTime            string   `xml:"publishTime,attr"`
	MinimumUpdatePeriod    string   `xml:"minimumUpdatePeriod,attr"`
	MinBufferTime          string   `xml:"minBufferTime,attr"`
	TimeShiftBufferDepth   string   `xml:"timeShiftBufferDepth,attr"`
	MaxSegmentDuration     string   `xml:"maxSegmentDuration,attr"`
	Periods                []Period `xml:"Period"`
}

type Period struct {
	ID             string          `xml:"id,attr"`
	Start          string          `xml:"start,attr"`
	AdaptationSets []AdaptationSet `xml:"AdaptationSet"`
}

type AdaptationSet struct {
	ID                 string             `xml:"id,attr"`
	Group              string             `xml:"group,attr"`
	MimeType           string             `xml:"mimeType,attr"`
	Lang               string             `xml:"lang,attr"`
	ContentType        string             `xml:"contentType,attr"`
	SegmentAlignment   string             `xml:"segmentAlignment,attr"`
	MaxFrameRate       string             `xml:"maxFrameRate,attr"`
	FrameRate          string             `xml:"frameRate,attr"`
	Codecs             string             `xml:"codecs,attr"`
	EssentialProperty  []EssentialProperty `xml:"EssentialProperty"`
	Representations    []Representation    `xml:"Representation"`
}

type EssentialProperty struct {
	SchemeIdUri string `xml:"schemeIdUri,attr"`
	Value       string `xml:"value,attr"`
}

type Representation struct {
	ID                 string `xml:"id,attr"`
	Bandwidth          string `xml:"bandwidth,attr"`
	Width              string `xml:"width,attr"`
	Height             string `xml:"height,attr"`
	FrameRate          string `xml:"frameRate,attr"`
	Codecs             string `xml:"codecs,attr"`
	AudioSamplingRate  string `xml:"audioSamplingRate,attr"`
	SAR                string `xml:"sar,attr"`
}

// parseMPDManifest parses an MPD manifest and returns stream information
func parseMPDManifest(content string, manifestURL string) (*Output, error) {
	var mpd MPD
	if err := xml.Unmarshal([]byte(content), &mpd); err != nil {
		return nil, NewParsingError(manifestURL, "MPD", err)
	}

	var streams []StreamInfo
	var videoStreams []StreamInfo
	var audioStreams []StreamInfo
	var subtitleStreams []StreamInfo

	for _, period := range mpd.Periods {
		for _, adaptationSet := range period.AdaptationSets {
			// Skip trick-play streams
			if isTrickModeStream(adaptationSet) {
				continue
			}

			for _, rep := range adaptationSet.Representations {
				switch {
				case isVideoStream(adaptationSet):
					stream := createVideoStream(adaptationSet, rep)
					videoStreams = append(videoStreams, stream)

				case isAudioStream(adaptationSet):
					stream := createAudioStream(adaptationSet, rep)
					audioStreams = append(audioStreams, stream)

				case isSubtitleStream(adaptationSet):
					stream := createSubtitleStream(adaptationSet, rep)
					subtitleStreams = append(subtitleStreams, stream)
				}
			}
		}
	}

	// Combine streams in ffprobe order: videos, then audio, then subtitles
	streamIndex := 0
	streams = append(streams, assignStreamIDs(videoStreams, &streamIndex)...)
	streams = append(streams, assignStreamIDs(audioStreams, &streamIndex)...)
	streams = append(streams, assignStreamIDs(subtitleStreams, &streamIndex)...)

	return &Output{Streams: streams}, nil
}

// Helper functions
func isTrickModeStream(adaptationSet AdaptationSet) bool {
	for _, prop := range adaptationSet.EssentialProperty {
		if prop.SchemeIdUri == "http://dashif.org/guidelines/trickmode" {
			return true
		}
	}
	return false
}

func isVideoStream(adaptationSet AdaptationSet) bool {
	return adaptationSet.ContentType == "video" || strings.Contains(adaptationSet.MimeType, "video")
}

func isAudioStream(adaptationSet AdaptationSet) bool {
	return adaptationSet.ContentType == "audio" || strings.Contains(adaptationSet.MimeType, "audio")
}

func isSubtitleStream(adaptationSet AdaptationSet) bool {
	return adaptationSet.ContentType == "text" || strings.Contains(adaptationSet.MimeType, "application")
}

func createVideoStream(adaptationSet AdaptationSet, rep Representation) StreamInfo {
	resolution := ""
	if rep.Width != "" && rep.Height != "" {
		resolution = rep.Width + "x" + rep.Height
	}

	frameRate := getFrameRate(rep, adaptationSet)
	codecString := getCodecString(rep, adaptationSet)
	videoCodec := parseVideoCodec(codecString)
	pixFmt := getPixelFormat(codecString, videoCodec)

	return StreamInfo{
		Type:       "Video",
		Codec:      videoCodec,
		PixFmt:     pixFmt,
		Resolution: resolution,
		FrameRate:  frameRate,
	}
}

func createAudioStream(adaptationSet AdaptationSet, rep Representation) StreamInfo {
	codecString := getCodecString(rep, adaptationSet)
	codec := parseAudioCodec(codecString)

	sampleRate := rep.AudioSamplingRate
	if sampleRate == "" {
		sampleRate = "48000"
	}
	sampleRate += " Hz"

	bitRateKbps := ""
	if rep.Bandwidth != "" {
		if br, err := strconv.Atoi(rep.Bandwidth); err == nil {
			bitRateKbps = fmt.Sprintf("%d kb/s", br/1000)
		}
	}

	return StreamInfo{
		Type:       "Audio",
		Codec:      codec,
		BitRate:    bitRateKbps,
		Channels:   "stereo",
		SampleFmt:  "fltp",
		SampleRate: sampleRate,
		Language:   adaptationSet.Lang,
	}
}

func createSubtitleStream(adaptationSet AdaptationSet, rep Representation) StreamInfo {
	codec := "stpp" // Default for DASH subtitles
	if strings.Contains(rep.Codecs, "wvtt") {
		codec = "webvtt"
	}

	bitRateKbps := ""
	if rep.Bandwidth != "" {
		if br, err := strconv.Atoi(rep.Bandwidth); err == nil {
			bitRateKbps = fmt.Sprintf("%d kb/s", br/1000)
		}
	}

	return StreamInfo{
		Type:     "Subtitle",
		Codec:    codec,
		BitRate:  bitRateKbps,
		Language: adaptationSet.Lang,
	}
}

func getFrameRate(rep Representation, adaptationSet AdaptationSet) string {
	frameRate := rep.FrameRate
	if frameRate == "" {
		if adaptationSet.FrameRate != "" {
			frameRate = adaptationSet.FrameRate
		} else if adaptationSet.MaxFrameRate != "" {
			frameRate = adaptationSet.MaxFrameRate
		} else {
			frameRate = "25" // default
		}
	}

	// Clean up frame rate: remove "/1"
	if strings.Contains(frameRate, "/") {
		parts := strings.Split(frameRate, "/")
		if len(parts) > 0 {
			frameRate = parts[0]
		}
	}

	return frameRate
}

func getCodecString(rep Representation, adaptationSet AdaptationSet) string {
	if rep.Codecs != "" {
		return rep.Codecs
	}
	return adaptationSet.Codecs
}

func assignStreamIDs(streams []StreamInfo, streamIndex *int) []StreamInfo {
	for i := range streams {
		langSuffix := ""
		if streams[i].Language != "" {
			langSuffix = fmt.Sprintf("(%s)", streams[i].Language)
		}
		streams[i].StreamID = fmt.Sprintf("0:%d%s", *streamIndex, langSuffix)
		*streamIndex++
	}
	return streams
}