#!/bin/bash

set -euo pipefail

# Get the latest release version if not provided
VERSION=${1:-$(curl -s https://api.github.com/repos/prompt-ops/pops/releases/latest | grep '"tag_name"' | cut -d '"' -f 4)}

# Check if the version is valid
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Error: Invalid version format: '$VERSION'. Expected format: vX.Y.Z"
  exit 1
fi

echo "Installing Prompt-Ops $VERSION..."

# Detect OS and ARCH
OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
x86_64) ARCH="amd64" ;;
arm64 | aarch64) ARCH="arm64" ;;
*)
  echo "Error: Unsupported architecture: $ARCH"
  exit 1
  ;;
esac

echo "Detected OS: $OS"
echo "Detected ARCH: $ARCH"

# Construct download URL
URL="https://github.com/prompt-ops/pops/releases/download/$VERSION/pops-${OS}-${ARCH}"

# Download binary
echo "Downloading Prompt-Ops $VERSION from $URL..."
curl -Lo pops "$URL"
chmod +x pops

# Move to /usr/local/bin (requires sudo)
echo "Installing Prompt-Ops to /usr/local/bin..."
sudo mv pops /usr/local/bin/

# Verify installation
if command -v pops >/dev/null 2>&1; then
  echo "Prompt-Ops $VERSION installed successfully!"
  pops version
else
  echo "Error: Installation failed. 'pops' command not found."
  exit 1
fi
