APP := gdss
TEST := gdss_test
BINARY_NAME := $(APP)
BIN_DIR := bin
GO_FILES := $(wildcard *.go)
VERSION_FILE := version.txt
VERSION ?= $(shell cat $(VERSION_FILE) 2> /dev/null || echo "latest")

.PHONY: all build run test clean gotool fmt vet lint docker docker-build docker-push help modules version release \
        build-linux build-windows

all: build

build: modules $(BIN_DIR)/$(BINARY_NAME)

$(BIN_DIR)/$(BINARY_NAME): $(GO_FILES)
	@mkdir -p $(BIN_DIR)
	@echo "Building $(BINARY_NAME) with version $(VERSION)..."
	@VERSION=$(VERSION) go build -ldflags "-X main.Version=$(VERSION)" -o $@ $^ || exit 1

run: build
	@echo "Running $(BINARY_NAME) with version $(VERSION)..."
	@VERSION=$(VERSION) go run -race main.go || exit 1


test:
	@echo "Running tests..."
	@go test ./... -v || exit 1


gotool: fmt vet

fmt:
	@echo "Formatting Go code..."
	@go fmt ./... || exit 1

vet:
	@echo "Vetting Go code..."
	@go vet ./... || exit 1

lint:
	@echo "Running Go linters (if installed)..."
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not found. Install it if you want to run linters."


modules:
	@echo "Tidying Go modules..."
	@go mod tidy || exit 1


docker: docker-build

docker-build: build
	@echo "Building Docker image $(APP):$(VERSION)..."
	@docker build --build-arg VERSION=$(VERSION) -t $(APP):$(VERSION) . || exit 1

docker-rm:
	@echo "Removing Docker image $(APP):$(VERSION)..."
	@docker rmi $(APP):$(VERSION) || true
	@docker rmi $(APP):latest || true

docker-push:
	@echo "Pushing Docker image $(APP):$(VERSION) to remote..."
	@docker push $(APP):$(VERSION) || exit 1
	@docker push $(APP):latest || exit 1


version:
	@echo "Current version: $(VERSION)"

release: test lint build docker
	@echo "Release $(VERSION) completed."


build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION)" -o $(BIN_DIR)/$(BINARY_NAME)-linux .

build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION)" -o $(BIN_DIR)/$(BINARY_NAME).exe .

clean: 
	@echo "Cleaning build artifacts..."
	@go clean
	@rm -rf $(BIN_DIR) || true
	@rm -rf $(TEST) || true
	@rm -rf $(APP) || true

rm: clean
	@echo "Removing Docker image $(APP):$(VERSION)..."
	@docker rmi $(APP):$(VERSION) || true
	@docker rmi $(APP):latest || true

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@echo "  all            - Builds the application (default)"
	@echo "  build          - Builds the Go code into ./$(BIN_DIR)/$(BINARY_NAME)"
	@echo "  run            - Runs the Go code with race detector"
	@echo "  test           - Runs all Go tests"
	@echo "  clean          - Cleans build output and artifacts"
	@echo "  gotool         - Runs formatting and vetting"
	@echo "    fmt          - Formats code with gofmt"
	@echo "    vet          - Vets code with go vet"
	@echo "  lint           - Runs golangci-lint (if installed)"
	@echo "  modules        - Runs go mod tidy"
	@echo "  version        - Prints current version"
	@echo "  docker         - Builds Docker image (alias for docker-build)"
	@echo "  docker-build   - Builds Docker image with version tag"
	@echo "  docker-push    - Pushes Docker image to remote registry"
	@echo "  release        - Run test, lint, build, and push docker image"
	@echo "  build-linux    - Cross-compiles for Linux amd64"
	@echo "  build-windows  - Cross-compiles for Windows amd64"
