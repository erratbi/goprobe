#!/bin/bash

# MPD URL with decryption keys
URL="https://dsh-m006.p-cdnlive-edge020304-dual.scy.canalplus-cdn.net/__token__id=1a569a48e47690c700c0f71662c87679~hmac=6d2114baa9bb694f530073a9f707530612577a89d4d0a36aaa72374fa9e12049/live/disk/beinsports1-hd/dash-fhd/beinsports1-hd.mpd?decryption_key=ae9080c611bc471b954fc73379dfdb71:5879516e23e11038350d7a9751eec25c,bbad6d5e337d4a3ca82417875562db85:45bab48bed539b8f44f141b1c8126847,644e4c24c7be4ee5b09c4446d0e1062e:c68c6e85ab92749d57adac6b76d9fab9,21f8ff631609426abeb0c58bbc00d771:98f0c9113d3f171570e39622b8dda74f"

echo "=== FFprobe MPD Analysis ==="
echo

echo "1. Basic JSON output (format + streams):"
echo "----------------------------------------"
time ffprobe -v quiet -print_format json -show_format -show_streams -analyzeduration 10000 -probesize 10000 -fflags +nobuffer+fastseek -avioflags direct -max_analyze_duration 10000 -fps_mode passthrough "$URL" > basic_probe.json 2>&1
if [ $? -eq 0 ]; then
    echo "✓ Success - Output saved to basic_probe.json"
    jq '.format.format_name, .streams | length' basic_probe.json 2>/dev/null || echo "jq not available for pretty formatting"
else
    echo "✗ Failed - Check basic_probe.json for errors"
fi


  