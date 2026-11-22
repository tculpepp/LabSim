DIST := dist
WASM_EXEC := $(shell go env GOROOT)/lib/wasm/wasm_exec.js

.PHONY: all test run wasm serve clean

all: test wasm

test:
	go vet ./...
	go test ./...

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
