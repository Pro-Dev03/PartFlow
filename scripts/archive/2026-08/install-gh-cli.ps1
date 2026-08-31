# GitHub CLI Installation Script for Windows (PowerShell)
# This script attempts to install GitHub CLI using available package managers

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "GitHub CLI Installation Script" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""

# Check if gh is already installed
$ghCommand = Get-Command gh -ErrorAction SilentlyContinue
if ($ghCommand) {
    Write-Host "✓ GitHub CLI is already installed!" -ForegroundColor Green
    gh --version
    Write-Host ""
    Write-Host "To login, run: gh auth login" -ForegroundColor Yellow
    exit 0
}

Write-Host "Searching for available package managers..." -ForegroundColor Yellow

# Try Winget (Windows Package Manager)
$wingetCommand = Get-Command winget -ErrorAction SilentlyContinue
if ($wingetCommand) {
    Write-Host "✓ Found Winget" -ForegroundColor Green
    Write-Host "Installing GitHub CLI using Winget..." -ForegroundColor Yellow
    winget install --id GitHub.cli -e --accept-source-agreements --accept-package-agreements
    Write-Host "✓ Installation completed!" -ForegroundColor Green
    Write-Host ""
    Write-Host "To login, run: gh auth login" -ForegroundColor Yellow
    exit 0
}

# Try Chocolatey
$chocoCommand = Get-Command choco -ErrorAction SilentlyContinue
if ($chocoCommand) {
    Write-Host "✓ Found Chocolatey" -ForegroundColor Green
    Write-Host "Installing GitHub CLI using Chocolatey..." -ForegroundColor Yellow
    choco install gh -y
    Write-Host "✓ Installation completed!" -ForegroundColor Green
    Write-Host ""
    Write-Host "You may need to restart your terminal or refresh your PATH" -ForegroundColor Yellow
    Write-Host "To login, run: gh auth login" -ForegroundColor Yellow
    exit 0
}

# Try Scoop
$scoopCommand = Get-Command scoop -ErrorAction SilentlyContinue
if ($scoopCommand) {
    Write-Host "✓ Found Scoop" -ForegroundColor Green
    Write-Host "Installing GitHub CLI using Scoop..." -ForegroundColor Yellow
    scoop install gh
    Write-Host "✓ Installation completed!" -ForegroundColor Green
    Write-Host ""
    Write-Host "You may need to restart your terminal or refresh your PATH" -ForegroundColor Yellow
    Write-Host "To login, run: gh auth login" -ForegroundColor Yellow
    exit 0
}

# If no package manager found, provide manual instructions
Write-Host "⚠ No package manager found (Winget, Chocolatey, or Scoop)" -ForegroundColor Red
Write-Host ""
Write-Host "Please install GitHub CLI manually:" -ForegroundColor Yellow
Write-Host ""
Write-Host "Option 1: Install Winget (Recommended)" -ForegroundColor Cyan
Write-Host "  - Download from: https://aka.ms/winget" -ForegroundColor Gray
Write-Host "  - Then run: winget install --id GitHub.cli" -ForegroundColor Gray
Write-Host ""
Write-Host "Option 2: Install Chocolatey" -ForegroundColor Cyan
Write-Host "  - Run in PowerShell (as Administrator):" -ForegroundColor Gray
Write-Host "    Set-ExecutionPolicy Bypass -Scope Process -Force;" -ForegroundColor Gray
Write-Host "    [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072;" -ForegroundColor Gray
Write-Host "    iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))" -ForegroundColor Gray
Write-Host "  - Then run: choco install gh" -ForegroundColor Gray
Write-Host ""
Write-Host "Option 3: Manual Download" -ForegroundColor Cyan
Write-Host "  - Download from: https://github.com/cli/cli/releases/latest" -ForegroundColor Gray
Write-Host "  - Extract and add to PATH" -ForegroundColor Gray
Write-Host ""
exit 1
