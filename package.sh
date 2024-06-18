#!/bin/bash

PLATFORMS=("windows/amd64" "windows/386" "darwin/arm64" "linux/386" "linux/amd64" "linux/arm" "linux/arm64")

for PLATFORM in "${PLATFORMS[@]}"
do
    OS=$(echo $PLATFORM | cut -d '/' -f 1)
    ARCH=$(echo $PLATFORM | cut -d '/' -f 2)
    OUTPUT_NAME="curls_1.0.0_${OS}_${ARCH}"
    if [ $OS = "windows" ]; then
        OUTPUT_NAME+='.exe'
    fi

    echo "Building for $PLATFORM..."
    env GOOS=$OS GOARCH=$ARCH go build -o $OUTPUT_NAME main.go
done
