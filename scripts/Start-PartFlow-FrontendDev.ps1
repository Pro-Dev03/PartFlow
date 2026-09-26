$ErrorActionPreference = 'Stop'
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$frontendDir = Join-Path $repoRoot 'frontend'

$env:VITE_DEPLOYMENT_MODE = 'development'
$env:VITE_DEVELOPMENT_API_URL = 'http://127.0.0.1:8080/api/v1'
$env:VITE_API_BASE_URL_LOCAL = 'http://127.0.0.1:8080/api/v1'
$env:VITE_CLOUD_API_URL = 'https://partflow-api.onrender.com/api/v1'

try {
    Set-Location $frontendDir
    npm.cmd run dev -- --host 127.0.0.1
} finally {
    Set-Location $repoRoot
    @('VITE_DEPLOYMENT_MODE', 'VITE_DEVELOPMENT_API_URL', 'VITE_API_BASE_URL_LOCAL', 'VITE_CLOUD_API_URL') |
        ForEach-Object { Remove-Item "Env:$_" -ErrorAction SilentlyContinue }
}
