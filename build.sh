#!/bin/sh
# build.sh — build remote-agent-mcp for the local platform or cross-compile
# for other OS/architectures. Outputs go to ./output/.
#
# Usage:
#   ./build.sh [target]
#
# Targets:
#   local          Build for the current platform (default)
#   linux-amd64    Linux x86_64
#   linux-arm64    Linux ARM64 (e.g. Raspberry Pi, ARM servers)
#   darwin-amd64   macOS x86_64
#   darwin-arm64   macOS Apple Silicon
#   windows-amd64  Windows x86_64
#   all            Build every target above
#   clean          Remove the ./output/ directory
#
# Examples:
#   ./build.sh                 # local build
#   ./build.sh linux-arm64     # cross-compile for Linux ARM64
#   ./build.sh all             # build all platforms
#   ./build.sh clean           # remove output/

set -e

BINARY="remote-agent-mcp"
OUTPUT_DIR="output"

usage() {
    sed -n '2,20p' "$0" | sed 's/^# \{0,1\}//'
    exit 1
}

# build <os> <arch> [suffix]
build() {
    os="$1"
    arch="$2"
    suffix="$3"
    name="${BINARY}-${os}-${arch}${suffix}"
    mkdir -p "$OUTPUT_DIR"
    echo "==> building ${name}"
    CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -o "$OUTPUT_DIR/$name" .
}

build_local() {
    mkdir -p "$OUTPUT_DIR"
    echo "==> building ${BINARY} (local)"
    go build -o "$OUTPUT_DIR/$BINARY" .
}

clean() {
    echo "==> removing ${OUTPUT_DIR}/"
    rm -rf "$OUTPUT_DIR"
}

target="${1:-local}"

case "$target" in
    local)
        build_local
        ;;
    linux-amd64)
        build linux amd64
        ;;
    linux-arm64)
        build linux arm64
        ;;
    darwin-amd64)
        build darwin amd64
        ;;
    darwin-arm64)
        build darwin arm64
        ;;
    windows-amd64)
        build windows amd64 .exe
        ;;
    all)
        build_local
        build linux amd64
        build linux arm64
        build darwin amd64
        build darwin arm64
        build windows amd64 .exe
        ;;
    clean)
        clean
        ;;
    *)
        echo "unknown target: $target" >&2
        echo
        usage
        ;;
esac

echo "done."
