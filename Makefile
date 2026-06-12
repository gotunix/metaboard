.PHONY: build clean run test help

BINARY_NAME=metaboard
MAIN_PATH=./cmd/metaboard

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

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build   Build the Go binary"
	@echo "  clean   Remove the Go binary"
	@echo "  run     Build and run the dashboard"
	@echo "  test    Run all tests with coverage"
	@echo "  help    Show this help message"
