#!/bin/bash

# Navigate to the test-env directory
cd "$(dirname "$0")"

# Inject the bin directory containing the mock ffmpeg into the PATH
export PATH="$(pwd)/bin:$PATH"

echo "Starting cam-recorder in test mode..."
echo "Using mock ffmpeg: $(which ffmpeg)"
echo "Using config: $(pwd)/config.json"
echo "Press Ctrl+C to stop."
echo "----------------------------------------"

# Run the cam-recorder binary from the parent directory
exec ../cam-recorder
