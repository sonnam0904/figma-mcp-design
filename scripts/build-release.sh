#!/usr/bin/env bash
# Build cross-platform archives for GitHub Release. Usage: build-release.sh <version e.g. v1.2.3>
set -euo pipefail

VERSION="${1:?usage: build-release.sh <version>}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

rm -rf dist
mkdir -p dist

platforms=(
  "darwin amd64"
  "darwin arm64"
  "linux amd64"
  "linux arm64"
  "windows amd64"
  "windows arm64"
)

for pair in "${platforms[@]}"; do
  read -r GOOS GOARCH <<<"$pair"
  rm -rf build
  mkdir -p build
  ext=""
  if [[ "$GOOS" == "windows" ]]; then
    ext=".exe"
  fi
  CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -trimpath -ldflags="-s -w" -o "build/mcp-server${ext}" ./cmd/mcp-server
  CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -trimpath -ldflags="-s -w" -o "build/relay${ext}" ./cmd/relay

  name="figma-mcp-design-${VERSION}-${GOOS}-${GOARCH}"
  if [[ "$GOOS" == "windows" ]]; then
    (cd build && zip -q "../dist/${name}.zip" "mcp-server${ext}" "relay${ext}")
  else
    tar -czf "dist/${name}.tar.gz" -C build "mcp-server${ext}" "relay${ext}"
  fi
done

(
  cd dist
  shopt -s nullglob
  : >SHASUMS256.txt
  for f in *.tar.gz *.zip; do
    sha256sum "$f"
  done | sort -k2 >SHASUMS256.txt
  test -s SHASUMS256.txt
)
