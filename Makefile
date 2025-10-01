GO ?= go
BINARY_NAME ?= devtools
TARGET_OS ?= $(shell $(GO) env GOOS)
TARGET_ARCH ?= $(shell $(GO) env GOARCH)
BUILD_DIR ?= bin/$(TARGET_OS)_$(TARGET_ARCH)
PACKAGE_NAME ?= $(BINARY_NAME)-$(TARGET_OS)-$(TARGET_ARCH).zip
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.buildDate=$(BUILD_DATE)'

COLOR_TITLE := \033[1;36m
COLOR_CMD := \033[1;33m
COLOR_DESC := \033[0;37m
NO_COLOR := \033[0m

.PHONY: build run test tidy clean package package-macos help

help: ## Show available make targets
	@printf '\n'
	@printf '%s\n' '------------------------------------------------'
	@printf "$(COLOR_TITLE)DevTools Utility Targets$(NO_COLOR)\n"
	@printf "$(COLOR_DESC)Usage: make <target> [TARGET_OS=… TARGET_ARCH=…]$(NO_COLOR)\n\n"
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*##"} {printf "  $(COLOR_CMD)%-12s$(NO_COLOR) $(COLOR_DESC)%s$(NO_COLOR)\n", $$1, $$2}'

build: ## Compile the devtools binary into $(BUILD_DIR)/
	@mkdir -p $(BUILD_DIR)
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) $(GO) build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) .

run: ## Run the devtools menu using go run
	$(GO) run -ldflags "$(LDFLAGS)" .

test: ## Execute go test for all packages
	$(GO) test ./...

tidy: ## Update go.mod/go.sum with go mod tidy
	$(GO) mod tidy

clean: ## Remove build artifacts
	rm -rf bin $(PACKAGE_NAME)

package: build ## Build and zip the binary for distribution
	cd $(BUILD_DIR) && zip -r ../$(PACKAGE_NAME) $(BINARY_NAME)

package-macos: ## Build macOS arm64 and amd64 packages
	$(MAKE) package TARGET_OS=darwin TARGET_ARCH=arm64
	$(MAKE) package TARGET_OS=darwin TARGET_ARCH=amd64
