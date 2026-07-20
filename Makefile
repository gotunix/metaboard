.PHONY: build clean run test install uninstall fmt lint pre-commit help

BINARY_NAME=metaboard
MAIN_PATH=./cmd/metaboard
PREFIX?=/usr/local

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) $(MAIN_PATH)

clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)

run: build
	./$(BINARY_NAME)

test:
	@echo "Running tests..."
	go test ./... -v -cover

fmt:
	@echo "Formatting Go files..."
	go fmt ./...

lint:
	@echo "Running golangci-lint..."
	golangci-lint run

pre-commit:
	@echo "Running pre-commit checks..."
	pre-commit run --all-files

install: build
	@echo "Installing $(BINARY_NAME) to $(PREFIX)/bin..."
	mkdir -p $(PREFIX)/bin
	cp $(BINARY_NAME) $(PREFIX)/bin/$(BINARY_NAME)

uninstall:
	@echo "Removing $(BINARY_NAME) from $(PREFIX)/bin..."
	rm -f $(PREFIX)/bin/$(BINARY_NAME)

help:
	@echo "Usage: make [target] [PREFIX=/path/to/install]"
	@echo ""
	@echo "Targets:"
	@echo "  build       Build the Go binary"
	@echo "  clean       Remove the Go binary"
	@echo "  run         Build and run the dashboard"
	@echo "  test        Run all tests with coverage"
	@echo "  fmt         Format Go code"
	@echo "  lint        Lint code with golangci-lint"
	@echo "  pre-commit  Run all pre-commit hooks"
	@echo "  install     Install binary to $(PREFIX)/bin"
	@echo "  uninstall   Remove binary from $(PREFIX)/bin"
	@echo "  help        Show this help message"
