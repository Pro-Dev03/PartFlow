# GitHub CLI Installation Script - Simple Version
# Installs Chocolatey first, then GitHub CLI

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
    Read-Host "Press Enter to exit"
    exit 0
}

# Check if Chocolatey is installed
$chocoCommand = Get-Command choco -ErrorAction SilentlyContinue
if (-not $chocoCommand) {
    Write-Host "Chocolatey not found. Installing Chocolatey..." -ForegroundColor Yellow
    Write-Host "This requires Administrator privileges." -ForegroundColor Yellow
    Write-Host ""

    # Check if running as Administrator
    $currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
    $isAdmin = $currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)

    if (-not $isAdmin) {
        Write-Host "⚠ This script requires Administrator privileges to install Chocolatey." -ForegroundColor Red
        Write-Host "Please run PowerShell as Administrator and execute this script again." -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Alternatively, you can manually install GitHub CLI from:" -ForegroundColor Cyan
        Write-Host "https://github.com/cli/cli/releases/latest" -ForegroundColor Gray
        Read-Host "Press Enter to exit"
        exit 1
    }

    # Install Chocolatey
    Set-ExecutionPolicy Bypass -Scope Process -Force
    $protocol = [System.Net.ServicePointManager]::SecurityProtocol
    $protocol = $protocol -bor 3072
    [System.Net.ServicePointManager]::SecurityProtocol = $protocol
    iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

    # Refresh environment variables
    $machinePath = [System.Environment]::GetEnvironmentVariable("Path","Machine")
    $userPath = [System.Environment]::GetEnvironmentVariable("Path","User")
    $env:Path = "$machinePath;$userPath"

    Write-Host "✓ Chocolatey installed!" -ForegroundColor Green
    Write-Host ""
}

# Install GitHub CLI using Chocolatey
Write-Host "Installing GitHub CLI using Chocolatey..." -ForegroundColor Yellow
choco install gh -y

Write-Host "✓ Installation completed!" -ForegroundColor Green
Write-Host ""
Write-Host "Please restart your terminal for the changes to take effect." -ForegroundColor Yellow
Write-Host "Then run: gh auth login" -ForegroundColor Cyan
Write-Host ""
Read-Host "Press Enter to exit"
