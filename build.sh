#!/bin/bash

# Set Go project name (change this as needed)
PROJECT_NAME="HypTool"

# Output directory for builds
OUTPUT_DIR="./builds"
mkdir -p "$OUTPUT_DIR"

# List of OS and architectures to build for
OS_ARCHS=(
    "linux amd64"
    "linux arm64"
    "linux 386"
    "linux arm"
    "darwin amd64"
    "darwin arm64"
    "windows amd64"
    "windows arm64"
    "windows 386"
    "freebsd amd64"
    "freebsd arm64"
    "openbsd amd64"
    "openbsd arm64"
    "netbsd amd64"
    "netbsd arm64"
)

echo "Starting Go cross-compilation..."

for TARGET in "${OS_ARCHS[@]}"; do
    read -r GOOS GOARCH <<< "$TARGET"

    # Set output binary name
    OUTPUT_NAME="${PROJECT_NAME}-${GOOS}-${GOARCH}"
    if [ "$GOOS" == "windows" ]; then
        OUTPUT_NAME+=".exe"
    fi

    echo "Building for $GOOS/$GOARCH..."
    
    env GOOS="$GOOS" GOARCH="$GOARCH" go build -o "$OUTPUT_DIR/$OUTPUT_NAME" .
    
    if [ $? -eq 0 ]; then
        echo "Successfully built: $OUTPUT_NAME"
    else
        echo "Failed to build: $OUTPUT_NAME"
    fi
done

echo "Build process completed!"
