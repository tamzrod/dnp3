# Makefile for DNP3 Engineering Workbench
# Supports Linux development and Windows cross-compilation

.PHONY: all build test clean run windows linux deps fmt lint

# Default target
all: build

# Build for current OS
build:
	go build -ldflags="-s -w" -o workbench .

# Build for Linux
linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-s -w" -o workbench-linux .

# Build for Windows (cross-compile from Linux)
# Requires: mingw-w64 (sudo apt install mingw-w64)
windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
		go build -ldflags="-s -w" -o workbench.exe .

# Build for Windows (from Windows or with cross-compiler)
windows-native:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-s -w" -o workbench.exe .

# Run locally
run:
	go run .

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f workbench workbench.exe workbench-linux coverage.out coverage.html

# Download and tidy dependencies
deps:
	go mod tidy
	go mod download

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run ./...

# Verify dependencies
verify:
	go mod verify

# Show help
help:
	@echo "DNP3 Engineering Workbench Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  all       Build for current OS (default)"
	@echo "  build     Build for current OS"
	@echo "  linux     Build for Linux"
	@echo "  windows   Cross-compile for Windows (from Linux)"
	@echo "  run       Run locally with 'go run'"
	@echo "  test      Run tests"
	@echo "  test-cover Run tests with coverage report"
	@echo "  clean     Remove build artifacts"
	@echo "  deps      Download and tidy dependencies"
	@echo "  fmt       Format code"
	@echo "  lint      Run linter"
	@echo "  verify    Verify dependencies"
	@echo "  help      Show this help"
