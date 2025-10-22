.PHONY: build clean test install deps

# Default target
all: deps build

# Install dependencies
deps:
	go mod tidy
	go mod download

# Build
build:
	go build -o cmplist cmplist.go

# Release build (optimized)
build-release:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o cmplist cmplist.go

# Cross compilation
build-linux:
	GOOS=linux GOARCH=amd64 go build -o cmplist-linux cmplist.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -o cmplist-windows.exe cmplist.go

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o cmplist-darwin cmplist.go

# Test
test:
	go test -v ./...

# Cleanup
clean:
	rm -f cmplist cmplist-* *.exe

# Install
install: build
	cp cmplist /usr/local/bin/

# Update dependencies for development
update-deps:
	go get -u ./...
	go mod tidy

# Format
fmt:
	go fmt ./...

# Lint
lint:
	golangci-lint run

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build binary"
	@echo "  build-release - Build optimized binary"
	@echo "  build-linux   - Build Linux binary"
	@echo "  build-windows - Build Windows binary"
	@echo "  build-darwin  - Build macOS binary"
	@echo "  deps          - Install dependencies"
	@echo "  test          - Run tests"
	@echo "  clean         - Remove generated files"
	@echo "  install       - Install to system"
	@echo "  update-deps   - Update dependencies"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter"
	@echo "  help          - Show this help"