package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/erratbi/goprobe/probe"
)

func main() {
	var proxyURL = flag.String("proxy", "", "Proxy URL (e.g., http://proxy:8080)")
	var userAgent = flag.String("ua", "", "Custom User-Agent string")
	var timeout = flag.Int("timeout", 30, "Timeout in seconds")
	var disableCompression = flag.Bool("no-compression", false, "Disable gzip/deflate compression")
	var disableCamouflage = flag.Bool("no-camouflage", false, "Disable browser-like headers")
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] <URL>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nAnalyzes streaming manifests (DASH MPD and HLS M3U8) for stream information.\n\n")
		fmt.Fprintf(os.Stderr, "OPTIONS:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEXAMPLES:\n")
		fmt.Fprintf(os.Stderr, "  %s https://example.com/manifest.mpd\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -proxy http://proxy:8080 https://example.com/manifest.mpd\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -ua \"MyApp/1.0\" -timeout 10 https://example.com/manifest.m3u8\n", os.Args[0])
	}
	
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	manifestURL := flag.Arg(0)
	
	// Setup options
	opts := &probe.ProbeOptions{
		ProxyURL:           *proxyURL,
		UserAgent:          *userAgent,
		TimeoutSeconds:     *timeout,
		DisableCompression: *disableCompression,
		DisableCamouflage:  *disableCamouflage,
	}

	// Probe the manifest
	output, err := probe.ProbeManifest(manifestURL, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Output JSON
	jsonData, err := output.OutputJSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonData))
}