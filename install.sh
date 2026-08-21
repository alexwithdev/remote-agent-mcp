#!/bin/sh
# install.sh — install remote-agent-mcp for the current Linux platform.
#
# Downloads the release binary from GitHub, verifies its SHA256 checksum,
# and installs it into a bin directory on PATH.
#
# Usage:
#   ./install.sh                 # install latest release
#   ./install.sh v1.2.3          # install a specific version
#   ./install.sh --version v1.2.3
#   ./install.sh --prefix /opt/bin
#   ./install.sh --help
#
# Exit codes:
#   0  success
#   1  error (bad args, download/verify/install failure)
#   2  unsupported platform

set -eu

REPO="alexwithdev/remote-agent-mcp"
BINARY="remote-agent-mcp"
API="https://api.github.com/repos/${REPO}/releases"
DL="https://github.com/${REPO}/releases/download"

# --- argument parsing -------------------------------------------------------

VERSION=""
PREFIX=""

usage() {
    sed -n '2,16p' "$0" | sed 's/^# \{0,1\}//'
    exit 0
}

while [ "$#" -gt 0 ]; do
    case "$1" in
        --help|-h)
            usage
            ;;
        --version)
            [ "$#" -ge 2 ] || { echo "error: --version requires a value" >&2; exit 1; }
            VERSION="$2"
            shift 2
            ;;
        --prefix)
            [ "$#" -ge 2 ] || { echo "error: --prefix requires a value" >&2; exit 1; }
            PREFIX="$2"
            shift 2
            ;;
        -*)
            echo "error: unknown option: $1" >&2
            usage
            ;;
        *)
            # positional version argument
            [ -z "$VERSION" ] || { echo "error: version specified twice" >&2; exit 1; }
            VERSION="$1"
            shift
            ;;
    esac
done

# --- platform detection -----------------------------------------------------

detect_os_arch() {
    os="$(uname -s)"
    arch="$(uname -m)"

    case "$os" in
        Linux) os="linux" ;;
        *)
            echo "error: unsupported OS: $os (install.sh supports Linux only)" >&2
            exit 2
            ;;
    esac

    case "$arch" in
        x86_64|amd64)  arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *)
            echo "error: unsupported architecture: $arch" >&2
            exit 2
            ;;
    esac

    echo "${os}-${arch}"
}

# --- resolve version --------------------------------------------------------

resolve_version() {
    if [ -n "$VERSION" ]; then
        echo "$VERSION"
        return
    fi
    # latest release tag from the GitHub API, parsed without jq
    curl -fsSL "${API}/latest" \
        | grep -o '"tag_name"[[:space:]]*:[[:space:]]*"[^"]*"' \
        | head -n1 \
        | sed 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/'
}

# --- main -------------------------------------------------------------------

need_cmd() {
    command -v "$1" >/dev/null 2>&1 || {
        echo "error: required command not found: $1" >&2
        exit 1
    }
}

need_cmd curl
need_cmd uname
need_cmd sha256sum

PLATFORM="$(detect_os_arch)"
VERSION="$(resolve_version)"
[ -n "$VERSION" ] || { echo "error: could not determine latest version" >&2; exit 1; }

# strip a leading 'v' for the asset name if present
ASSET_VERSION="$VERSION"
case "$ASSET_VERSION" in
    v*) ASSET_VERSION="${ASSET_VERSION#v}" ;;
esac

ASSET="${BINARY}-${PLATFORM}"
URL="${DL}/${VERSION}/${ASSET}"
CHECKSUM_URL="${DL}/${VERSION}/checksums.txt"

# choose install directory
if [ -n "$PREFIX" ]; then
    INSTALL_DIR="$PREFIX"
elif [ -w /usr/local/bin ]; then
    INSTALL_DIR="/usr/local/bin"
elif [ -d "$HOME/.local/bin" ] || mkdir -p "$HOME/.local/bin" 2>/dev/null; then
    INSTALL_DIR="$HOME/.local/bin"
else
    echo "error: no writable install directory (try --prefix)" >&2
    exit 1
fi

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

echo "==> downloading ${ASSET} (${VERSION})"
# -fL without -s: curl shows a progress meter when stderr is a terminal and
# stays quiet in non-interactive environments (CI, pipes).
curl -fL -o "$TMPDIR/$ASSET" "$URL"

echo "==> downloading checksums.txt"
curl -fL -o "$TMPDIR/checksums.txt" "$CHECKSUM_URL"

echo "==> verifying SHA256 checksum"
(
    cd "$TMPDIR"
    grep -F "$ASSET" checksums.txt | sha256sum -c - >/dev/null 2>&1
) || {
    echo "error: checksum verification failed for $ASSET" >&2
    exit 1
}

echo "==> installing to ${INSTALL_DIR}"
mkdir -p "$INSTALL_DIR"
install -m 0755 "$TMPDIR/$ASSET" "$INSTALL_DIR/$BINARY"

echo "==> installed ${BINARY} ${VERSION} to ${INSTALL_DIR}/${BINARY}"
echo "    ensure ${INSTALL_DIR} is on your PATH"
