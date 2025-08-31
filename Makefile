# /Makefile
PKG ?= ./...
BIN ?= .build/hclalign
COVER ?= coverage.out
MINCOV ?= 95

GO ?= go

.PHONY: init tidy fmt strip lint vet test test-race fuzz cover build clean

init: ## download and verify modules
	@$(GO) mod download
	@$(GO) mod verify
	@$(GO) version
	@if command -v terraform >/dev/null 2>&1; then \
	terraform version; \
	fi
	@$(GO) run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1 version

tidy: ## tidy modules
	@$(GO) mod tidy -v

fmt: ## format code and regenerate test fixtures
	@$(GO) run mvdan.cc/gofumpt@v0.6.0 -l -w .
	@if command -v terraform >/dev/null 2>&1; then \
	terraform fmt -recursive tests/cases; \
	else \
	echo "terraform not found; skipping terraform fmt"; \
	fi
	@find tests/cases -name in.tf -print0 | \
	xargs -0 -n1 -I{} sh -c 'dir=$$(dirname {}); $(GO) run ./cmd/hclalign --stdin --stdout < {} > $$dir/fmt.tf; $(GO) run ./cmd/hclalign --stdin --stdout --all < {} > $$dir/aligned.tf'

strip: ## normalize Go file comments and enforce policy
	@$(GO) run ./tools/stripcomments
	@$(GO) run ./cmd/commentcheck
	@git diff --exit-code

lint: ## run linters
	@$(GO) run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1 run --timeout=5m

vet: ## vet code
	@$(GO) vet $(PKG)

test: ## run tests
	@$(GO) test -shuffle=on -cover $(PKG)

test-race: ## run tests with race detector
	@$(GO) test -race -shuffle=on $(PKG)

fuzz: ## run fuzz tests
	@$(GO) test ./... -run=^$ -fuzz=Fuzz -fuzztime=5s
	@$(GO) test -run=^$$ -fuzz=Fuzz -fuzztime=10s ./internal/align
	@$(GO) test -run=^$$ -fuzz=Fuzz -fuzztime=10s ./internal/engine
	@$(GO) test -run=^$$ -fuzz=Fuzz -fuzztime=10s ./internal/hcl

cover: ## run coverage check
	@$(GO) test -shuffle=on -covermode=atomic -coverpkg=./... -coverprofile=$(COVER) ./...
	@$(GO) tool cover -func=$(COVER) | $(GO) run ./tools/covercheck $(MINCOV)

build: ## build binary
	@mkdir -p $(dir $(BIN))
	@$(GO) build -trimpath -ldflags="-s -w" -buildvcs=false -o $(BIN) ./cmd/hclalign

clean: ## remove build artifacts
	@rm -rf .build $(COVER)

