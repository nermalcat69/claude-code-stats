#!/bin/sh
set -e

REPO="nermalcat69/claude-code-stats"
BIN="claude-stats"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
  darwin) OS="macos" ;;
  linux)  OS="linux" ;;
  *)      echo "Unsupported OS: $OS"; exit 1 ;;
esac

case "$ARCH" in
  arm64|aarch64) ARCH="arm64" ;;
  x86_64|amd64)  ARCH="amd64" ;;
  *)             echo "Unsupported arch: $ARCH"; exit 1 ;;
esac

ASSET="${BIN}-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"

echo "Downloading $ASSET..."
curl -fsSL "$URL" -o "$BIN"
chmod +x "$BIN"

DEST="${HOME}/.local/bin"
mkdir -p "$DEST"
mv "$BIN" "$DEST/$BIN"

echo "Installed to $DEST/$BIN"
echo "Run: $BIN"
