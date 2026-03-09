#!/usr/bin/env bash
# machinist installer — downloads the latest release binary from GitHub.
# Usage: curl -fsSL https://raw.githubusercontent.com/moinsen-dev/machinist/main/install.sh | bash
set -euo pipefail

REPO="moinsen-dev/machinist"
INSTALL_DIR="${MACHINIST_INSTALL_DIR:-/usr/local/bin}"

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  arm64)   ARCH="arm64" ;;
  aarch64) ARCH="arm64" ;;
  *)
    echo "Error: Unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
if [ "$OS" != "darwin" ]; then
  echo "Error: machinist is macOS-only. Detected: $OS" >&2
  exit 1
fi

# Get latest release tag
echo "Fetching latest release..."
LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"v?([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
  echo "Error: Could not determine latest version" >&2
  exit 1
fi

echo "Installing machinist v${LATEST} (${OS}/${ARCH})..."

TARBALL="machinist_${LATEST}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/v${LATEST}/${TARBALL}"

# Download and extract
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

curl -fsSL "$URL" -o "${TMPDIR}/${TARBALL}"
tar -xzf "${TMPDIR}/${TARBALL}" -C "$TMPDIR"

# Install binary
if [ -w "$INSTALL_DIR" ]; then
  mv "${TMPDIR}/machinist" "${INSTALL_DIR}/machinist"
else
  echo "Installing to ${INSTALL_DIR} (requires sudo)..."
  sudo mv "${TMPDIR}/machinist" "${INSTALL_DIR}/machinist"
fi

chmod +x "${INSTALL_DIR}/machinist"

echo ""
echo "machinist v${LATEST} installed to ${INSTALL_DIR}/machinist"
echo ""
echo "Quick start:"
echo "  machinist snapshot                    # capture your current setup"
echo "  machinist dmg snapshot.toml           # bundle into a DMG"
echo "  machinist restore                     # restore on a new Mac"
