# gRPC Setup Documentation

## Overview

This document describes the gRPC setup for the Librus project, including protobuf generation and attachments handling.

## Protobuf Generation

### Prerequisites

- `protoc` (Protocol Buffers compiler)
- `protoc-gen-go` (Go protobuf generator)
- `protoc-gen-go-grpc` (Go gRPC generator)

### Installation

```bash
# Install protoc (macOS)
brew install protobuf

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Generation

```bash
# Using Makefile
make proto-gen

# Using script
./scripts/generate_proto.sh

# Manual generation
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/librus_scraper.proto
```

### Generated Files

- `proto/librus_scraper.pb.go` - Protobuf message definitions
- `proto/librus_scraper_grpc.pb.go` - gRPC client and server interfaces

## Attachments Handling

### Directory Structure

```
playwright/
├── attachments/           # Shared attachments directory
│   ├── .gitignore        # Ignore attachment files
│   └── {uuid}/           # Individual message attachments
│       ├── file1.pdf
│       └── file2.jpg
```

### Docker Volumes

Both services share the same attachments directory:

```yaml
# docker-compose.yaml
services:
  bot:
    volumes:
      - ./playwright/attachments:/app/attachments
  
  librus-scraper:
    volumes:
      - ./playwright/attachments:/app/attachments
```

### Configuration

The `pkg/config` package handles paths automatically:

- **Docker**: `/app/attachments`
- **Local**: `./playwright/attachments`

### Environment Variables

- `ATTACHMENTS_BASE_DIR` - Override default attachments directory
- `GRPC_SCRAPER_ADDRESS` - gRPC scraper service address (default: localhost:50051)

## Workflow

1. **gRPC Scraper Service** downloads attachments to `{ATTACHMENTS_BASE_DIR}/{uuid}/`
2. **Go Bot Service** receives message with `attachments_dir` field containing the UUID
3. **Go Bot Service** reads files from `{ATTACHMENTS_BASE_DIR}/{uuid}/` and sends to Telegram
4. **Go Bot Service** cleans up the directory after sending

## Testing

```bash
# Test protobuf generation
go test -v proto_test.go

# Test configuration
go test ./pkg/config -v

# Build project
go build .
```
