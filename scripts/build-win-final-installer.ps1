[CmdletBinding()]
param(
  [string]$Version,
  [switch]$SkipTests
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$repoRoot = Split-Path -Parent $PSScriptRoot
$frontendSource = Join-Path $repoRoot 'frontend'
$backendSource = Join-Path $repoRoot 'backend'
$releaseRoot = Join-Path $repoRoot 'dist\windows-release'
$certificatePath = Join-Path $frontendSource 'build\PartFlow-Internal-Code-Signing.cer'
$installerIncludePath = Join-Path $frontendSource 'build\installer.nsh'
$packagePath = Join-Path $frontendSource 'package.json'
$lockPath = Join-Path $frontendSource 'package-lock.json'

foreach ($toolName in @('go.exe', 'node.exe', 'npm.cmd', 'robocopy.exe')) {
  if (-not (Get-Command $toolName -ErrorAction SilentlyContinue)) {
    throw "Required build tool is missing: $toolName"
  }
}

foreach ($path in @($packagePath, $lockPath, $certificatePath, $installerIncludePath)) {
  if (-not (Test-Path -LiteralPath $path -PathType Leaf)) {
    throw "Required Windows build input is missing: $path"
  }
}

$package = Get-Content -LiteralPath $packagePath -Raw -Encoding UTF8 | ConvertFrom-Json
if ([string]::IsNullOrWhiteSpace($Version)) {
  $Version = [string]$package.version
}
if ($Version -notmatch '^\d+\.\d+\.\d+(?:[-+][0-9A-Za-z.-]+)?$') {
  throw "Version must be a valid semantic version; received '$Version'."
}

$publicCertificate = Get-PfxCertificate -FilePath $certificatePath
if ($publicCertificate.Subject -ne 'CN=PartFlow Internal Code Signing') {
  throw "The public certificate subject is not PartFlow Internal Code Signing."
}
if ((Get-Date) -lt $publicCertificate.NotBefore -or (Get-Date) -gt $publicCertificate.NotAfter) {
  throw 'The configured public code-signing certificate is outside its validity period.'
}

$signingCertificate = Get-ChildItem Cert:\CurrentUser\My -ErrorAction SilentlyContinue |
  Where-Object { $_.Thumbprint -eq $publicCertificate.Thumbprint -and $_.HasPrivateKey } |
  Select-Object -First 1
if (-not $signingCertificate) {
  throw 'The matching code-signing private key is not available in Cert:\CurrentUser\My. Follow docs/MANUAL-INTERNAL-SIGNING-SETUP.md on this build machine.'
}

$trustedRoot = Get-ChildItem Cert:\CurrentUser\Root -ErrorAction SilentlyContinue |
  Where-Object { $_.Thumbprint -eq $publicCertificate.Thumbprint } |
  Select-Object -First 1
$trustedPublisher = Get-ChildItem Cert:\CurrentUser\TrustedPublisher -ErrorAction SilentlyContinue |
  Where-Object { $_.Thumbprint -eq $publicCertificate.Thumbprint } |
  Select-Object -First 1
if (-not $trustedRoot -or -not $trustedPublisher) {
  throw 'The public certificate must be trusted in the current user Root and TrustedPublisher stores. Follow docs/MANUAL-INTERNAL-SIGNING-SETUP.md.'
}

New-Item -ItemType Directory -Force -Path $releaseRoot | Out-Null
$stageRoot = Join-Path $releaseRoot ('.electron-build-' + [guid]::NewGuid().ToString('N'))
$stageRepo = Join-Path $stageRoot 'repo'
$stageFrontend = Join-Path $stageRepo 'frontend'
$stageBackendOutput = Join-Path $stageRepo 'dist\windows-release'
$bundleStage = Join-Path $stageRoot 'partflow-bundle'
$bundlePath = Join-Path $releaseRoot 'partflow-bundle'
$backupBundlePath = $null

try {
  New-Item -ItemType Directory -Force -Path $stageFrontend, $stageBackendOutput, $bundleStage | Out-Null

  $copyArguments = @(
    $frontendSource,
    $stageFrontend,
    '/E',
    '/COPY:DAT',
    '/R:2',
    '/W:1',
    '/XD',
    (Join-Path $frontendSource 'node_modules'),
    (Join-Path $frontendSource 'dist'),
    (Join-Path $frontendSource 'dist-electron-internal'),
    (Join-Path $frontendSource 'playwright-report'),
    (Join-Path $frontendSource 'test-results'),
    '/XF',
    '.env',
    '.env.*',
    '*.db',
    '*.sqlite',
    '*.sqlite3',
    '*.log'
  )
  & robocopy.exe @copyArguments | Out-Null
  if ($LASTEXITCODE -ge 8) {
    throw "Could not stage the frontend sources (robocopy exit code $LASTEXITCODE)."
  }

  $backendExecutable = Join-Path $stageBackendOutput 'partflow-api.exe'
  Push-Location $backendSource
  try {
    & go.exe build -trimpath -ldflags '-s -w' -o $backendExecutable ./cmd/api
    if ($LASTEXITCODE -ne 0) {
      throw "Backend build failed with exit code $LASTEXITCODE."
    }
  }
  finally {
    Pop-Location
  }

  $backendSignature = Set-AuthenticodeSignature -FilePath $backendExecutable -Certificate $signingCertificate -HashAlgorithm SHA256
  if ($backendSignature.Status -ne [System.Management.Automation.SignatureStatus]::Valid) {
    throw "Backend signing failed: $($backendSignature.StatusMessage)"
  }

  Push-Location $stageFrontend
  try {
    & npm.cmd ci
    if ($LASTEXITCODE -ne 0) {
      throw "npm ci failed with exit code $LASTEXITCODE."
    }

    & npm.cmd version $Version --no-git-tag-version --allow-same-version
    if ($LASTEXITCODE -ne 0) {
      throw "Could not set the staged frontend version to $Version."
    }

    if (-not $SkipTests) {
      & npm.cmd run test:run
      if ($LASTEXITCODE -ne 0) {
        throw "Frontend tests failed with exit code $LASTEXITCODE."
      }
    }

    & npm.cmd run build:check
    if ($LASTEXITCODE -ne 0) {
      throw "Frontend type-check or production build failed with exit code $LASTEXITCODE."
    }

    & .\node_modules\.bin\electron-builder.cmd --win --x64 --publish never
    if ($LASTEXITCODE -ne 0) {
      throw "Electron Builder failed with exit code $LASTEXITCODE."
    }
  }
  finally {
    Pop-Location
  }

  $electronOutput = Join-Path $stageFrontend 'dist-electron-internal'
  $setupPath = Join-Path $electronOutput "PartFlow-$Version-setup.exe"
  $portablePath = Join-Path $electronOutput "PartFlow-$Version-portable.exe"
  $unpackedPath = Join-Path $electronOutput 'win-unpacked'
  $unpackedAppPath = Join-Path $unpackedPath 'PartFlow.exe'
  $unpackedResourcesPath = Join-Path $unpackedPath 'resources'
  $appArchivePath = Join-Path $unpackedResourcesPath 'app.asar'
  $unpackedBackendPath = Join-Path $unpackedPath 'resources\backend\partflow-api.exe'

  foreach ($artifactPath in @($setupPath, $portablePath, $unpackedAppPath, $unpackedBackendPath)) {
    if (-not (Test-Path -LiteralPath $artifactPath -PathType Leaf)) {
      throw "Electron Builder output is missing: $artifactPath"
    }
  }

  $requiredPackagedResources = @(
    $appArchivePath,
    (Join-Path $unpackedResourcesPath 'PartFlow-Internal-Code-Signing.cer'),
    (Join-Path $unpackedResourcesPath 'partflow-logo.png'),
    (Join-Path $unpackedResourcesPath 'partflow-logo.ico'),
    (Join-Path $unpackedResourcesPath 'backend\partflow-api.exe')
  )
  foreach ($resourcePath in $requiredPackagedResources) {
    if (-not (Test-Path -LiteralPath $resourcePath -PathType Leaf)) {
      throw "Electron package is missing a required runtime resource: $resourcePath"
    }
  }

  $asarValidationCode = @'
const asar = require("@electron/asar");
const archivePath = process.argv[2];
const entries = new Set(
  asar.listPackage(archivePath)
    .map((entry) => entry.replaceAll("\\", "/").replace(/^\/+/, "")),
);
const requiredEntries = [
  "electron/main.js",
  "electron/preload.js",
  "dist/index.html",
  "dist/fonts/NotoNaskhArabic.ttf",
];
const missingEntries = requiredEntries.filter((entry) => !entries.has(entry));
if (missingEntries.length) {
  console.error("app.asar is missing required entries: " + missingEntries.join(", "));
  process.exit(1);
}

const mainSource = asar.extractFile(archivePath, "electron/main.js").toString("utf8");
const requiredPathContracts = [
  "app.setPath('userData', path.join(app.getPath('appData'), 'PartFlow'))",
  "getUserDataPath('data', 'partflow.db')",
  "getUserDataPath('data', 'product-images')",
  "getUserDataPath('data', 'part-type-images')",
  "getUserDataPath('data', 'category-images')",
  "getUserDataPath('logs', 'backend.log')",
  "getUserDataPath('offline-grant-public-key.txt')",
  "PARTFLOW_LOCAL_DB_PATH: getLocalDatabasePath()",
  "path.join(process.resourcesPath, 'backend', 'partflow-api.exe')",
];
const missingContracts = requiredPathContracts.filter((contract) => !mainSource.includes(contract));
if (missingContracts.length) {
  console.error("Electron runtime path contracts are missing: " + missingContracts.join(", "));
  process.exit(1);
}

const stylesheetEntry = asar.listPackage(archivePath).find((entry) =>
  /^dist\/assets\/index-[^/]+\.css$/.test(entry.replaceAll("\\", "/").replace(/^\/+/, "")),
);
if (!stylesheetEntry) {
  console.error("app.asar is missing the main application stylesheet.");
  process.exit(1);
}
const stylesheetArchivePath = stylesheetEntry.startsWith("\\") || stylesheetEntry.startsWith("/")
  ? stylesheetEntry.slice(1)
  : stylesheetEntry;
const stylesheet = asar.extractFile(archivePath, stylesheetArchivePath).toString("utf8");
const requiredMaxWidthRules = [
  ".max-w-\\[20rem\\]{max-width:20rem}",
  ".max-w-\\[24rem\\]{max-width:24rem}",
  ".max-w-\\[28rem\\]{max-width:28rem}",
  ".max-w-\\[32rem\\]{max-width:32rem}",
  ".max-w-\\[36rem\\]{max-width:36rem}",
  ".max-w-\\[42rem\\]{max-width:42rem}",
  ".max-w-\\[56rem\\]{max-width:56rem}",
  ".max-w-\\[64rem\\]{max-width:64rem}",
  ".max-w-\\[80rem\\]{max-width:80rem}",
];
const missingMaxWidthRules = requiredMaxWidthRules.filter((rule) => !stylesheet.includes(rule));
if (missingMaxWidthRules.length) {
  console.error("Dialog width rules are missing or malformed: " + missingMaxWidthRules.join(", "));
  process.exit(1);
}

console.log("Validated Electron ASAR entries, persistent paths, and dialog width rules.");
'@
  Push-Location $stageFrontend
  try {
    $asarValidationScriptPath = Join-Path $stageFrontend '.validate-electron-package.cjs'
    [System.IO.File]::WriteAllText(
      $asarValidationScriptPath,
      $asarValidationCode,
      [System.Text.UTF8Encoding]::new($false)
    )
    & node.exe $asarValidationScriptPath $appArchivePath
    if ($LASTEXITCODE -ne 0) {
      throw "Packaged Electron path validation failed with exit code $LASTEXITCODE."
    }
  }
  finally {
    Pop-Location
  }

  foreach ($signedPath in @($setupPath, $portablePath, $unpackedAppPath, $unpackedBackendPath)) {
    $signature = Get-AuthenticodeSignature -FilePath $signedPath
    if ($signature.Status -ne [System.Management.Automation.SignatureStatus]::Valid) {
      throw "Signature validation failed for '$signedPath': $($signature.StatusMessage)"
    }
  }

  Copy-Item -LiteralPath $setupPath -Destination $bundleStage
  Copy-Item -LiteralPath $portablePath -Destination $bundleStage
  Copy-Item -LiteralPath $backendExecutable -Destination (Join-Path $bundleStage 'partflow-api.exe')
  Copy-Item -LiteralPath $unpackedPath -Destination $bundleStage -Recurse
  Copy-Item -LiteralPath (Join-Path $repoRoot 'docs\SUBSCRIPTION-OFFLINE-STRATEGY.md') -Destination (Join-Path $bundleStage 'SUBSCRIPTION-OFFLINE-STRATEGY.md')

  @'
@echo off
start "" "%~dp0win-unpacked\PartFlow.exe"
'@ | Set-Content -LiteralPath (Join-Path $bundleStage 'launch-partflow.bat') -Encoding UTF8

  @"
PartFlow $Version for Windows x64

Run PartFlow-$Version-setup.exe to install, or PartFlow-$Version-portable.exe for portable use.
The embedded backend is partflow-api.exe and also appears in win-unpacked\resources\backend.
The SQLite database is %APPDATA%\PartFlow\data\partflow.db.
Persistent images use %APPDATA%\PartFlow\data\product-images, category-images, and part-type-images.
Backend logs use %APPDATA%\PartFlow\logs\backend.log.
For offline access, the public key file is %APPDATA%\PartFlow\offline-grant-public-key.txt.
User data remains under %APPDATA%\PartFlow and is not removed by uninstall.
For offline access, install the public offline-grant key as documented in SUBSCRIPTION-OFFLINE-STRATEGY.md.
This internally signed package trusts the PartFlow internal certificate on the current Windows user only.
"@ | Set-Content -LiteralPath (Join-Path $bundleStage 'README.txt') -Encoding UTF8

  $commit = (& git.exe -C $repoRoot rev-parse --short HEAD 2>$null | Out-String).Trim()
  if ($LASTEXITCODE -ne 0) { $commit = 'unknown' }
  $gitStatus = @(& git.exe -C $repoRoot status --porcelain --untracked-files=all 2>$null)
  if ($LASTEXITCODE -ne 0) {
    throw 'Could not determine whether the source working tree contains uncommitted changes.'
  }
  $workingTreeDirty = $gitStatus.Count -gt 0

  $manifestFiles = @(
    (Join-Path $bundleStage "PartFlow-$Version-setup.exe"),
    (Join-Path $bundleStage "PartFlow-$Version-portable.exe"),
    (Join-Path $bundleStage 'partflow-api.exe'),
    (Join-Path $bundleStage 'win-unpacked\PartFlow.exe'),
    (Join-Path $bundleStage 'win-unpacked\resources\app.asar'),
    (Join-Path $bundleStage 'win-unpacked\resources\PartFlow-Internal-Code-Signing.cer'),
    (Join-Path $bundleStage 'win-unpacked\resources\partflow-logo.png'),
    (Join-Path $bundleStage 'win-unpacked\resources\partflow-logo.ico'),
    (Join-Path $bundleStage 'win-unpacked\resources\backend\partflow-api.exe')
  ) | ForEach-Object {
    $hash = Get-FileHash -LiteralPath $_ -Algorithm SHA256
    [ordered]@{
      path = $_.Substring($bundleStage.Length + 1).Replace('\', '/')
      sha256 = $hash.Hash
      bytes = (Get-Item -LiteralPath $_).Length
    }
  }

  [ordered]@{
    product = 'PartFlow'
    version = $Version
    architecture = 'x64'
    builtAtUtc = [DateTime]::UtcNow.ToString('o')
    sourceCommit = $commit
    workingTreeDirty = $workingTreeDirty
    signingCertificate = $publicCertificate.Subject
    files = @($manifestFiles)
  } | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath (Join-Path $bundleStage 'release-manifest.json') -Encoding UTF8

  if (Test-Path -LiteralPath $bundlePath) {
    $backupBundlePath = Join-Path $releaseRoot ("partflow-bundle.previous-{0}-{1}" -f (Get-Date -Format 'yyyyMMdd-HHmmss'), $PID)
    Move-Item -LiteralPath $bundlePath -Destination $backupBundlePath
  }

  try {
    Move-Item -LiteralPath $bundleStage -Destination $bundlePath
  }
  catch {
    if ($backupBundlePath -and (Test-Path -LiteralPath $backupBundlePath) -and -not (Test-Path -LiteralPath $bundlePath)) {
      Move-Item -LiteralPath $backupBundlePath -Destination $bundlePath
    }
    throw
  }

  Write-Output "Windows Electron release bundle prepared: $bundlePath"
}
finally {
  if (Test-Path -LiteralPath $stageRoot) {
    $resolvedStage = [System.IO.Path]::GetFullPath($stageRoot)
    $resolvedReleaseRoot = [System.IO.Path]::GetFullPath($releaseRoot).TrimEnd('\') + '\'
    if (-not $resolvedStage.StartsWith($resolvedReleaseRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
      throw "Refusing to remove build staging outside the release directory: $resolvedStage"
    }
    Remove-Item -LiteralPath $resolvedStage -Recurse -Force
  }
}
