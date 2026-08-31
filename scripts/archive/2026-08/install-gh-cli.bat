@echo off
echo =========================================
echo GitHub CLI Installation Script
echo =========================================
echo.

REM Check if gh is already installed
where gh >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo GitHub CLI is already installed!
    gh --version
    echo.
    echo To login, run: gh auth login
    pause
    exit /b 0
)

REM Check if Chocolatey is installed
where choco >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo Chocolatey not found. Installing Chocolatey...
    echo This requires Administrator privileges.
    echo.

    REM Check if running as Administrator
    net session >nul 2>&1
    if %ERRORLEVEL% NEQ 0 (
        echo This script requires Administrator privileges to install Chocolatey.
        echo Please run this script as Administrator.
        echo.
        echo Alternatively, you can manually install GitHub CLI from:
        echo https://github.com/cli/cli/releases/latest
        pause
        exit /b 1
    )

    REM Install Chocolatey
    powershell -NoProfile -ExecutionPolicy Bypass -Command "Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))"

    echo Chocolatey installed!
    echo.
)

REM Install GitHub CLI using Chocolatey
echo Installing GitHub CLI using Chocolatey...
choco install gh -y

echo Installation completed!
echo.
echo Please restart your terminal for the changes to take effect.
echo Then run: gh auth login
echo.
pause
