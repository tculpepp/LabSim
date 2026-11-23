DIST := dist
WASM_EXEC := $(shell go env GOROOT)/lib/wasm/wasm_exec.js
WASM_RUNNER := $(shell go env GOROOT)/lib/wasm/go_js_wasm_exec
# Pure logic packages (no Ebitengine) can run under Node.
PURE_PKGS = $(shell go list ./internal/... | grep -v -e /scenes -e /ui)

VERSION := $(shell cat VERSION)

.PHONY: all test test-wasm run wasm serve size clean

all: test test-wasm wasm

test:
	go vet ./...
	GOOS=js GOARCH=wasm go vet ./...
	go test ./...

# Run logic tests compiled to WASM under Node, with and without localStorage.
test-wasm:
	GOOS=js GOARCH=wasm go test -exec="$(WASM_RUNNER)" $(PURE_PKGS)
	LS=$$(mktemp -d)/ls.json; NODE_OPTIONS="--localstorage-file=$$LS" GOOS=js GOARCH=wasm go test -count=1 -exec="$(WASM_RUNNER)" ./internal/save

# Native desktop window (fast dev loop)
run:
	go run ./cmd/labsim

# Browser / PWA build into dist/
wasm:
	mkdir -p $(DIST)/icons
	GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o $(DIST)/labsim.wasm ./cmd/labsim
	cp $(WASM_EXEC) $(DIST)/
	cp web/index.html web/manifest.webmanifest $(DIST)/
	go run ./tools/genicon $(DIST)/icons
	# New cache name for every build, so installed copies pick up updates (T-4).
	BUILD="$(VERSION)-$$(cat $(DIST)/labsim.wasm | shasum -a 256 | cut -c1-8)"; \
	sed "s/__BUILD__/$$BUILD/" web/sw.js > $(DIST)/sw.js; \
	echo "service worker cache: labsim-$$BUILD"

# Download size check (T-10: under 15 MB compressed).
size: wasm
	@RAW=$$(wc -c < $(DIST)/labsim.wasm); GZ=$$(gzip -9 -c $(DIST)/labsim.wasm | wc -c); \
	echo "labsim.wasm: $$((RAW/1024)) KB raw, $$((GZ/1024)) KB gzipped"; \
	test $$GZ -lt 15728640 || (echo "over the 15 MB budget" && exit 1)

serve: wasm
	go run ./tools/serve -dir $(DIST)

clean:
	rm -rf $(DIST)
