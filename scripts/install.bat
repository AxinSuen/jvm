@echo off
setlocal
title Java Version Manager Installer

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

:: Banner
powershell -Command "Write-Host '========================================' -ForegroundColor Cyan"
powershell -Command "Write-Host '  Java Version Manager Installer' -ForegroundColor Cyan"
powershell -Command "Write-Host '========================================' -ForegroundColor Cyan"

.\jvm.exe setup

if %errorlevel% neq 0 goto :failed
goto :end

:failed
powershell -Command "Write-Host ([char]0x5B+[char]0x5931+[char]0x8D25+[char]0x5D+' '+[char]0x521D+[char]0x59CB+[char]0x5316+[char]0x5931+[char]0x8D25+[char]0x3002) -ForegroundColor Red"

:end
echo.
powershell -Command "Write-Host ([char]0x8BF7+[char]0x6309+[char]0x56DE+[char]0x8F66+[char]0x952E+[char]0x9000+[char]0x51FA+[char]0x2e+[char]0x2e+[char]0x2e)"
pause > nul
