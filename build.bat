@echo off
setlocal
title Java Version Manager Builder

REM 1. Start Build
powershell -Command "Write-Host '========================================' -ForegroundColor Cyan; Write-Host ([char]0x6B63+[char]0x5728+[char]0x6784+[char]0x5EFA+' Java Version Manager...') -ForegroundColor Cyan; Write-Host '========================================' -ForegroundColor Cyan"

REM 2. Create bin directory
if not exist "bin" mkdir bin

REM 3. Compile Go source
powershell -Command "Write-Host ([char]0x5B+ '1/4' + [char]0x5D + ' ' + [char]0x6B63+[char]0x5728+[char]0x7F16+[char]0x8BD1+' Go '+[char]0x6E90+[char]0x7801+'...')"
go build -o bin/jvm.exe src/main.go
if %errorlevel% neq 0 (
    powershell -Command "Write-Host ([char]0x6784+[char]0x5EFA+[char]0x5931+[char]0x8D25+[char]0xFF01) -ForegroundColor Red"
    echo.
    powershell -Command "Write-Host ([char]0x8BF7+[char]0x6309+[char]0x56DE+[char]0x8F66+[char]0x952E+[char]0x9000+[char]0x51FA+[char]0x2e+[char]0x2e+[char]0x2e)"
    pause > nul
    exit /b
)

REM 4. Copy distribution scripts
powershell -Command "Write-Host ([char]0x5B+ '2/4' + [char]0x5D + ' ' + [char]0x6B63+[char]0x5728+[char]0x590D+[char]0x5236+[char]0x5206+[char]0x53D1+[char]0x811A+[char]0x672C+'...')"
copy /y scripts\install.bat bin\ >nul
copy /y scripts\uninstall.bat bin\ >nul

REM 5. Copy Documentation
powershell -Command "Write-Host ([char]0x5B+ '3/4' + [char]0x5D + ' ' + [char]0x6B63+[char]0x5728+[char]0x590D+[char]0x5236+[char]0x6587+[char]0x6863+'...')"
if exist "README.txt" copy /y "README.txt" bin\ >nul
if exist "README.md" copy /y "README.md" bin\ >nul
if exist "LICENSE" copy /y "LICENSE" bin\ >nul

REM 6. 复制更新日志
powershell -Command "Write-Host ([char]0x5B+ '4/4' + [char]0x5D + ' ' + [char]0x6B63+[char]0x5728+[char]0x590D+[char]0x5236+[char]0x66F4+[char]0x65B0+[char]0x65E5+[char]0x5FD7+'...')"
if exist "updateLog.txt" copy /y "updateLog.txt" bin\ >nul

REM 7. 构建成功
powershell -Command "Write-Host '' ; Write-Host '========================================' -ForegroundColor Green; Write-Host ([char]0x6784+[char]0x5EFA+[char]0x6210+[char]0x529F+[char]0xFF01) -ForegroundColor Green; Write-Host ([char]0x8BF7+[char]0x68C0+[char]0x67E5+' \"bin\" '+[char]0x76EE+[char]0x5F55+[char]0x3002) -ForegroundColor Green; Write-Host '========================================' -ForegroundColor Green"
echo.
powershell -Command "Write-Host ([char]0x8BF7+[char]0x6309+[char]0x56DE+[char]0x8F66+[char]0x952E+[char]0x9000+[char]0x51FA+[char]0x2e+[char]0x2e+[char]0x2e)"
pause > nul
