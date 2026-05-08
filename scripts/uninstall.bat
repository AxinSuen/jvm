@echo off
setlocal
title Java Version Manager Uninstaller

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
set "MSG_BANNER=Write-Host '========================================' -ForegroundColor Cyan"
set "MSG_TITLE=Write-Host '  Java Version Manager Uninstaller' -ForegroundColor Cyan"
set "MSG_CLEAN=Write-Host 'Cleaning up system environment variables...'"
set "MSG_SUCCESS=Write-Host '[Cleanup Success] All system associations have been removed.' -ForegroundColor Green"
set "MSG_MANUAL=Write-Host 'You can now manually delete this folder.'"
set "MSG_FAIL=Write-Host '[Failed] Error occurred during cleanup.' -ForegroundColor Red"

powershell -Command "%MSG_BANNER%"
powershell -Command "%MSG_TITLE%"
powershell -Command "%MSG_BANNER%"

powershell -Command "%MSG_CLEAN%"

:: Generate temp ps1 script
set "PS_FILE=%temp%\jvm_cleanup.ps1"
echo $ErrorActionPreference = 'Stop' > "%PS_FILE%"
echo $currentDir = '%~dp0' >> "%PS_FILE%"
echo try { >> "%PS_FILE%"
echo     $oldJavaHome = [Environment]::GetEnvironmentVariable('JAVA_HOME', 'Machine') >> "%PS_FILE%"
echo     $oldJavaBin = if ($oldJavaHome) { Join-Path $oldJavaHome 'bin' } else { 'NOMATCH_STRING_XYZZY' } >> "%PS_FILE%"
echo     Write-Host ('1. Removing JAVA_HOME...') -ForegroundColor Green >> "%PS_FILE%"
echo     [Environment]::SetEnvironmentVariable('JAVA_HOME', $null, 'Machine') >> "%PS_FILE%"
echo     Write-Host ('2. Cleaning up PATH...') -ForegroundColor Green >> "%PS_FILE%"
echo     $path = [Environment]::GetEnvironmentVariable('Path', 'Machine') >> "%PS_FILE%"
echo     $pathArray = $path -split ';' >> "%PS_FILE%"
echo     $newPathArray = $pathArray ^| Where-Object { $_ -ne '%%JAVA_HOME%%\bin' -and $_ -ne '%%JAVA_HOME%%' -and $_ -ne $oldJavaBin -and $_ -ne $currentDir -and $_ -ne $currentDir.TrimEnd('\') -and $_ -notlike '*\jvm\current*' -and $_ -ne '' } >> "%PS_FILE%"
echo     $newPath = $newPathArray -join ';' >> "%PS_FILE%"
echo     [Environment]::SetEnvironmentVariable('Path', $newPath, 'Machine') >> "%PS_FILE%"
echo     Write-Host ('3. Deleting settings.txt...') -ForegroundColor Green >> "%PS_FILE%"
echo     $settingsFile = Join-Path $currentDir 'settings.txt' >> "%PS_FILE%"
echo     if (Test-Path $settingsFile) { Remove-Item $settingsFile -Force } >> "%PS_FILE%"
echo } catch { >> "%PS_FILE%"
echo     Write-Host ('`n[Error] Cleanup failed') -ForegroundColor Red >> "%PS_FILE%"
echo     exit 1 >> "%PS_FILE%"
echo } >> "%PS_FILE%"

powershell -ExecutionPolicy Bypass -File "%PS_FILE%"
set "EXIT_CODE=%errorlevel%"
if exist "%PS_FILE%" del "%PS_FILE%"

if exist "%~dp0settings.txt" del /f /q "%~dp0settings.txt" >nul 2>&1

if %EXIT_CODE% neq 0 goto :failed

powershell -Command "%MSG_SUCCESS%"
powershell -Command "%MSG_MANUAL%"
goto :end

:failed
powershell -Command "Write-Host '[Failed] Cleanup failed' -ForegroundColor Red"
goto :end

:end
echo.
echo Press any key to exit...
pause > nul
