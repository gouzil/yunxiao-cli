#!/usr/bin/env bash
set -euo pipefail

version="${VERSION:-dev}"
commit="${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo none)}"
date="${DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"

platforms=(
  "darwin/arm64"
  "darwin/amd64"
  "linux/amd64"
  "linux/arm64"
  "windows/amd64"
)

mkdir -p dist
for platform in "${platforms[@]}"; do
  os="${platform%/*}"
  arch="${platform#*/}"
  output="dist/yunxiao-${os}-${arch}"
  if [[ "$os" == "windows" ]]; then
    output="${output}.exe"
  fi
  GOOS="$os" GOARCH="$arch" go build \
    -ldflags "-s -w -X main.version=${version} -X main.commit=${commit} -X main.date=${date}" \
    -o "$output" ./cmd/yunxiao
done
