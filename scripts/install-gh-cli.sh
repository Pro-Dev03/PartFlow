#!/bin/bash

# GitHub CLI Installation Script for Windows
# This script attempts to install GitHub CLI using available package managers

set -e

echo "========================================="
echo "GitHub CLI Installation Script"
echo "========================================="
echo ""

# Check if gh is already installed
if command -v gh &> /dev/null; then
    echo "✓ GitHub CLI is already installed!"
    gh --version
    echo ""
    echo "To login, run: gh auth login"
    exit 0
fi

echo "Searching for available package managers..."

# Try Winget (Windows Package Manager)
if command -v winget &> /dev/null; then
    echo "✓ Found Winget"
    echo "Installing GitHub CLI using Winget..."
    winget install --id GitHub.cli -e --accept-source-agreements --accept-package-agreements
    echo "✓ Installation completed!"
    echo ""
    echo "To login, run: gh auth login"
    exit 0
fi

# Try Chocolatey
if command -v choco &> /dev/null; then
    echo "✓ Found Chocolatey"
    echo "Installing GitHub CLI using Chocolatey..."
    choco install gh -y
    echo "✓ Installation completed!"
    echo ""
    echo "You may need to restart your terminal or refresh your PATH"
    echo "To login, run: gh auth login"
    exit 0
fi

# Try Scoop
if command -v scoop &> /dev/null; then
    echo "✓ Found Scoop"
    echo "Installing GitHub CLI using Scoop..."
    scoop install gh
    echo "✓ Installation completed!"
    echo ""
    echo "You may need to restart your terminal or refresh your PATH"
    echo "To login, run: gh auth login"
    exit 0
fi

# If no package manager found, provide manual instructions
echo "⚠ No package manager found (Winget, Chocolatey, or Scoop)"
echo ""
echo "Please install GitHub CLI manually:"
echo ""
echo "Option 1: Install Winget (Recommended)"
echo "  - Download from: https://aka.ms/winget"
echo "  - Then run: winget install --id GitHub.cli"
echo ""
echo "Option 2: Install Chocolatey"
echo "  - Run: Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))"
echo "  - Then run: choco install gh"
echo ""
echo "Option 3: Manual Download"
echo "  - Download from: https://github.com/cli/cli/releases/latest"
echo "  - Extract and add to PATH"
echo ""
exit 1
