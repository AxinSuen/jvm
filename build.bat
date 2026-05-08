@echo off
setlocal
title Java Version Manager Builder

REM 1. Start Build
powershell -Command "Write-Host '========================================' -ForegroundColor Cyan; Write-Host ' Building Java Version Manager...' -ForegroundColor Cyan; Write-Host '========================================' -ForegroundColor Cyan"

REM 2. Create bin directory
if not exist "bin" mkdir bin

REM 3. Compile Go source
powershell -Command "Write-Host '[1/4] Compiling Go source...'"
go build -o bin/jvm.exe src/main.go
if %errorlevel% neq 0 (
    powershell -Command "Write-Host 'Build failed!' -ForegroundColor Red"
    echo.
    echo Press any key to exit...
    pause > nul
    exit /b
)

REM 4. Copy distribution scripts
powershell -Command "Write-Host '[2/4] Copying distribution scripts...'"
copy /y scripts\install.bat bin\install.bat >nul
copy /y scripts\uninstall.bat bin\uninstall.bat >nul

REM 5. Copy Documentation
powershell -Command "Write-Host '[3/4] Copying documentation...'"
if exist "README.txt" copy /y "README.txt" bin\ >nul
if exist "README.md" copy /y "README.md" bin\ >nul
if exist "LICENSE" copy /y "LICENSE" bin\ >nul

REM 6. Copy Update Log
powershell -Command "Write-Host '[4/4] Copying update log...'"
if exist "updateLog.txt" copy /y "updateLog.txt" bin\ >nul

REM 7. Build Successful
powershell -Command "Write-Host '' ; Write-Host '========================================' -ForegroundColor Green; Write-Host ' Build Successful!' -ForegroundColor Green; Write-Host ' Please check the \"bin\" directory.' -ForegroundColor Green; Write-Host '========================================' -ForegroundColor Green"
echo.
echo Press any key to exit...
pause > nul
