param(
    [switch]$AllowDatabaseWrites
)

$ErrorActionPreference = 'Stop'
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$backendDir = Join-Path $repoRoot 'backend'

if ($AllowDatabaseWrites) {
    $confirmation = Read-Host 'This connects PartFlow to the REAL store database with writes enabled. Type REAL-DATABASE-WRITES to continue'
    if ($confirmation -cne 'REAL-DATABASE-WRITES') {
        throw 'Live database startup cancelled.'
    }
    $readOnly = 'false'
    Write-Host 'Starting local API with real database writes enabled; automatic schema changes remain disabled.' -ForegroundColor Yellow
} else {
    $readOnly = 'true'
    Write-Host 'Starting local API in read-only database mode. Login/session operations that write to the database will fail.' -ForegroundColor Yellow
}

$secureDatabaseUrl = Read-Host 'Enter the rotated Supabase PostgreSQL URL (input is hidden)' -AsSecureString
$databaseUrl = ([System.Net.NetworkCredential]::new('', $secureDatabaseUrl)).Password
if ($databaseUrl -notmatch '^postgres(ql)?://') {
    $databaseUrl = $null
    throw 'Expected a PostgreSQL connection URL.'
}

$localDataDir = Join-Path $env:TEMP 'PartFlow-LocalLive'
$env:APP_ENV = 'development'
$env:SERVER_MODE = 'debug'
$env:SERVER_HOST = '127.0.0.1'
$env:SERVER_PORT = '8080'
$env:DB_CONNECTION_MODE = 'cloud'
$env:DATABASE_URL = $databaseUrl
$env:PARTFLOW_DISABLE_SCHEMA_AUTO_ENSURE = 'true'
$env:PARTFLOW_DB_READ_ONLY = $readOnly
$env:PARTFLOW_LOCAL_DB_PATH = Join-Path $localDataDir 'device-data.sqlite'
$env:JWT_SECRET = [Convert]::ToBase64String([Security.Cryptography.RandomNumberGenerator]::GetBytes(48))
$env:DISABLE_AUTH = 'false'
$env:PARTFLOW_REQUIRE_CLOUD_AUTH = 'true'
$env:PARTFLOW_ALLOW_LOCAL_AUTH_BYPASS = 'false'
$env:USE_SUPABASE_AUTH = 'false'
$env:CLOUD_API_URL = 'https://partflow-api.onrender.com/api/v1'
$env:CORS_ALLOWED_ORIGINS = 'https://partflow-hpv7.onrender.com,http://localhost:5173,http://localhost:5174,http://localhost:5175,http://127.0.0.1:5173,http://127.0.0.1:5174,http://127.0.0.1:5175'
$env:LOG_LEVEL = 'debug'

try {
    Set-Location $backendDir
    Write-Host 'Local API: http://127.0.0.1:8080/api/v1'
    Write-Host 'Cloud login and subscription checks: https://partflow-api.onrender.com/api/v1'
    go run ./cmd/api
} finally {
    Set-Location $repoRoot
    @(
        'APP_ENV', 'SERVER_MODE', 'SERVER_HOST', 'SERVER_PORT', 'DB_CONNECTION_MODE',
        'DATABASE_URL', 'PARTFLOW_DISABLE_SCHEMA_AUTO_ENSURE', 'PARTFLOW_DB_READ_ONLY',
        'PARTFLOW_LOCAL_DB_PATH', 'JWT_SECRET', 'DISABLE_AUTH', 'PARTFLOW_REQUIRE_CLOUD_AUTH',
        'PARTFLOW_ALLOW_LOCAL_AUTH_BYPASS', 'USE_SUPABASE_AUTH', 'CLOUD_API_URL',
        'CORS_ALLOWED_ORIGINS', 'LOG_LEVEL'
    ) | ForEach-Object { Remove-Item "Env:$_" -ErrorAction SilentlyContinue }
    $databaseUrl = $null
    $secureDatabaseUrl = $null
}
