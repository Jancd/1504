@echo off
title Video Generator - Launcher
echo ========================================
echo Video Generator Development Launcher
echo ========================================
echo.

REM Check environment
if not exist "go.mod" (
    echo ERROR: Please run this script from the project root directory
    echo Current directory: %CD%
    pause
    exit /b 1
)

echo Environment check...
go version >nul 2>&1 || (
    echo ERROR: Go is not installed or not in PATH
    echo Please install Go 1.21+ from https://golang.org/dl/
    pause
    exit /b 1
)

node --version >nul 2>&1 || (
    echo ERROR: Node.js is not installed or not in PATH  
    echo Please install Node.js 16+ from https://nodejs.org/
    pause
    exit /b 1
)

echo OK: Environment ready
echo.

echo Choose startup method:
echo.
echo 1. Auto start (recommended) - Opens both services in new windows
echo 2. Manual start - Start backend and frontend separately  
echo 3. Backend only - Start only the backend service
echo 4. Frontend only - Start only the frontend service
echo 5. Exit
echo.
set /p choice="Enter your choice (1-5): "

if "%choice%"=="1" goto auto_start
if "%choice%"=="2" goto manual_start  
if "%choice%"=="3" goto backend_only
if "%choice%"=="4" goto frontend_only
if "%choice%"=="5" goto exit
goto invalid_choice

:auto_start
echo.
echo Starting both services automatically...
echo.
start "Backend Service" cmd /c "scripts\start-backend.bat"
timeout /t 3 /nobreak >nul
start "Frontend Service" cmd /c "scripts\start-frontend.bat"
timeout /t 3 /nobreak >nul
start http://localhost:3000
echo.
echo Services started! Check the opened windows.
echo Backend: http://localhost:8080
echo Frontend: http://localhost:3000
goto end

:manual_start
echo.
echo Manual startup instructions:
echo.
echo 1. Open first terminal and run: scripts\start-backend.bat
echo 2. Open second terminal and run: scripts\start-frontend.bat  
echo 3. Open browser to: http://localhost:3000
echo.
goto end

:backend_only
echo.
echo Starting backend service only...
call scripts\start-backend.bat
goto end

:frontend_only
echo.
echo Starting frontend service only...
echo Make sure backend is running at http://localhost:8080
call scripts\start-frontend.bat
goto end

:invalid_choice
echo Invalid choice. Please try again.
pause
goto start

:exit
echo Goodbye!
goto end

:end
echo.
echo Press any key to close this window...
pause >nul