# /Makefile
APP := hclalign
PKG := ./...
BUILD_DIR := ./.build
COVERPROFILE := $(BUILD_DIR)/coverage.out

GO ?= go

.PHONY: init tidy fmt strip lint vet test test-race cover build clean

init: ## download and verify modules
	$(GO) mod download
	$(GO) mod verify

tidy: ## tidy modules
	$(GO) mod tidy

fmt: ## format code
	$(GO) run mvdan.cc/gofumpt@v0.6.0 -w .
	gofmt -s -w .
	$(MAKE) strip
	@if command -v terraform >/dev/null 2>&1; then \
	terraform fmt -recursive tests/cases; \
	fi
	$(GO) run ./cmd/hclalign tests/cases

strip: ## normalize Go file comments and enforce policy
	$(GO) run ./tools/stripcomments --repo-root "$(PWD)"
	$(GO) run ./cmd/commentcheck

lint: ## run linters
	       $(GO) run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1 run --timeout=5m

vet: ## vet code
	$(GO) vet $(PKG)

test: ## run tests
	mkdir -p $(BUILD_DIR)
	$(GO) test -race -shuffle=on -coverprofile=$(COVERPROFILE) $(PKG)

test-race: ## run tests with race detector
	$(GO) test $(PKG) -race -shuffle=on

cover: export COVER_THRESH ?= 95
cover: ## run coverage check
	mkdir -p $(BUILD_DIR)
	$(GO) test -race -shuffle=on -covermode=atomic -coverpkg=./... -coverprofile=$(COVERPROFILE) ./...
	$(GO) tool cover -func=$(COVERPROFILE) | awk -v thresh=$(COVER_THRESH) '/^total:/ { sub(/%/, "", $$3); if ($$3+0 < thresh) { printf "coverage %.1f%% is below %d%%\n", $$3, thresh; exit 1 } }'

build: ## build binary
	mkdir -p $(BUILD_DIR)
	$(GO) build -trimpath -ldflags="-s -w" -buildvcs=false -o $(BUILD_DIR)/$(APP) ./cmd/hclalign

clean: ## remove build artifacts
	rm -rf $(BUILD_DIR) $(COVERPROFILE)
