@echo off
setlocal
title Java Version Manager Installer

:: 1. Admin check
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo Requesting Administrator Privileges...
    powershell -Command "Start-Process '%~f0' -Verb RunAs"
    exit /b
)

cd /d "%~dp0"
if not exist "jvm.exe" (
    echo [ERROR] jvm.exe not found!
    pause
    exit /b
)

echo ========================================
echo   Java Version Manager Installer
echo ========================================
echo.
echo Initializing environment...
.\jvm.exe setup

if %errorlevel% neq 0 (
    echo.
    powershell -Command "Write-Host '[FAILED] Initialization failed.' -ForegroundColor Red"
) else (
    echo.
    powershell -Command "Write-Host '[SUCCESS] JVM environment configured!' -ForegroundColor Green"
    echo Please RESTART your terminal/CMD windows.
)

echo.
pause
