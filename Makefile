# /Makefile
PKG ?= ./...
BIN ?= ./.build/hclalign
COVER ?= ./.build/coverage.out
MINCOV ?= 95

GO ?= go

.PHONY: init tidy fmt strip lint vet test test-race fuzz cover build clean

init: ## download and verify modules
	$(GO) mod download
	$(GO) mod verify
	$(GO) version
	@if command -v terraform >/dev/null 2>&1; then \
	terraform version; \
	fi
	$(GO) run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1 version

tidy: ## tidy modules
	$(GO) mod tidy

fmt: ## format code and regenerate test fixtures
	$(GO) run mvdan.cc/gofumpt@v0.6.0 -w .
	@if command -v terraform >/dev/null 2>&1; then \
	terraform fmt -recursive tests/cases; \
	else \
	echo "terraform not found; skipping terraform fmt"; \
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
	$(GO) test -shuffle=on -cover $(PKG)

test-race: ## run tests with race detector
	$(GO) test -race -shuffle=on $(PKG)

fuzz: ## run fuzz tests
	$(GO) test ./... -run=^$ -fuzz=Fuzz -fuzztime=5s
	$(GO) test -run=^$$ -fuzz=Fuzz -fuzztime=10s ./internal/align
	$(GO) test -run=^$$ -fuzz=Fuzz -fuzztime=10s ./internal/engine
	$(GO) test -run=^$$ -fuzz=Fuzz -fuzztime=10s ./internal/hcl

cover: ## run coverage check
	mkdir -p $(dir $(COVER))
	$(GO) test -shuffle=on -covermode=atomic -coverpkg=./... -coverprofile=$(COVER) ./...
	$(GO) tool cover -func=$(COVER) | awk -v min=$(MINCOV) '/^total:/ { sub(/%/, "", $$3); if ($$3+0 < min) { printf "coverage %.1f%% is below %d%%\n", $$3, min; exit 1 } }'

build: ## build binary
	mkdir -p $(dir $(BIN))
	$(GO) build -trimpath -ldflags="-s -w" -buildvcs=false -o $(BIN) ./cmd/hclalign

clean: ## remove build artifacts
	rm -rf $(dir $(BIN))
