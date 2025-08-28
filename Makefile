# Makefile for the HCL Align project.

BINARY_NAME=hclalign
MODULE_NAME=github.com/oferchen/hclalign

all: build

build:
	@echo "Compiling the project..."
	go build -o ${BINARY_NAME} main.go

run: build
	@echo "Running the project..."
	./${BINARY_NAME}

deps:
	@echo "Checking and downloading dependencies..."
	go mod download

tidy:
	@echo "Tidying and verifying module dependencies..."
	go mod tidy

test:
	@echo "Running tests..."
	go test ./...

clean:
        @echo "Cleaning up..."
        go clean -modcache -fuzzcache
        rm -f ${BINARY_NAME}

init:
	@echo "Initializing Go module..."
	go mod init ${MODULE_NAME}

help:
	@echo "Makefile commands:"
	@echo "all    - Compiles the project."
	@echo "build  - Builds the binary executable."
	@echo "run    - Runs the compiled binary."
	@echo "deps   - Downloads the project dependencies."
	@echo "tidy   - Tidies and verifies the module dependencies."
	@echo "test   - Runs all the tests."
	@echo "clean  - Cleans up the project."
	@echo "init   - Initializes a new Go module."
	@echo "help   - Prints this help message."

