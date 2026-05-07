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

powershell -Command "Write-Host ([char]0x6B63+[char]0x5728+[char]0x521D+[char]0x59CB+[char]0x5316+[char]0x73AF+[char]0x5883+[char]0x2e+[char]0x2e+[char]0x2e)"
.\jvm.exe setup

if %errorlevel% neq 0 goto :failed

powershell -Command "Write-Host ([char]0x5B+[char]0x6210+[char]0x529F+[char]0x5D+' JVM '+[char]0x73AF+[char]0x5883+[char]0x914D+[char]0x7F6E+[char]0x5B8C+[char]0x6210+[char]0xFF01) -ForegroundColor Green"
powershell -Command "Write-Host ([char]0x8BF7+[char]0x91CD+[char]0x542F+[char]0x7EC8+[char]0x7AEF+[char]0x6216+[char]0x547D+[char]0x4EE4+[char]0x884C+[char]0x7A97+[char]0x53E3+[char]0x4EE5+[char]0x751F+[char]0x6548+[char]0x3002)"
goto :end

:failed
powershell -Command "Write-Host ([char]0x5B+[char]0x5931+[char]0x8D25+[char]0x5D+' '+[char]0x521D+[char]0x59CB+[char]0x5316+[char]0x5931+[char]0x8D25+[char]0x3002) -ForegroundColor Red"

:end
echo.
pause
