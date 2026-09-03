# Herdr Web — Makefile
BIN      := bin/herdr-bridge
PKG      := ./cmd/herdr-bridge
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  := -s -w -X main.version=$(VERSION)
DIST     := internal/webui/dist

.PHONY: all dev build web go-build test test-go test-web lint fmt run clean release snapshot check screenshots

all: build

## dev: run the SvelteKit dev server (proxies to a running bridge on :7331)
dev:
	cd web && npm run dev

## web: build the SvelteKit UI and embed it into the Go binary tree
web:
	cd web && npm install --no-audit --no-fund && npm run build
	rm -rf $(DIST)
	mkdir -p $(DIST)
	cp -r web/build/. $(DIST)/

## build: build the single self-contained bridge binary (web embedded)
build: web go-build

go-build:
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN) $(PKG)

## run: build then run the bridge
run: build
	./$(BIN)

## test: run Go and web unit tests
test: test-go test-web

test-go:
	go test ./...

test-web:
	cd web && npm run test:unit

## lint: vet Go and type-check Svelte
lint:
	go vet ./...
	cd web && npm run check

## fmt: format Go
fmt:
	gofmt -w $(shell find . -name '*.go' -not -path './web/*')

## check: everything CI runs
check: lint test

## release: cut a release with GoReleaser (needs a tag)
release:
	goreleaser release --clean

## snapshot: local release build without publishing
snapshot:
	goreleaser release --snapshot --clean

## screenshots: capture UI screenshots (requires `make run` in another shell)
screenshots:
	cd web && npm install --no-audit --no-fund >/dev/null && npx playwright install chromium >/dev/null
	node web/scripts/screenshots.mjs http://127.0.0.1:7331 docs/screenshots

## clean: remove build artifacts
clean:
	rm -rf bin dist web/build $(DIST)
	mkdir -p $(DIST) && echo "<!doctype html><title>herdrweb</title>" > $(DIST)/index.html
