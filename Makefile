# Makefile for Tax App

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary output name
BINARY_NAME=taxapp
MAIN_PATH=cmd/taxapp/main.go
LEGACY_MAIN=main.go

# Environment
ENV?=dev

.PHONY: all build clean test cover run deps fmt vet tidy help

all: clean build

# Build the application using the new structure
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete"

# Clean build files
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Add this to your Makefile
cover:
	@echo "Generating coverage report..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	open coverage.html

# Run the application with specified environment
run:
	@echo "Running $(BINARY_NAME) with $(ENV) environment..."
	$(GORUN) $(MAIN_PATH) $(ENV)

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) vendor

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Vet code for potential issues
vet:
	@echo "Vetting code..."
	$(GOCMD) vet ./...

# Update go.mod to include all used packages
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

# Show help
help:
	@echo "Available commands:"
	@echo "  make build        - Build the application"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make test         - Run tests"
	@echo "  make cover        - Generate coverage report"
	@echo "  make run          - Run the application (ENV=dev|test|prod)"
	@echo "  make deps         - Download dependencies"
	@echo "  make fmt          - Format code"
	@echo "  make vet          - Static analysis"
	@echo "  make tidy         - Tidy go.mod file"


