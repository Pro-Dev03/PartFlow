#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FRONTEND_DIR="$ROOT_DIR/frontend"
BACKEND_DIR="$ROOT_DIR/backend"
DIST_DIR="$ROOT_DIR/dist/windows"

mkdir -p "$DIST_DIR"

echo "[1/3] Building Go backend for Windows from Linux..."
cd "$BACKEND_DIR"
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o "$DIST_DIR/partflow-api.exe" ./cmd/api

echo "[2/3] Installing frontend dependencies..."
cd "$FRONTEND_DIR"
if [ -f package-lock.json ]; then
  npm ci --no-fund --no-audit
else
  npm install --no-fund --no-audit
fi

echo "[3/3] Building frontend and preparing Windows desktop bundle..."
npm run build
npx electron-builder --win --dir --publish never

echo "Windows-capable artifacts generated in: $DIST_DIR"
