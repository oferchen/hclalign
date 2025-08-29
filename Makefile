# /Makefile
APP := hclalign
PKG := ./...
BUILD_DIR := ./.build
COVERPROFILE := $(BUILD_DIR)/coverage.out

GO ?= go

.PHONY: all init tidy fmt lint vet test test-race cover build clean commentcheck fix-comments

all: build

init:
	$(GO) mod download
	$(GO) mod verify

tidy:
	$(GO) mod tidy

fmt:
	$(GO) run mvdan.cc/gofumpt@latest -w .
	$(GO) fmt $(PKG)

lint:
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run --timeout=5m

vet:
	$(GO) vet $(PKG)

test:
	mkdir -p $(BUILD_DIR)
	$(GO) test $(PKG) -shuffle=on -cover -covermode=atomic -coverprofile=$(COVERPROFILE)

test-race:
	mkdir -p $(BUILD_DIR)
	$(GO) test $(PKG) -race -shuffle=on

cover: test
	$(GO) tool cover -func=$(COVERPROFILE) | tee $(BUILD_DIR)/coverage.txt
	@cov=$$(grep total: $(BUILD_DIR)/coverage.txt | awk '{print $$3}' | sed 's/%//'); \
	awk 'BEGIN{exit !('"$$cov"' >= 95)}' </dev/null || { echo "Coverage $$cov% < 95%"; exit 1; }

build:
	mkdir -p $(BUILD_DIR)
	$(GO) build -trimpath -ldflags="-s -w" -o $(BUILD_DIR)/$(APP) .

commentcheck:
	$(GO) run ./cmd/commentcheck --mode=ci

fix-comments:
	$(GO) run ./cmd/commentcheck --mode=fix

clean:
	rm -rf $(BUILD_DIR)
