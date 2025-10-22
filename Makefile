.PHONY: build build-all clean test deps list-functions

# List of all functions
FUNCTIONS := time-api hello-world receipt-processor

# Build all functions
build-all:
	@echo "Building all Lambda functions..."
	@mkdir -p dist
	@for func in $(FUNCTIONS); do \
		echo "Building $$func..."; \
		GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o dist/$$func/bootstrap ./functions/$$func; \
		echo "✓ $$func built successfully"; \
	done
	@echo "All functions built successfully!"

# Build a specific function
build:
	@if [ -z "$(FUNCTION)" ]; then \
		echo "Error: FUNCTION variable not set. Usage: make build FUNCTION=time-api"; \
		exit 1; \
	fi
	@echo "Building $(FUNCTION)..."
	@mkdir -p dist/$(FUNCTION)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o dist/$(FUNCTION)/bootstrap ./functions/$(FUNCTION)
	@echo "✓ Build complete: dist/$(FUNCTION)/bootstrap"

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf dist/
	@echo "Clean complete"

# Run tests for all functions
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests for a specific function
test-function:
	@if [ -z "$(FUNCTION)" ]; then \
		echo "Error: FUNCTION variable not set. Usage: make test-function FUNCTION=time-api"; \
		exit 1; \
	fi
	@echo "Testing $(FUNCTION)..."
	@go test -v ./functions/$(FUNCTION)

# Create deployment packages (zips)
zip-all: build-all
	@echo "Creating deployment packages..."
	@for func in $(FUNCTIONS); do \
		echo "Zipping $$func..."; \
		cd dist/$$func && zip -q ../$$func.zip bootstrap && cd ../..; \
		echo "✓ dist/$$func.zip created"; \
	done
	@echo "All deployment packages created!"

# Create zip for a specific function
zip: build
	@if [ -z "$(FUNCTION)" ]; then \
		echo "Error: FUNCTION variable not set. Usage: make zip FUNCTION=time-api"; \
		exit 1; \
	fi
	@echo "Creating deployment package for $(FUNCTION)..."
	@cd dist/$(FUNCTION) && zip -q ../$(FUNCTION).zip bootstrap
	@echo "✓ dist/$(FUNCTION).zip created"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed"

# List all functions
list-functions:
	@echo "Available functions:"
	@for func in $(FUNCTIONS); do \
		echo "  - $$func"; \
	done

# Run a quick health check
check:
	@echo "Running health check..."
	@go mod verify
	@go vet ./...
	@echo "✓ Health check passed"

