#!/usr/bin/env bash
# Runs a Go wasm test binary under Node with a trimmed environment: Go's wasm
# runner rejects large environments (CI runners set many variables).
set -euo pipefail
runner="$(go env GOROOT)/lib/wasm/go_js_wasm_exec"
if [[ -n "${NODE_OPTIONS:-}" ]]; then
	exec env -i PATH="$PATH" NODE_OPTIONS="$NODE_OPTIONS" "$runner" "$@"
fi
exec env -i PATH="$PATH" "$runner" "$@"
