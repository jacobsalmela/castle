# Colors for terminal output
GREEN  := \033[0;32m
YELLOW := \033[0;33m
BLUE   := \033[0;34m
CYAN   := \033[0;36m
RED    := \033[0;31m
RESET  := \033[0m

# Common flags
CGO_FLAGS := CGO_CFLAGS="-Wno-deprecated-declarations"

# Binary name
BINARY := chronotome

.PHONY: help run d bin build bin-linux bin-linux-arm64 bin-darwin bin-darwin-arm64 bin-windows bin-all clean test vet lint fmt tidy deps deps-system deps-linux deps-darwin deps-windows coverage install

# Default target: run during development, otherwise, help
.DEFAULT_GOAL := run

help: ## Show this help message
	@echo "$(CYAN)Available targets:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(RESET) %s\n", $$1, $$2}'

run: ## Run the project (deletes save.json first)
	@echo "$(BLUE)Running the project...$(RESET)"
	rm -f save.json
	$(CGO_FLAGS) go run cmd/main.go

d: run ## Alias for run

# Default build (current platform)
bin: ## Build binary for current platform
	@echo "$(BLUE)Building the project...$(RESET)"
	$(CGO_FLAGS) go build -o bin/$(BINARY) cmd/main.go
	@echo "$(GREEN)Build complete: bin/$(BINARY)$(RESET)"

# Platform-specific builds
bin-linux: ## Build for Linux (amd64)
	@echo "$(BLUE)Building for Linux...$(RESET)"
	@if [ "$$(uname -s)" = "Linux" ] && [ -z "$$DISPLAY" ]; then \
		xvfb-run -a env CGO_CFLAGS="-Wno-deprecated-declarations" GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -o bin/$(BINARY)-linux-amd64 cmd/main.go; \
	else \
		$(CGO_FLAGS) GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -o bin/$(BINARY)-linux-amd64 cmd/main.go; \
	fi
	@echo "$(GREEN)Build complete: bin/$(BINARY)-linux-amd64$(RESET)"

bin-linux-arm64: ## Build for Linux (arm64)
	@echo "$(BLUE)Building for Linux ARM64...$(RESET)"
	@if [ "$$(uname -s)" = "Linux" ] && [ -z "$$DISPLAY" ]; then \
		xvfb-run -a env CGO_CFLAGS="-Wno-deprecated-declarations" GOOS=linux GOARCH=arm64 CGO_ENABLED=1 go build -ldflags "-w -s" -o bin/$(BINARY)-linux-arm64 cmd/main.go; \
	else \
		$(CGO_FLAGS) GOOS=linux GOARCH=arm64 CGO_ENABLED=1 go build -ldflags "-w -s" -o bin/$(BINARY)-linux-arm64 cmd/main.go; \
	fi
	@echo "$(GREEN)Build complete: bin/$(BINARY)-linux-arm64$(RESET)"

bin-darwin: ## Build for macOS (amd64)
	@echo "$(BLUE)Building for macOS...$(RESET)"
	$(CGO_FLAGS) GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -ldflags "-w -s" -o bin/$(BINARY)-darwin-amd64 cmd/main.go
	@echo "$(GREEN)Build complete: bin/$(BINARY)-darwin-amd64$(RESET)"

bin-darwin-arm64: ## Build for macOS (arm64/Apple Silicon)
	@echo "$(BLUE)Building for macOS ARM64 (Apple Silicon)...$(RESET)"
	$(CGO_FLAGS) GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -ldflags "-w -s" -o bin/$(BINARY)-darwin-arm64 cmd/main.go
	@echo "$(GREEN)Build complete: bin/$(BINARY)-darwin-arm64$(RESET)"

bin-windows: ## Build for Windows (amd64)
	@echo "$(BLUE)Building for Windows...$(RESET)"
	$(CGO_FLAGS) GOOS=windows GOARCH=amd64 go build -ldflags "-w -s" -o bin/$(BINARY)-windows-amd64.exe cmd/main.go
	@echo "$(GREEN)Build complete: bin/$(BINARY)-windows-amd64.exe$(RESET)"

# Build all platforms
bin-all: bin-linux bin-linux-arm64 bin-darwin bin-darwin-arm64 bin-windows ## Build for all platforms
	@echo "$(GREEN)All platform builds complete.$(RESET)"

# Remove build artifacts and generated files
clean: ## Remove build artifacts and caches
	@echo "$(YELLOW)Cleaning build artifacts...$(RESET)"
	$(CGO_FLAGS) go clean -cache -modcache -testcache
	@rm -rf bin
	@rm -f save.json
	@echo "$(GREEN)Clean complete$(RESET)"

test: ## Run all tests
	@echo "$(BLUE)Running tests...$(RESET)"
	@if [ "$$(uname -s)" = "Linux" ] && [ -z "$$DISPLAY" ]; then \
		xvfb-run -a env CGO_CFLAGS="-Wno-deprecated-declarations" go test ./...; \
	else \
		$(CGO_FLAGS) go test ./...; \
	fi
	@echo "$(GREEN)Tests complete$(RESET)"

vet: ## Run go vet for static analysis
	@echo "$(BLUE)Running go vet...$(RESET)"
	$(CGO_FLAGS) go vet ./...
	@echo "$(GREEN)Vet complete$(RESET)"

tools:
	@echo "$(BLUE)Installing development tools...$(RESET)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "$(GREEN)Tools installed$(RESET)"

mod-patch: ## Tidy and create a patch for go.mod and go.sum
	@echo "$(BLUE)Updating all go modules to latest patch/minor versions...$(RESET)"
	go get -u=patch ./...
	go mod tidy
	@echo "$(GREEN)Modules updated to latest patch/minor versions$(RESET)"

lint: ## Run golangci-lint (requires golangci-lint installed)
	@echo "$(BLUE)Running linter...$(RESET)"
	golangci-lint run ./...
	@echo "$(GREEN)Lint complete$(RESET)"

fmt: ## Format all Go files
	@echo "$(BLUE)Formatting code...$(RESET)"
	go fmt ./...
	@echo "$(GREEN)Formatted$(RESET)"

tidy: ## Tidy and verify go.mod
	@echo "$(BLUE)Tidying modules...$(RESET)"
	go mod tidy
	go mod verify
	@echo "$(GREEN)Modules tidied$(RESET)"

deps: ## Download dependencies
	@echo "$(BLUE)Downloading dependencies...$(RESET)"
	go mod download
	@echo "$(GREEN)Dependencies downloaded$(RESET)"

deps-system: ## Install platform-specific system dependencies
	@echo "$(BLUE)Detecting platform and installing system dependencies...$(RESET)"
	@if [ "$(shell uname -s)" = "Linux" ]; then \
		$(MAKE) deps-linux; \
	elif [ "$(shell uname -s)" = "Darwin" ]; then \
		$(MAKE) deps-darwin; \
	else \
		echo "$(YELLOW)No system dependencies needed for this platform$(RESET)"; \
	fi

deps-linux: ## Install Linux system dependencies (ebiten/glfw and alsa)
	@echo "$(BLUE)Installing Linux dependencies...$(RESET)"
	sudo apt-get update
	sudo apt-get install -y \
		libgl1-mesa-dev \
		libxi-dev \
		libxxf86vm-dev \
		libx11-dev \
		libxrandr-dev \
		libxinerama-dev \
		libxcursor-dev \
		libxi-dev \
		libasound2-dev \
		pkg-config \
		xvfb
	@echo "$(GREEN)Linux dependencies installed$(RESET)"

deps-darwin: ## Install macOS system dependencies
	@echo "$(BLUE)Installing macOS dependencies...$(RESET)"
	brew install glfw
	@echo "$(GREEN)macOS dependencies installed$(RESET)"

deps-windows: ## Install Windows system dependencies
	@echo "$(YELLOW)No additional system dependencies needed for Windows$(RESET)"

coverage: ## Run tests with coverage report
	@echo "$(BLUE)Running tests with coverage...$(RESET)"
	@if [ "$$(uname -s)" = "Linux" ] && [ -z "$$DISPLAY" ]; then \
		xvfb-run -a $(CGO_FLAGS) go test -coverprofile=coverage.out ./...; \
	else \
		$(CGO_FLAGS) go test -coverprofile=coverage.out ./...; \
	fi
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report: coverage.html$(RESET)"

install: bin ## Install binary to GOPATH/bin
	@echo "$(BLUE)Installing $(BINARY)...$(RESET)"
	cp bin/$(BINARY) $(GOPATH)/bin/$(BINARY)
	@echo "$(GREEN)Installed to $(GOPATH)/bin/$(BINARY)$(RESET)"

# Benchmark targets (internal)
.bench-ecs: ## Run ECS benchmark (fast sizes)
	@if [ "$$(uname -s)" = "Linux" ] && [ -z "$$DISPLAY" ]; then \
		xvfb-run -a $(CGO_FLAGS) go test ./ecs -bench BenchmarkEntitiesWithImplementations -benchmem -run ^$$; \
	else \
		$(CGO_FLAGS) go test ./ecs -bench BenchmarkEntitiesWithImplementations -benchmem -run ^$$; \
	fi

.bench-ecs-100k: ## Run ECS benchmark (N=100k)
	@if [ "$$(uname -s)" = "Linux" ] && [ -z "$$DISPLAY" ]; then \
		xvfb-run -a $(CGO_FLAGS) go test ./ecs -run TestEntitiesWithLarge100k -v; \
	else \
		$(CGO_FLAGS) go test ./ecs -run TestEntitiesWithLarge100k -v; \
	fi

.bench-ecs-500k: ## Run ECS benchmark (N=500k)
	@if [ "$$(uname -s)" = "Linux" ] && [ -z "$$DISPLAY" ]; then \
		xvfb-run -a $(CGO_FLAGS) go test ./ecs -run TestEntitiesWithLarge500k -v; \
	else \
		$(CGO_FLAGS) go test ./ecs -run TestEntitiesWithLarge500k -v; \
	fi

test_draw: ## Run draw system tests
	@echo "$(BLUE)Running draw tests...$(RESET)"
	@if [ "$$(uname -s)" = "Linux" ] && [ -z "$$DISPLAY" ]; then \
		xvfb-run -a $(CGO_FLAGS) go test ./systems/draw -v -count=1; \
		xvfb-run -a $(CGO_FLAGS) go test ./... -count=1; \
	else \
		$(CGO_FLAGS) go test ./systems/draw -v -count=1; \
		$(CGO_FLAGS) go test ./... -count=1; \
	fi
	@echo "$(GREEN)Draw tests complete$(RESET)"

test_update: ## Run update system tests
	@echo "$(BLUE)Running update tests...$(RESET)"
	@if [ "$$(uname -s)" = "Linux" ] && [ -z "$$DISPLAY" ]; then \
		xvfb-run -a $(CGO_FLAGS) go test ./systems/update -v -count=1; \
		xvfb-run -a $(CGO_FLAGS) go test ./... -count=1; \
	else \
		$(CGO_FLAGS) go test ./systems/update -v -count=1; \
		$(CGO_FLAGS) go test ./... -count=1; \
	fi
	@echo "$(GREEN)Update tests complete$(RESET)"
