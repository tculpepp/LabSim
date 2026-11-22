DIST := dist
WASM_EXEC := $(shell go env GOROOT)/lib/wasm/wasm_exec.js
WASM_RUNNER := $(shell go env GOROOT)/lib/wasm/go_js_wasm_exec
# Pure logic packages (no Ebitengine) can run under Node.
PURE_PKGS = $(shell go list ./internal/... | grep -v -e /scenes -e /ui)

.PHONY: all test test-wasm run wasm serve clean

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
	cp web/index.html web/manifest.webmanifest web/sw.js $(DIST)/
	go run ./tools/genicon $(DIST)/icons

serve: wasm
	go run ./tools/serve -dir $(DIST)

clean:
	rm -rf $(DIST)
