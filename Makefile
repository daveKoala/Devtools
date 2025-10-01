GO ?= go
BINARY_NAME ?= devtools
TARGET_OS ?= $(shell $(GO) env GOOS)
TARGET_ARCH ?= $(shell $(GO) env GOARCH)
BUILD_DIR ?= bin/$(TARGET_OS)_$(TARGET_ARCH)
PACKAGE_NAME ?= $(BINARY_NAME)-$(TARGET_OS)-$(TARGET_ARCH).zip

COLOR_TITLE := \033[1;36m
COLOR_CMD := \033[1;33m
COLOR_DESC := \033[0;37m
NO_COLOR := \033[0m

.PHONY: build run test tidy clean package help

help: ## Show available make targets
	@printf '\n'
	@printf '%s\n' '------------------------------------------------'
	@printf "$(COLOR_TITLE)DevTools Utility Targets$(NO_COLOR)\n"
	@printf "$(COLOR_DESC)Usage: make <target> [TARGET_OS=… TARGET_ARCH=…]$(NO_COLOR)\n\n"
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*##"} {printf "  $(COLOR_CMD)%-12s$(NO_COLOR) $(COLOR_DESC)%s$(NO_COLOR)\n", $$1, $$2}'

build: ## Compile the devtools binary into $(BUILD_DIR)/
	@mkdir -p $(BUILD_DIR)
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) $(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) .

run: ## Run the devtools menu using go run
	$(GO) run .

test: ## Execute go test for all packages
	$(GO) test ./...

tidy: ## Update go.mod/go.sum with go mod tidy
	$(GO) mod tidy

clean: ## Remove build artifacts
	rm -rf bin $(PACKAGE_NAME)

# make package TARGET_OS=darwin TARGET_ARCH=amd64
package: build ## Build and zip the binary for distribution
	cd $(BUILD_DIR) && zip -r ../$(PACKAGE_NAME) $(BINARY_NAME)
