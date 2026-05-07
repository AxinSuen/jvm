@echo off
setlocal
title Java Version Manager Installer

:: Admin check
net session >nul 2>&1
if %errorLevel% neq 0 (
    powershell -Command "Write-Host '[Prompt] Requesting administrator privileges...'"
    powershell -Command "Start-Process '%~f0' -Verb RunAs"
    exit /b
)

cd /d "%~dp0"

:: Banner
powershell -Command "Write-Host '========================================' -ForegroundColor Cyan"
powershell -Command "Write-Host '  Java Version Manager Installer' -ForegroundColor Cyan"
powershell -Command "Write-Host '========================================' -ForegroundColor Cyan"

powershell -Command "Write-Host 'Initializing environment...'"
.\jvm.exe setup

if %errorlevel% neq 0 goto :failed

powershell -Command "Write-Host '[Success] JVM environment configuration completed!' -ForegroundColor Green"
powershell -Command "Write-Host 'Please restart your terminal or command line window to take effect.'"
goto :end

:failed
powershell -Command "Write-Host '[Failed] Initialization failed.' -ForegroundColor Red"

:end
echo.
pause
