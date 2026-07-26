#!/bin/bash
# KDE Runtime Bootstrap Setup
# This script runs automatically when an OpenHands conversation starts
#
# Usage: Place this file at .openhands/setup.sh in your repository
# The script will be automatically executed at the start of each OpenHands conversation

set -e

# Ensure PATH includes Go
export PATH="$PATH:/usr/local/go/bin"

echo "=========================================="
echo "KDE Runtime Bootstrap Setup"
echo "=========================================="

# Change to project directory
cd /workspace/project/dnp3

# Install PyYAML (required for KDE Runtime)
if ! python3 -c "import yaml" 2>/dev/null; then
    echo "[1/4] Installing PyYAML..."
    pip install pyyaml --quiet
    echo "      PyYAML installed successfully"
else
    echo "[1/4] PyYAML already installed"
fi

# Install Go toolchain (required for Go projects)
if ! command -v go &> /dev/null; then
    echo "[2/4] Installing Go toolchain..."
    
    # Download Go if not present
    if [ ! -f /tmp/go1.22.5.linux-amd64.tar.gz ]; then
        curl -sL https://go.dev/dl/go1.22.5.linux-amd64.tar.gz -o /tmp/go1.22.5.linux-amd64.tar.gz
    fi
    
    # Install Go to /usr/local
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf /tmp/go1.22.5.linux-amd64.tar.gz
    
    echo "      Go 1.22.5 installed successfully"
else
    echo "[2/4] Go already installed: $(go version 2>/dev/null || echo 'unknown')"
fi

# Verify Go is in PATH
export PATH="$PATH:/usr/local/go/bin"

# Download Go module dependencies
if [ -f go.mod ]; then
    echo "[3/4] Downloading Go module dependencies..."
    go mod download 2>/dev/null || echo "      Warning: Could not download Go modules (may need network)"
    echo "      Go dependencies ready"
else
    echo "[3/4] No go.mod found, skipping dependency download"
fi

echo ""
echo "=========================================="
echo "Bootstrap Setup Complete"
echo "=========================================="

# Run KDE bootstrap gates
echo ""
echo "Running KDE Bootstrap Gates..."
python3 .kde/bootstrap/gates.py --project-type go || true

echo ""
echo "Runtime ready for investigation."
