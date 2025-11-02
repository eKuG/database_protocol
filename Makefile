.PHONY: all build test bench clean docker docker-push k8s-deploy k8s-delete run help

# Variables
BINARY_NAME=protocol-server
DOCKER_IMAGE=protocol-server
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-X main.Version=${VERSION} -w -s"

# Default target
all: clean test build

# Help target
help:
	@echo "Available targets:"
	@echo "  make build       - Build the binary"
	@echo "  make test        - Run unit tests"
	@echo "  make bench       - Run benchmarks"
	@echo "  make run         - Run the application"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make docker      - Build Docker image"
	@echo "  make docker-push - Push Docker image to registry"
	@echo "  make k8s-deploy  - Deploy to Kubernetes"
	@echo "  make k8s-delete  - Remove from Kubernetes"
	@echo "  make fmt         - Format Go code"
	@echo "  make vet         - Run go vet"
	@echo "  make lint        - Run golangci-lint"

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME) .

# Run tests
test:
	@echo "Running tests..."
	$(GO) test $(GOFLAGS) -race -coverprofile=coverage.out ./...
	@echo "Coverage report:"
	@$(GO) tool cover -func=coverage.out

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem -benchtime=10s ./...

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out
	@rm -f *.prof
	@$(GO) clean -cache

# Format code
fmt:
	@echo "Formatting code..."
	@$(GO) fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	@$(GO) vet ./...

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin"; \
	fi

# Build Docker image
docker:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest

# Push Docker image to registry
docker-push: docker
	@echo "Pushing Docker image..."
	docker push $(DOCKER_IMAGE):$(VERSION)
	docker push $(DOCKER_IMAGE):latest

# Deploy to Kubernetes
k8s-deploy:
	@echo "Deploying to Kubernetes..."
	kubectl apply -f k8s-deployment.yaml
	@echo "Waiting for rollout..."
	kubectl -n protocol-integration rollout status statefulset/protocol-server

# Delete from Kubernetes
k8s-delete:
	@echo "Removing from Kubernetes..."
	kubectl delete -f k8s-deployment.yaml

# Get Kubernetes status
k8s-status:
	@echo "Kubernetes deployment status:"
	kubectl -n protocol-integration get all -l app=protocol-server

# Run load test (requires vegeta)
load-test: build
	@echo "Running load test..."
	@if command -v vegeta &> /dev/null; then \
		echo "GET http://localhost:9000/health" | vegeta attack -duration=30s -rate=100 | vegeta report; \
	else \
		echo "vegeta not installed. Install with: go install github.com/tsenart/vegeta@latest"; \
	fi

# Profile CPU
profile-cpu: build
	@echo "CPU profiling..."
	$(GO) test -cpuprofile=cpu.prof -bench=. ./...
	$(GO) tool pprof cpu.prof

# Profile Memory
profile-mem: build
	@echo "Memory profiling..."
	$(GO) test -memprofile=mem.prof -bench=. ./...
	$(GO) tool pprof mem.prof

# CI/CD pipeline simulation
ci: fmt vet test bench
	@echo "CI pipeline completed successfully!"

# Install development dependencies
dev-deps:
	@echo "Installing development dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/tsenart/vegeta@latest
	@echo "Dependencies installed!"
