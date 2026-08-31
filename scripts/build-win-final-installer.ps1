param(
    [string]$Version = "0.0.0",
    [string]$OutputDir = "$(Join-Path $PSScriptRoot '..\dist\windows-release')",
    [string]$FrontendDir = "$(Join-Path $PSScriptRoot '..\frontend')",
    [string]$BackendDir = "$(Join-Path $PSScriptRoot '..\backend')"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Write-Step([string]$Message) {
    Write-Host ""
    Write-Host "=== $Message ===" -ForegroundColor Cyan
}

$resolvedOutputDir = [System.IO.Path]::GetFullPath($OutputDir)
$resolvedFrontendDir = [System.IO.Path]::GetFullPath($FrontendDir)
$resolvedBackendDir = [System.IO.Path]::GetFullPath($BackendDir)

New-Item -ItemType Directory -Path $resolvedOutputDir -Force | Out-Null

Write-Step "Validating toolchain"
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw "Go is required but was not found in PATH."
}
if (-not (Get-Command npm -ErrorAction SilentlyContinue)) {
    throw "Node.js/npm is required but was not found in PATH."
}

Write-Step "Building backend for Windows"
Push-Location $resolvedBackendDir
try {
    $backendOut = Join-Path $resolvedOutputDir 'partflow-api.exe'
    go build -trimpath -ldflags "-s -w" -o $backendOut ./cmd/api
}
finally {
    Pop-Location
}

Write-Step "Installing frontend dependencies"
Push-Location $resolvedFrontendDir
try {
    if (Test-Path 'package-lock.json') {
        npm ci --no-fund --no-audit
    }
    else {
        npm install --no-fund --no-audit
    }
}
finally {
    Pop-Location
}

Write-Step "Building web app"
Push-Location $resolvedFrontendDir
try {
    npm run build
}
finally {
    Pop-Location
}

Write-Step "Packaging desktop installer"
Push-Location $resolvedFrontendDir
try {
    npx electron-builder --win --publish never
}
finally {
    Pop-Location
}

$bundleDir = Join-Path $resolvedOutputDir 'partflow-bundle'
New-Item -ItemType Directory -Path $bundleDir -Force | Out-Null

$backendBundlePath = Join-Path $bundleDir 'partflow-api.exe'
Copy-Item (Join-Path $resolvedOutputDir 'partflow-api.exe') $backendBundlePath -Force

$installerFiles = Get-ChildItem (Join-Path $resolvedFrontendDir 'dist-electron') -File -Recurse -Filter *.exe | Select-Object -ExpandProperty FullName
foreach ($installerFile in $installerFiles) {
    Copy-Item $installerFile $bundleDir -Force
}

$launchScript = @'
@echo off
setlocal
set "THISDIR=%~dp0"
start "" "%THISDIR%partflow-api.exe"
ping 127.0.0.1 -n 3 >nul
start "" "%THISDIR%PartFlow.exe"
'@
$launchScriptPath = Join-Path $bundleDir 'launch-partflow.bat'
Set-Content -Path $launchScriptPath -Value $launchScript -Encoding ASCII

$readmePath = Join-Path $bundleDir 'README.txt'
@"
PartFlow Windows Runtime Bundle
=============================

1. Start the app with launch-partflow.bat
2. The script launches the backend API and the desktop app together
3. The backend is located here: partflow-api.exe
4. The desktop app is located here: PartFlow.exe

This bundle is intended for local desktop usage and offline-first operations.

Online mode:
Set PARTFLOW_CLOUD_DATABASE_URL in the environment before launching PartFlow.
The value must be a PostgreSQL connection string with SSL enabled. The application
starts in Offline mode if this setting is unavailable.
"@ | Set-Content -Path $readmePath -Encoding UTF8

Write-Step "Final bundle created"
Write-Host "Backend: $backendBundlePath"
Write-Host "Bundle: $bundleDir"
Write-Host "Installer files copied to bundle:"
Get-ChildItem $bundleDir -File | ForEach-Object { Write-Host " - $($_.Name)" }
