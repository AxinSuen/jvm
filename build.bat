@echo off
setlocal
title JVM Project Builder

echo ========================================
echo   Java Version Manager - Building...
echo ========================================
echo.

:: 1. Create bin directory if not exists
if not exist "bin" mkdir bin

:: 2. Compile Go source
echo [1/3] Compiling Go source...
go build -o bin/jvm.exe src/main.go
if %errorlevel% neq 0 (
    echo.
    echo [ERROR] Compilation failed!
    pause
    exit /b
)

:: 3. Copy scripts from scripts/ to bin/
echo [2/3] Copying distribution scripts...
copy /y scripts\install.bat bin\ >nul
copy /y scripts\uninstall.bat bin\ >nul

:: 4. Copy Documentation (Safe Copy)
echo [3/3] Copying documentation...
if exist "README.txt" copy /y "README.txt" bin\ >nul
if exist "README.md" copy /y "README.md" bin\ >nul

echo.
echo ========================================
echo   BUILD SUCCESSFUL!
echo   Check the "bin" folder for your package.
echo ========================================
echo.
pause
