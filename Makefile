# /Makefile
APP := hclalign
PKG := ./...
BUILD_DIR := ./.build
COVERPROFILE := $(BUILD_DIR)/coverage.out

GO ?= go

FMT_PKGS := $(shell $(GO) list $(PKG))
FMT_DIRS := $(shell $(GO) list -f '{{.Dir}}' $(PKG))

.PHONY: all init tidy fmt lint vet vuln test test-race cover cover-html build ci clean

all: build

init:
	$(GO) mod download
	$(GO) mod verify

tidy:
	$(GO) mod tidy

fmt:
	$(GO) fmt $(FMT_PKGS)
	$(GO) run mvdan.cc/gofumpt@latest -w $(FMT_DIRS)

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
	$(GO) run ./internal/ci/covercheck

cover-html: test
	$(GO) tool cover -html=$(COVERPROFILE) -o $(BUILD_DIR)/coverage.html

build:
	mkdir -p $(BUILD_DIR)
	$(GO) build -trimpath -buildvcs=false -ldflags="-s -w" -o $(BUILD_DIR)/$(APP) ./cmd/hclalign

vuln:
	$(GO) run golang.org/x/vuln/cmd/govulncheck@latest $(PKG)

ci: tidy fmt lint vuln test cover build

clean:
	rm -rf $(BUILD_DIR)
