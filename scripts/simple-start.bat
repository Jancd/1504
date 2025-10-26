@echo off
echo Simple startup test
echo.

echo Step 1: Check directory
dir go.mod >nul 2>&1
if %errorlevel% equ 0 (
    echo OK: In project directory
) else (
    echo ERROR: Not in project directory
    pause
    exit
)

echo.
echo Step 2: Check Go
go version
echo Go check result: %errorlevel%

echo.
echo Step 3: Check Node
node --version  
echo Node check result: %errorlevel%

echo.
echo Step 4: Start backend manually
echo Starting: go run cmd/server/main.go
echo Press Ctrl+C to stop, then start frontend separately
go run cmd/server/main.go