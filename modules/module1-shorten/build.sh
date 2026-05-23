#!/bin/bash
set -e

echo "Building module1-shorten..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
zip function.zip bootstrap
rm bootstrap
echo "Build complete: function.zip created"
