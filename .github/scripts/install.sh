#!/bin/bash

set -e

# Get the latest release version if not provided
VERSION=${1:-$(curl -s https://api.github.com/repos/prompt-ops/pops/releases/latest | grep tag_name | cut -d '"' -f 4)}

# Check if the version is valid
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Invalid version format: '$VERSION'. Expected format: vX.Y.Z"
  exit 1
fi

echo "Installing Prompt-Ops $VERSION..."

# Detect OS and ARCH
OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [[ "$ARCH" == "x86_64" ]]; then
  ARCH="amd64"
elif [[ "$ARCH" == "aarch64" ]]; then
  ARCH="arm64"
else
  echo "Unsupported architecture: $ARCH"
  exit 1
fi

echo "Detected OS: $OS"
echo "Detected ARCH: $ARCH"

# Construct download URL
URL="https://github.com/prompt-ops/pops/releases/download/$VERSION/pops-${OS}-${ARCH}"

echo "Downloading Prompt-Ops $VERSION from $URL..."
curl -Lo pops "$URL"
chmod +x pops

# Move to /usr/local/bin
echo "Installing Prompt-Ops to /usr/local/bin..."
sudo mv pops /usr/local/bin/

echo "Prompt-Ops $VERSION installed successfully!"
