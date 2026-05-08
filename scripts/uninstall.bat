@echo off
setlocal
title Java Version Manager Uninstaller

:: Admin check
if "%1"=="--elevated" goto :has_admin
fltmc >nul 2>&1
if %errorLevel% neq 0 (
    powershell -Command "Write-Host ([char]0x5B+[char]0x63D0+[char]0x793A+[char]0x5D+' '+[char]0x6B63+[char]0x5728+[char]0x83B7+[char]0x53D6+[char]0x7BA1+[char]0x7406+[char]0x5458+[char]0x6743+[char]0x9650+'...')"
    powershell -Command "Start-Process '%~f0' -ArgumentList '--elevated' -Verb RunAs"
    exit /b
)
:has_admin

cd /d "%~dp0"
set "MSG_BANNER=Write-Host '========================================' -ForegroundColor Cyan"
set "MSG_TITLE=Write-Host '  Java Version Manager Uninstaller' -ForegroundColor Cyan"
set "MSG_CLEAN=Write-Host ([char]0x6B63+[char]0x5728+[char]0x6E05+[char]0x7406+[char]0x7CFB+[char]0x7EDF+[char]0x73AF+[char]0x5883+[char]0x53D8+[char]0x91CF+'...')"
set "MSG_SUCCESS=Write-Host ([char]0x5B+[char]0x6E05+[char]0x7406+[char]0x6210+[char]0x529F+[char]0x5D+' '+[char]0x6240+[char]0x6709+[char]0x7CFB+[char]0x7EDF+[char]0x5173+[char]0x8054+[char]0x5DF2+[char]0x79FB+[char]0x9664+[char]0x3002) -ForegroundColor Green"
set "MSG_MANUAL=Write-Host ([char]0x60A8+[char]0x73B0+[char]0x5728+[char]0x53EF+[char]0x4EE5+[char]0x624B+[char]0x52A8+[char]0x5220+[char]0x9664+[char]0x6B64+[char]0x6587+[char]0x4EF6+[char]0x5939+[char]0x3002)"
set "MSG_FAIL=Write-Host ([char]0x5B+[char]0x5931+[char]0x8D25+[char]0x5D+' '+[char]0x6E05+[char]0x7406+[char]0x8FC7+[char]0x7A0B+[char]0x51FA+[char]0x9519+[char]0x3002) -ForegroundColor Red"

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
echo     Write-Host ([char]0x31 + '. ' + [char]0x6b63+[char]0x5728+[char]0x79fb+[char]0x9664 + ' JAVA_HOME...') -ForegroundColor Green >> "%PS_FILE%"
echo     [Environment]::SetEnvironmentVariable('JAVA_HOME', $null, 'Machine') >> "%PS_FILE%"
echo     Write-Host ([char]0x32 + '. ' + [char]0x6b63+[char]0x5728+[char]0x6e05+[char]0x7406+[char]0x7cfb+[char]0x7edf + ' PATH...') -ForegroundColor Green >> "%PS_FILE%"
echo     $path = [Environment]::GetEnvironmentVariable('Path', 'Machine') >> "%PS_FILE%"
echo     $pathArray = $path -split ';' >> "%PS_FILE%"
echo     $newPathArray = $pathArray ^| Where-Object { $_ -ne '%%JAVA_HOME%%\bin' -and $_ -ne '%%JAVA_HOME%%' -and $_ -ne $oldJavaBin -and $_ -ne $currentDir -and $_ -ne $currentDir.TrimEnd('\') -and $_ -notlike '*\jvm\current*' -and $_ -ne '' } >> "%PS_FILE%"
echo     $newPath = $newPathArray -join ';' >> "%PS_FILE%"
echo     [Environment]::SetEnvironmentVariable('Path', $newPath, 'Machine') >> "%PS_FILE%"
echo     Write-Host ([char]0x33 + '. ' + [char]0x6b63+[char]0x5728+[char]0x5220+[char]0x9664 + ' settings.txt...') -ForegroundColor Green >> "%PS_FILE%"
echo     $settingsFile = Join-Path $currentDir 'settings.txt' >> "%PS_FILE%"
echo     if (Test-Path $settingsFile) { Remove-Item $settingsFile -Force } >> "%PS_FILE%"
echo } catch { >> "%PS_FILE%"
echo     Write-Host ([char]0x0a+[char]0x5b+[char]0x9519+[char]0x8bef+[char]0x5d+[char]0x20+[char]0x6e05+[char]0x7406+[char]0x5931+[char]0x8d25) -ForegroundColor Red >> "%PS_FILE%"
echo     exit 1 >> "%PS_FILE%"
echo } >> "%PS_FILE%"

powershell -ExecutionPolicy Bypass -File "%PS_FILE%"
set "EXIT_CODE=%errorlevel%"
if exist "%PS_FILE%" del "%PS_FILE%"

if exist "%~dp0settings.txt" del /f /q "%~dp0settings.txt" >nul 2>&1

if %EXIT_CODE% neq 0 goto :failed

powershell -Command "%MSG_SUCCESS%"
powershell -Command "%MSG_MANUAL%"
echo.
powershell -Command "Write-Host ([char]0x8BF7+[char]0x6309+[char]0x56DE+[char]0x8F66+[char]0x952E+[char]0x9000+[char]0x51FA+[char]0x2e+[char]0x2e+[char]0x2e)"
pause > nul
goto :end

:failed
powershell -Command "%MSG_FAIL%"
echo.
powershell -Command "Write-Host ([char]0x8BF7+[char]0x6309+[char]0x56DE+[char]0x8F66+[char]0x952E+[char]0x9000+[char]0x51FA+[char]0x2e+[char]0x2e+[char]0x2e)"
pause > nul

:end
echo.
pause
