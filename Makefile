# /Makefile
APP := hclalign
PKG := ./...
BUILD_DIR := ./.build
COVERPROFILE := coverage.out
COVER_THRESH ?= 95

GO ?= go

FMT_PKGS := $(shell $(GO) list $(PKG))
FMT_DIRS := $(shell $(GO) list -f '{{.Dir}}' $(PKG))

.PHONY: all init tidy fmt lint vet vuln nocomments test test-race cover cover-html build ci clean

all: build

init:
	$(GO) mod download
	$(GO) mod verify

tidy:
	$(GO) mod tidy

fmt:
	gofumpt -l -w . && gofmt -s -w .

lint:
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run --timeout=5m

nocomments:
	mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/nocomments ./tools/nocomments
	$(BUILD_DIR)/nocomments $(shell git ls-files '*.go')
	$(GO) run ./cmd/commentcheck
	$(GO) build $(PKG)

vet:
	$(GO) vet $(PKG)

test:
	$(GO) test -shuffle=on $(PKG)

test-race:
	mkdir -p $(BUILD_DIR)
	$(GO) test $(PKG) -race -shuffle=on

cover:
	$(GO) test -race -covermode=atomic -coverpkg=./... -coverprofile=$(COVERPROFILE) ./...
	$(GO) tool cover -func=$(COVERPROFILE) | tee /dev/stderr | awk -v thr=$(COVER_THRESH) '/^total:/ {sub("%", "", $$3); if ($$3+0 < thr) {printf("coverage %s%% is below %s%%\n", $$3, thr); exit 1}}'

cover-html: cover
	$(GO) tool cover -html=$(COVERPROFILE) -o $(BUILD_DIR)/coverage.html

build:
	mkdir -p $(BUILD_DIR)
	$(GO) build -trimpath -ldflags="-s -w" -buildvcs=false -o $(BUILD_DIR)/$(APP) ./cmd/hclalign

vuln:
	$(GO) run golang.org/x/vuln/cmd/govulncheck@latest $(PKG)

ci: tidy fmt nocomments lint vet vuln test-race cover build

clean:
	rm -rf $(BUILD_DIR) $(COVERPROFILE)
