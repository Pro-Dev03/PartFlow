param(
    [string]$Version = "0.0.0",
    [string]$OutputDir = "$(Join-Path $PSScriptRoot '..\dist\windows-release')",
    [string]$FrontendDir = "$(Join-Path $PSScriptRoot '..\frontend')",
    [string]$BackendDir = "$(Join-Path $PSScriptRoot '..\backend')",
    [string]$ElectronOutputDir = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Write-Step([string]$Message) {
    Write-Host ""
    Write-Host "=== $Message ===" -ForegroundColor Cyan
}

function Invoke-Native([string]$Command, [string[]]$Arguments) {
    & $Command @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "Command failed with exit code ${LASTEXITCODE}: $Command $($Arguments -join ' ')"
    }
}

$resolvedOutputDir = [System.IO.Path]::GetFullPath($OutputDir)
$resolvedFrontendDir = [System.IO.Path]::GetFullPath($FrontendDir)
$resolvedBackendDir = [System.IO.Path]::GetFullPath($BackendDir)
$releaseStamp = Get-Date -Format 'yyyyMMdd-HHmmss'
if ([string]::IsNullOrWhiteSpace($ElectronOutputDir)) {
    $ElectronOutputDir = Join-Path $env:TEMP "PartFlow-release-$Version-$releaseStamp"
}
$resolvedElectronOutputDir = [System.IO.Path]::GetFullPath($ElectronOutputDir)

if (Test-Path $resolvedOutputDir) {
    Remove-Item -Recurse -Force $resolvedOutputDir
}
New-Item -ItemType Directory -Path $resolvedOutputDir -Force | Out-Null
if (Test-Path $resolvedElectronOutputDir) {
    Remove-Item -Recurse -Force $resolvedElectronOutputDir
}

$backendInput = Join-Path $resolvedOutputDir 'partflow-api.exe'
$certificateInput = Join-Path $resolvedFrontendDir 'build\PartFlow-Internal-Code-Signing.cer'
if (-not (Test-Path $certificateInput)) {
    throw "Release certificate was not found: $certificateInput"
}

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
    Invoke-Native 'go' @('build', '-trimpath', '-ldflags', '-s -w', '-o', $backendInput, './cmd/api')
}
finally {
    Pop-Location
}

Write-Step "Installing frontend dependencies"
Push-Location $resolvedFrontendDir
try {
    if (Test-Path 'package-lock.json') {
        Invoke-Native 'npm' @('ci', '--no-fund', '--no-audit')
    }
    else {
        Invoke-Native 'npm' @('install', '--no-fund', '--no-audit')
    }
}
finally {
    Pop-Location
}

Write-Step "Building web app"
Push-Location $resolvedFrontendDir
try {
    Invoke-Native 'npm' @('run', 'build')
}
finally {
    Pop-Location
}

Write-Step "Packaging desktop installer"
$electronConfigPath = Join-Path $resolvedFrontendDir '.partflow-electron-builder.release.json'
$electronConfig = Get-Content (Join-Path $resolvedFrontendDir 'electron-builder.json') -Raw | ConvertFrom-Json
$electronConfig.directories.output = $resolvedElectronOutputDir
if ($null -eq $electronConfig.extraMetadata) {
    $electronConfig | Add-Member -MemberType NoteProperty -Name extraMetadata -Value ([pscustomobject]@{})
}
$electronConfig.extraMetadata | Add-Member -MemberType NoteProperty -Name version -Value $Version -Force
$electronConfig.extraResources[0].from = $backendInput
$electronConfig | ConvertTo-Json -Depth 10 | Set-Content -Path $electronConfigPath -Encoding UTF8
Push-Location $resolvedFrontendDir
try {
    Invoke-Native 'npm' @(
        'exec', '--no', '--', 'electron-builder', '--win', '--publish', 'never',
        '--config', $electronConfigPath
    )
}
finally {
    Pop-Location
    Remove-Item $electronConfigPath -Force -ErrorAction SilentlyContinue
}

$bundleDir = Join-Path $resolvedOutputDir 'partflow-bundle'
if (Test-Path $bundleDir) {
    Remove-Item -Recurse -Force $bundleDir
}
New-Item -ItemType Directory -Path $bundleDir -Force | Out-Null

$backendBundlePath = Join-Path $bundleDir 'partflow-api.exe'
Copy-Item (Join-Path $resolvedOutputDir 'partflow-api.exe') $backendBundlePath -Force

$expectedInstaller = Join-Path $resolvedElectronOutputDir "PartFlow-$Version-setup.exe"
$expectedPortable = Join-Path $resolvedElectronOutputDir "PartFlow-$Version-portable.exe"
$unpackedDir = Join-Path $resolvedElectronOutputDir 'win-unpacked'
foreach ($requiredPath in @($expectedInstaller, $expectedPortable, (Join-Path $unpackedDir 'PartFlow.exe'), (Join-Path $unpackedDir 'resources\backend\partflow-api.exe'))) {
    if (-not (Test-Path $requiredPath)) {
        throw "Required release artifact was not found: $requiredPath"
    }
}

Copy-Item $expectedInstaller $bundleDir -Force
Copy-Item $expectedPortable $bundleDir -Force
Copy-Item $unpackedDir (Join-Path $bundleDir 'win-unpacked') -Recurse -Force

$launchScript = @'
@echo off
setlocal
set "THISDIR=%~dp0"
start "" "%THISDIR%win-unpacked\PartFlow.exe"
'@
$launchScriptPath = Join-Path $bundleDir 'launch-partflow.bat'
Set-Content -Path $launchScriptPath -Value $launchScript -Encoding ASCII

$readmePath = Join-Path $bundleDir 'README.txt'
@"
PartFlow Windows Runtime Bundle
=============================

1. Install and run PartFlow-$Version-setup.exe, or run win-unpacked\PartFlow.exe for unpacked verification.
2. The desktop application starts and owns its embedded backend process.
3. The standalone partflow-api.exe is included for artifact inspection only.

This bundle is intended for local desktop usage and offline-first operations.

Online mode:
Set PARTFLOW_CLOUD_DATABASE_URL in the environment before launching PartFlow.
The value must be a PostgreSQL connection string with SSL enabled. The application
starts in Offline mode if this setting is unavailable.
"@ | Set-Content -Path $readmePath -Encoding UTF8

$manifest = [ordered]@{
    product = 'PartFlow'
    version = $Version
    builtAtUtc = (Get-Date).ToUniversalTime().ToString('o')
    gitCommit = (& git -C (Split-Path $PSScriptRoot -Parent) rev-parse HEAD 2>$null).Trim()
    artifacts = @(Get-ChildItem $bundleDir -File -Recurse | ForEach-Object {
        $relativePath = $_.FullName.Substring($bundleDir.Length).TrimStart('\', '/')
        [ordered]@{
            path = $relativePath
            bytes = $_.Length
            sha256 = (Get-FileHash $_.FullName -Algorithm SHA256).Hash
        }
    })
}
$manifestPath = Join-Path $bundleDir 'release-manifest.json'
$manifest | ConvertTo-Json -Depth 6 | Set-Content -Path $manifestPath -Encoding UTF8

Write-Step "Final bundle created"
Write-Host "Backend: $backendBundlePath"
Write-Host "Bundle: $bundleDir"
Write-Host "Electron output: $resolvedElectronOutputDir"
Write-Host "Installer files copied to bundle:"
Get-ChildItem $bundleDir -File | ForEach-Object { Write-Host " - $($_.Name)" }
