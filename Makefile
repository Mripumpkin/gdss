APP := gdss
BINARY_NAME := $(APP)
GO_FILES := $(wildcard *.go)

.PHONY: all build run test clean gotool help

all: build

build: $(BINARY_NAME)

$(BINARY_NAME): $(GO_FILES)
	@go build -o $@ $^

run:
	@go run -race main.go

test:
	@echo "Running tests..."
	@go test ./... -v

clean:
	@echo "Cleaning build artifacts..."
	@go clean
	@rm -f $(BINARY_NAME) # 显式删除二进制文件

gotool:
	@echo "Running Go formatting and vetting tools..."
	@go fmt ./...
	@go vet ./...

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@echo "  all      - Builds the application (default)"
	@echo "  build    - Builds the Go code, generating the binary '$(BINARY_NAME)'"
	@echo "  run      - Runs the Go code directly with race detection"
	@echo "  test     - Runs all Go tests in the project"
	@echo "  clean    - Removes the binary and Go build artifacts"
	@echo "  gotool   - Formats and vets the Go code"
	@echo "  help     - Shows this help message"