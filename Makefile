# Go parameters
BINARY_NAME=revcli
MAIN_PACKAGE=.
GO=go

# Build info
VERSION?=0.2.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Linker flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

.PHONY: all build run clean test lint install uninstall reinstall help

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## build: Build the binary
build:
	$(GO) build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PACKAGE)

## run: Run without building binary
run:
	$(GO) run $(MAIN_PACKAGE)

## review: Run the review command (requires GEMINI_API_KEY)
review:
	$(GO) run $(MAIN_PACKAGE) review

## review-staged: Review only staged changes
review-staged:
	$(GO) run $(MAIN_PACKAGE) review --staged

## install: Install to $GOPATH/bin
install:
	$(GO) install $(LDFLAGS) $(MAIN_PACKAGE)

## uninstall: Remove from $GOPATH/bin
uninstall:
	rm -f $(shell go env GOPATH)/bin/$(BINARY_NAME)

## reinstall: Uninstall and reinstall to $GOPATH/bin
reinstall:
	make uninstall && make install

## test: Run tests
test:
	$(GO) test -v ./...

## test-cover: Run tests with coverage
test-cover:
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

## lint: Run linters
lint:
	$(GO) vet ./...
	@which golangci-lint > /dev/null || echo "golangci-lint not installed"
	@which golangci-lint > /dev/null && golangci-lint run ./...

## fmt: Format code
fmt:
	$(GO) fmt ./...

## tidy: Tidy dependencies
tidy:
	$(GO) mod tidy

## clean: Remove build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

## deps: Download dependencies
deps:
	$(GO) mod download

## all: Clean, tidy, lint, test, and build
all: clean tidy lint test build
