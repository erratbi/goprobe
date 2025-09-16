package probe

import (
	"fmt"
	"net/url"
	"time"

	"github.com/imroc/req/v3"
)

// HTTPClient wraps the req client with manifest-specific configuration
type HTTPClient struct {
	client *req.Client
}

// NewHTTPClient creates a new HTTP client configured for manifest fetching
func NewHTTPClient(targetURL string, opts *ProbeOptions) (*HTTPClient, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing URL: %w", err)
	}

	client := createConfiguredClient(parsedURL, opts)
	
	return &HTTPClient{client: client}, nil
}

// FetchManifest fetches the manifest content from the given URL
func (h *HTTPClient) FetchManifest(manifestURL string) (string, error) {
	resp, err := h.client.R().Get(manifestURL)
	if err != nil {
		return "", fmt.Errorf("error fetching manifest: %w", err)
	}

	return resp.String(), nil
}

// createConfiguredClient creates a req client with all necessary headers and settings
func createConfiguredClient(parsedURL *url.URL, opts *ProbeOptions) *req.Client {
	// Set defaults
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	timeoutSeconds := 30
	
	if opts != nil {
		if opts.UserAgent != "" {
			userAgent = opts.UserAgent
		}
		if opts.TimeoutSeconds > 0 {
			timeoutSeconds = opts.TimeoutSeconds
		}
	}

	client := req.C().
		SetUserAgent(userAgent).
		SetTimeout(time.Duration(timeoutSeconds) * time.Second).
		EnableAutoReadResponse()

	// Configure compression
	if opts == nil || !opts.DisableCompression {
		client.EnableCompression()
	}

	// Configure camouflage headers
	if opts == nil || !opts.DisableCamouflage {
		origin := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
		referer := origin + "/"
		
		client.SetCommonHeaders(map[string]string{
			"Accept":          "application/dash+xml,application/vnd.ms-sstr+xml,application/vnd.apple.mpegurl,application/x-mpegURL,application/vnd.ms-playready.media.pya,application/vnd.ms-playready.media.pyv,video/mp4,audio/mp4,*/*",
			"Accept-Language": "en-US,en;q=0.9,fr;q=0.8",
			"Origin":          origin,
			"Referer":         referer,
			"DNT":             "1",
			"Connection":      "keep-alive",
			"Upgrade-Insecure-Requests": "1",
		})
	}

	// Add custom headers
	if opts != nil && len(opts.CustomHeaders) > 0 {
		client.SetCommonHeaders(opts.CustomHeaders)
	}

	// Configure proxy
	if opts != nil && opts.ProxyURL != "" {
		client.SetProxyURL(opts.ProxyURL)
	}

	return client
}