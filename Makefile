.DEFAULT_GOAL := help
APP_NAME := GHOSTman

.PHONY: help
help: ## Available commands
	@clear
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", 622821, 622822 } /^##@/ { printf "\n\033[0;33m%s\033[0m\n", substr(622820, 5) } ' 
	@echo ""

##@ Targets

.PHONY: run
run: ## Run the application
	go run main.go

.PHONY: build
build: ## Build the application
	go build -o GHOSTman main.go

.PHONY: build-all
build-all: ## Build for different platforms
	make build-darwin
	make build-linux
	make build-windows

.PHONY: build-darwin
build-darwin: ## Build for darwin
	GOOS=darwin GOARCH=amd64 go build -o build/GHOSTman-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o build/GHOSTman-darwin-arm64 main.go

.PHONY: build-linux
build-linux: ## Build for linux
	GOOS=linux GOARCH=amd64 go build -o build/GHOSTman-linux-amd64 main.go
	GOOS=linux GOARCH=arm64 go build -o build/GHOSTman-linux-arm64 main.go

.PHONY: build-windows
build-windows: ## Build for windows
	GOOS=windows GOARCH=amd64 go build -o build/GHOSTman-windows-amd64.exe main.go

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf build/
	rm -f GHOSTman
	rm -f GHOSTman.exe

.PHONY: deps
deps: ## Download and update dependencies
	go mod download
	go mod tidy

.PHONY: test
test: ## Run tests
	go test -v ./...

##@ Aliases

.PHONY: r
r: ## run App (command alias)
	@make run

.PHONY: b
b: ## build App (command alias)
	@make build

.PHONY: ba
ba: ## build all App (command alias)
	@make build-all

.PHONY: bd
bd: ## build darwin App (command alias)
	@make build-darwin

.PHONY: bl
bl: ## build linux App (command alias)
	@make build-linux

.PHONY: bw
bw: ## build windows App (command alias)
	@make build-windows

.PHONY: t
t: ## test App (command alias)
	@make test

.PHONY: c
c: ## clean App (command alias)
	@make clean
