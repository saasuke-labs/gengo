#!/usr/bin/env bash
set -euo pipefail

REPO="saasuke-labs/gengo"
INSTALL_DIR="$HOME/.local/bin"
TMP_DIR="$(mktemp -d)"
OS="$(uname | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
CHECKSUM_CMD="sha256sum"

# Map architecture names
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" && exit 1 ;;
esac

# macOS uses shasum
if [[ "$OS" == "darwin" ]]; then
  CHECKSUM_CMD="shasum -a 256"
fi

# Get latest release metadata
echo "Fetching latest Gengo release info..."
LATEST_JSON=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest")
TAG=$(echo "$LATEST_JSON" | jq -r .tag_name)
VERSION="${TAG#v}"
ASSET_NAME="gengo_${VERSION}_${OS}_${ARCH}.tar.gz"

echo "Latest version is $VERSION"
echo "Target asset: $ASSET_NAME"

# Download assets
echo "Downloading $ASSET_NAME..."
curl -sL -o "$TMP_DIR/$ASSET_NAME" \
  "https://github.com/${REPO}/releases/download/${TAG}/${ASSET_NAME}"

echo "Downloading checksums..."
curl -sL -o "$TMP_DIR/checksums.txt" \
  "https://github.com/${REPO}/releases/download/${TAG}/checksums.txt"

# Verify checksum
echo "Verifying checksum..."
cd "$TMP_DIR"
EXPECTED_SUM=$(grep "$ASSET_NAME" checksums.txt | awk '{print $1}')

if ! echo "$EXPECTED_SUM  $ASSET_NAME" | $CHECKSUM_CMD --check --status -; then
  echo "❌ Checksum verification failed!" && exit 1
else
  echo "✅ Checksum OK"
fi

# Extract and install
echo "Extracting..."
tar -xzf "$ASSET_NAME"

mkdir -p "$INSTALL_DIR"
install -m 755 gengo "$INSTALL_DIR/gengo"

# Suggest PATH update if needed
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
  echo "⚠️ $INSTALL_DIR is not in your PATH"
  echo "You can add it by adding this to your shell profile:"
  echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
else
  echo "✅ Gengo installed to $INSTALL_DIR and available in PATH"
fi

# Clean up
rm -rf "$TMP_DIR"
