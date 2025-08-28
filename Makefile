# Makefile for the HCL Align project.

BINARY_NAME=hclalign
MODULE_NAME=github.com/oferchen/hclalign

all: build

build:
	@echo "Compiling the project..."
	go build ./...

run: build
	@echo "Running the project..."
	./${BINARY_NAME}

deps:
	@echo "Checking and downloading dependencies..."
	go mod download

tidy:
	@echo "Tidying and verifying module dependencies..."
	go mod tidy -v
	git diff --exit-code go.mod go.sum

lint:
	@echo "Running linters..."
	golangci-lint run

fmt:
	@echo "Formatting code..."
	go fmt ./...

vet:
	@echo "Running go vet..."
	go vet ./...

test:
	@echo "Running tests..."
	go test -shuffle=on -cover ./...

test-race:
	@echo "Running tests with race detector..."
	go test -race -shuffle=on -cover ./...

fuzz:
	@echo "Running fuzz tests..."
	go test ./... -run Fuzz

fuzz-short:
	@echo "Running short fuzz tests..."
	go test ./... -run Fuzz -fuzztime=5s

vulncheck:
	@echo "Checking for vulnerabilities..."
	govulncheck ./... || true

clean:
	@echo "Cleaning up..."
	go clean -modcache -fuzzcache
	rm -f ${BINARY_NAME}

init:
	@echo "Initializing Go module..."
	go mod init ${MODULE_NAME}

ci: fmt vet lint test-race fuzz-short build
	@echo "Running CI pipeline..."

help:
	@echo "Makefile commands:"
	@echo "all       - Compiles the project."
	@echo "build     - Builds the binary executable."
	@echo "run       - Runs the compiled binary."
	@echo "deps      - Downloads the project dependencies."
	@echo "tidy      - Tidies and verifies the module dependencies."
	@echo "fmt       - Formats the code."
	@echo "vet       - Runs go vet."
	@echo "lint      - Runs golangci-lint."
	@echo "test      - Runs all the tests."
	@echo "test-race - Runs tests with the race detector."
	@echo "fuzz      - Runs fuzz tests."
	@echo "fuzz-short - Runs short fuzz tests."
	@echo "vulncheck - Checks for vulnerabilities using govulncheck."
	@echo "ci        - Runs formatting, vetting, linting, tests, fuzz, and build."
	@echo "clean     - Cleans up the project."
	@echo "init      - Initializes a new Go module."
	@echo "help      - Prints this help message."

