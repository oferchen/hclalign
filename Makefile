# /Makefile
APP := hclalign
PKG := ./...
BUILD_DIR := ./.build
COVERPROFILE := $(BUILD_DIR)/coverage.out

GO ?= go

FMT_PKGS := $(shell $(GO) list $(PKG))
FMT_DIRS := $(shell $(GO) list -f '{{.Dir}}' $(PKG))
GOFMT := $(shell $(GO) env GOROOT)/bin/gofmt

.PHONY: all init tidy fmt lint vet vuln sanitize test test-race fuzz-short cover cover-html build ci clean help

all: build ## build the project

init: ## download and verify modules
	$(GO) mod download
	$(GO) mod verify

tidy: ## tidy modules
	$(GO) mod tidy

fmt: ## format code
	$(GO) run mvdan.cc/gofumpt@latest -l -w $(FMT_DIRS)
	$(GOFMT) -s -w $(FMT_DIRS)
	@if command -v terraform >/dev/null 2>&1; then \
		terraform fmt -recursive tests/cases; \
	else \
		echo "terraform not found; skipping terraform fmt"; \
	fi
	$(GO) run ./cmd/hclalign --write tests/cases

lint: ## run linters
	$(GO) run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run --timeout=5m

sanitize: ## enforce comment policy
	mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/stripcomments ./tools/stripcomments
	$(BUILD_DIR)/stripcomments $(shell git ls-files '*.go')
	$(GO) run ./cmd/commentcheck
	$(GO) build $(PKG)

vet: ## vet code
	$(GO) vet $(PKG)

test: ## run tests
	mkdir -p $(BUILD_DIR)
	$(GO) test -race -shuffle=on -coverprofile=$(COVERPROFILE) $(PKG)

test-race: ## run tests with race detector
	mkdir -p $(BUILD_DIR)
	$(GO) test $(PKG) -race -shuffle=on

fuzz-short: ## short fuzzing run
	$(GO) test $(PKG) -run=^$ -fuzz=Fuzz -fuzztime=5s

cover: export COVER_THRESH ?= 95
cover: ## run coverage check
	mkdir -p $(BUILD_DIR)
	$(GO) test -race -shuffle=on -covermode=atomic -coverpkg=./... -coverprofile=$(COVERPROFILE) ./...
	$(GO) run ./internal/ci/covercheck $(COVERPROFILE)

cover-html: cover ## generate HTML coverage report
	$(GO) tool cover -html=$(COVERPROFILE) -o $(BUILD_DIR)/coverage.html

build: ## build binary
	mkdir -p $(BUILD_DIR)
	$(GO) build -trimpath -ldflags="-s -w" -buildvcs=false -o $(BUILD_DIR)/$(APP) ./cmd/hclalign

vuln: ## check vulnerabilities
	$(GO) run golang.org/x/vuln/cmd/govulncheck@latest $(PKG)

ci: tidy fmt sanitize lint vet vuln test test-race fuzz-short cover build ## run CI tasks

clean: ## remove build artifacts
	rm -rf $(BUILD_DIR) $(COVERPROFILE)

help: ## show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS=":.*?## "}; {printf "%-16s %s\n", $$1, $$2}'
