# Travio Monorepo Makefile

.PHONY: all build test clean proto

# List of all services
SERVICES := identity catalog inventory order payment notification fraud fulfillment search reporting queue audit subscription gateway

all: build

# Build all services
build: $(SERVICES)

$(SERVICES):
	@echo "Building $@..."
	@if [ -f services/$@/go.mod ]; then \
		cd services/$@ && go build -o ../../build/$@ ./cmd/...; \
	fi

# Test all services and pkg
test:
	@echo "Testing pkg..."
	@if [ -f pkg/go.mod ]; then \
		cd pkg && go test ./...; \
	fi
	@for service in $(SERVICES); do \
		echo "Testing $$service..."; \
		if [ -f services/$$service/go.mod ]; then \
			cd services/$$service && go test ./...; \
		fi \
	done

# Generate Protobuf code (requires buf)
proto:
	@echo "Generating protobufs..."
	# Assuming buf is installed and configured
	# buf generate

# Clean build artifacts
clean:
	@rm -rf build/*
