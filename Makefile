.PHONY: build install test lint clean deps check

# Build the CLI
build:
	go build -o bin/zod2go ./cmd/zod2go

# Install the CLI
install:
	go install ./cmd/zod2go

# Run tests
test:
	go test -v ./...

# Run linter
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out

# Install Go dependencies
deps:
	go mod tidy

# Install Node.js dependencies
deps-node:
	cd scripts && npm install

# Check all dependencies
check: build
	./bin/zod2go check

# Run all checks
all: deps lint test build

# Example: Generate Insomnia types (requires cloned repo)
example-insomnia:
	@if [ ! -d "/tmp/insomnia" ]; then \
		echo "Cloning Insomnia repository..."; \
		git clone --depth 1 https://github.com/Kong/insomnia.git /tmp/insomnia; \
	fi
	./bin/zod2go generate \
		-i /tmp/insomnia/packages/insomnia/src/common/import-v5-parser.ts \
		-o examples/insomnia/insomnia_gen.go \
		-p insomnia \
		--export InsomniaFileSchema
