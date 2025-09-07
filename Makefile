.PHONY: proto-gen clean-proto build test

# Generate protobuf files
proto-gen:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/librus_scraper.proto

# Clean generated protobuf files
clean-proto:
	rm -f proto/*.pb.go

# Build the application
build:
	go build -o bin/librus .

# Run tests
test:
	go test ./...

# Install protobuf tools
install-proto-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Full setup (install tools and generate proto)
setup: install-proto-tools proto-gen
