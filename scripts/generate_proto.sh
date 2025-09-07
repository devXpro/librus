#!/bin/bash

# Script to generate protobuf files for librus project

set -e

echo "🔧 Generating protobuf files..."

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "❌ protoc is not installed. Please install it first:"
    echo "   brew install protobuf"
    exit 1
fi

# Check if Go protobuf plugins are installed
if ! command -v protoc-gen-go &> /dev/null; then
    echo "❌ protoc-gen-go is not installed. Installing..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "❌ protoc-gen-go-grpc is not installed. Installing..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# Clean old generated files
echo "🧹 Cleaning old generated files..."
rm -f proto/*.pb.go

# Generate new files
echo "⚡ Generating new protobuf files..."
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/librus_scraper.proto

echo "✅ Protobuf files generated successfully!"
echo "📁 Generated files:"
ls -la proto/*.pb.go

# Run go mod tidy to ensure dependencies are up to date
echo "📦 Running go mod tidy..."
go mod tidy

echo "🎉 All done!"
