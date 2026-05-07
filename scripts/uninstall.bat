@echo off
setlocal
title Java Version Manager Uninstaller

:: Admin check
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo Requesting Administrator Privileges...
    powershell -Command "Start-Process '%~f0' -Verb RunAs"
    exit /b
)

cd /d "%~dp0"
echo ========================================
echo   Java Version Manager Uninstaller
echo ========================================
echo.
echo Cleaning up system environment variables...

set "PS_FILE=%temp%\jvm_cleanup.ps1"
echo $ErrorActionPreference = 'Stop' > "%PS_FILE%"
echo $currentDir = '%~dp0' >> "%PS_FILE%"
echo try { >> "%PS_FILE%"
echo     Write-Host '1. Removing JAVA_HOME...' -ForegroundColor Green >> "%PS_FILE%"
echo     [Environment]::SetEnvironmentVariable('JAVA_HOME', $null, 'Machine') >> "%PS_FILE%"
echo     Write-Host '2. Cleaning up System PATH entries...' -ForegroundColor Green >> "%PS_FILE%"
echo     $path = [Environment]::GetEnvironmentVariable('Path', 'Machine') >> "%PS_FILE%"
echo     $pathArray = $path -split ';' >> "%PS_FILE%"
echo     $newPathArray = $pathArray ^| Where-Object { $_ -ne '%%JAVA_HOME%%\bin' -and $_ -ne $currentDir -and $_ -ne $currentDir.TrimEnd('\') -and $_ -notlike '*\jvm\current*' -and $_ -ne '' } >> "%PS_FILE%"
echo     $newPath = $newPathArray -join ';' >> "%PS_FILE%"
echo     [Environment]::SetEnvironmentVariable('Path', $newPath, 'Machine') >> "%PS_FILE%"
echo     Write-Host '3. Deleting settings.txt...' -ForegroundColor Green >> "%PS_FILE%"
echo     $settingsFile = Join-Path $currentDir 'settings.txt' >> "%PS_FILE%"
echo     if (Test-Path $settingsFile) { Remove-Item $settingsFile -Force } >> "%PS_FILE%"
echo } catch { >> "%PS_FILE%"
echo     Write-Host \"`n[ERROR] Cleanup failed: $($_.Exception.Message)\" -ForegroundColor Red >> "%PS_FILE%"
echo     exit 1 >> "%PS_FILE%"
echo } >> "%PS_FILE%"

powershell -ExecutionPolicy Bypass -File "%PS_FILE%"
set "EXIT_CODE=%errorlevel%"
if exist "%PS_FILE%" del "%PS_FILE%"

:: Fallback deletion using native CMD
if exist "%~dp0settings.txt" del /f /q "%~dp0settings.txt" >nul 2>&1

if %EXIT_CODE% neq 0 (
    echo.
    powershell -Command "Write-Host '[FAILED] Cleanup process failed. Please check your permissions or antivirus.' -ForegroundColor Red"
) else (
    echo.
    powershell -Command "Write-Host '[CLEANUP SUCCESS] All system associations have been removed.' -ForegroundColor Green"
    echo You can now manually delete this folder.
)

pause
