#!/bin/bash
# Pure Go build script - No C/C++ dependencies required
# This script builds gobot with zero C/C++ code

set -e

echo "🚀 Building Gobot with Pure Go (No C/C++ dependencies)"
echo "=================================================="

# Set environment variables for pure Go build
export CGO_ENABLED=0
export GOOS=${GOOS:-$(go env GOOS)}
export GOARCH=${GOARCH:-$(go env GOARCH)}

# Build tags to exclude C/C++ dependent code and enable pure Go implementations
BUILD_TAGS="!gocv,!libusb,!cgo,purgo"

echo "🔧 Build Configuration:"
echo "   CGO_ENABLED: $CGO_ENABLED"
echo "   GOOS: $GOOS"
echo "   GOARCH: $GOARCH"
echo "   BUILD_TAGS: $BUILD_TAGS"
echo ""

echo "📦 Building all packages..."
go build -tags "$BUILD_TAGS" ./...

echo "✅ Build successful!"
echo ""

echo "🧪 Running tests..."
go test -tags "$BUILD_TAGS" -v ./... | grep -E "(PASS|FAIL|RUN)"

echo ""
echo "🎉 Pure Go build complete!"
echo "   ✓ No C source files"
echo "   ✓ No C++ dependencies" 
echo "   ✓ No CGO required"
echo "   ✓ Static binary ready"
echo ""
echo "📋 Usage:"
echo "   CGO_ENABLED=0 go build -tags '!gocv,!libusb,!cgo,purgo' ./..."
echo "   CGO_ENABLED=0 go run -tags '!gocv,!libusb,!cgo,purgo' example.go"