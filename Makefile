# /Makefile
APP := hclalign
PKG := ./...
BUILD_DIR := ./.build
COVERPROFILE := $(BUILD_DIR)/coverage.out
COVER_THRESH ?= 95

GO ?= go

.PHONY: all init tidy fmt lint nocomments vet vuln test test-race cover cover-html build ci clean

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
	$(GO) run ./cmd/commentcheck

vet:
	$(GO) vet $(PKG)

test:
	mkdir -p $(BUILD_DIR)
	$(GO) test $(PKG) -shuffle=on -cover -covermode=atomic -coverprofile=$(COVERPROFILE)

test-race:
	mkdir -p $(BUILD_DIR)
	$(GO) test $(PKG) -race -shuffle=on

cover: test
	$(GO) tool cover -func=$(COVERPROFILE) | tee /dev/stderr | awk -v thr=$(COVER_THRESH) '/^total:/ {sub("%", "", $$3); if ($$3+0 < thr) {printf("coverage %s%% is below %s%%\n", $$3, thr); exit 1}}'

cover-html: test
	$(GO) tool cover -html=$(COVERPROFILE) -o $(BUILD_DIR)/coverage.html

build:
	mkdir -p $(BUILD_DIR)
	$(GO) build -trimpath -ldflags="-s -w" -o $(BUILD_DIR)/$(APP) ./cmd/hclalign

vuln:
	$(GO) run golang.org/x/vuln/cmd/govulncheck@latest $(PKG)

ci: tidy fmt lint vet vuln nocomments test test-race cover build

clean:
	rm -rf $(BUILD_DIR)
