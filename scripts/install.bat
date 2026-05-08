@echo off
setlocal
title Java Version Manager Installer

:: Admin check
if "%1"=="--elevated" goto :has_admin
fltmc >nul 2>&1
if %errorLevel% neq 0 (
    powershell -Command "Write-Host '[Prompt] Requesting administrator privileges...'"
    powershell -Command "Start-Process '%~f0' -ArgumentList '--elevated' -Verb RunAs"
    exit /b
)
:has_admin

cd /d "%~dp0"

:: Banner
powershell -Command "Write-Host '========================================' -ForegroundColor Cyan"
powershell -Command "Write-Host '  Java Version Manager Installer' -ForegroundColor Cyan"
powershell -Command "Write-Host '========================================' -ForegroundColor Cyan"

.\jvm.exe setup

if %errorlevel% neq 0 goto :failed
goto :end

:failed
powershell -Command "Write-Host '[Failed] Initialization failed.' -ForegroundColor Red"
echo.
echo Press any key to exit...
pause > nul

:end
