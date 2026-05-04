# go-cad Makefile
# Usage: make [target]
#
# Targets:
#   build             Build desktop binary + WASM for the current platform
#   wasm              Compile WASM frontend only (web/main.wasm)
#   serve             Run the development HTTP server on $PORT (default 8080)
#   test              Run all Go tests
#   vet               Run go vet on all packages
#   lint              Run golangci-lint (requires golangci-lint in PATH)
#   clean             Remove build artifacts
#   release           Cross-compile binaries for all release platforms
#   release-snapshot  Dry-run cross-compile (no zip, no push) for local testing

MODULE   := go-cad
WASM_OUT := web/main.wasm
BIN_OUT  := cad

# Allow callers to override PORT (default 8080)
PORT ?= 8080

# ── Primary targets ────────────────────────────────────────────────────────────

.PHONY: all
all: vet test build wasm

.PHONY: build
build: wasm
        go build -o $(BIN_OUT) ./cmd/cad/

.PHONY: wasm
wasm:
        GOOS=js GOARCH=wasm go build -buildvcs=false -o $(WASM_OUT) ./cmd/wasm/

.PHONY: serve
serve: wasm
        PORT=$(PORT) go run ./cmd/serve/ -host 0.0.0.0

.PHONY: test
test:
        go test ./...

.PHONY: vet
vet:
        go vet ./...

.PHONY: lint
lint:
        golangci-lint run ./...

.PHONY: clean
clean:
        rm -f $(BIN_OUT) $(WASM_OUT)
        find . -name '*.test' -delete

# ── Cross-platform release ────────────────────────────────────────────────────
# Outputs go to dist/ with file names: go-cad-<os>-<arch>[.exe]

DIST     := dist
PLATFORMS := \
        linux/amd64 \
        linux/arm64 \
        darwin/amd64 \
        darwin/arm64 \
        windows/amd64

.PHONY: release
release: clean
        @mkdir -p $(DIST)
        @for platform in $(PLATFORMS); do \
                os=$$(echo $$platform | cut -d/ -f1); \
                arch=$$(echo $$platform | cut -d/ -f2); \
                ext=""; \
                if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
                outfile="$(DIST)/go-cad-$$os-$$arch$$ext"; \
                echo "Building $$outfile …"; \
                GOOS=$$os GOARCH=$$arch go build -ldflags="-s -w" -o "$$outfile" ./cmd/cad/ || exit 1; \
        done
        GOOS=js GOARCH=wasm go build -buildvcs=false -o $(DIST)/main.wasm ./cmd/wasm/
        cd web && zip -r ../$(DIST)/go-cad-web.zip . -x "*.DS_Store"
        @echo "Release artifacts written to $(DIST)/"

# ── Help ──────────────────────────────────────────────────────────────────────

# ── Release snapshot (dry-run, no zip) ────────────────────────────────────────
# Useful for local validation before pushing a tag.

.PHONY: release-snapshot
release-snapshot:
        @mkdir -p $(DIST)
        @for platform in $(PLATFORMS); do \
                os=$$(echo $$platform | cut -d/ -f1); \
                arch=$$(echo $$platform | cut -d/ -f2); \
                ext=""; \
                if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
                outfile="$(DIST)/go-cad-$$os-$$arch$$ext"; \
                echo "Snapshot: $$outfile"; \
                GOOS=$$os GOARCH=$$arch go build -ldflags="-s -w" -o "$$outfile" ./cmd/cad/ || exit 1; \
        done
        GOOS=js GOARCH=wasm go build -buildvcs=false -o $(DIST)/main.wasm ./cmd/wasm/
        @echo "Snapshot complete (no zip/push). Artifacts in $(DIST)/"

# ── Help ──────────────────────────────────────────────────────────────────────

.PHONY: help
help:
        @echo "go-cad Makefile targets:"
        @echo "  build              Desktop binary + WASM (current platform)"
        @echo "  wasm               WASM frontend → web/main.wasm"
        @echo "  serve              Dev server on PORT=$(PORT)"
        @echo "  test               Run all Go tests"
        @echo "  vet                go vet all packages"
        @echo "  lint               golangci-lint all packages"
        @echo "  release            Cross-compile for linux/darwin/windows + web zip"
        @echo "  release-snapshot   Dry-run cross-compile (no zip)"
        @echo "  clean              Remove build artifacts"
