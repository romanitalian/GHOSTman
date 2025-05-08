.PHONY: run build clean deps test

# Default target
all: build

# Run the application
run:
	go run main.go

# Build the application
build:
	go build -o demo-form main.go

# Build for different platforms
build-all: build-darwin build-linux build-windows

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o build/demo-form-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o build/demo-form-darwin-arm64 main.go

build-linux:
	GOOS=linux GOARCH=amd64 go build -o build/demo-form-linux-amd64 main.go
	GOOS=linux GOARCH=arm64 go build -o build/demo-form-linux-arm64 main.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -o build/demo-form-windows-amd64.exe main.go

# Clean build artifacts
clean:
	rm -rf build/
	rm -f demo-form
	rm -f demo-form.exe

# Download and update dependencies
deps:
	go mod download
	go mod tidy

# Run tests
test:
	go test -v ./...

# Create build directory
init:
	mkdir -p build

# Help command
help:
	@echo "Available commands:"
	@echo "  make run          - Run the application"
	@echo "  make build        - Build for current platform"
	@echo "  make build-all    - Build for all platforms"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make deps         - Download and update dependencies"
	@echo "  make test         - Run tests"
	@echo "  make init         - Create build directory"
	@echo "  make help         - Show this help message" 