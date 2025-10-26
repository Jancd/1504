@echo off
title Frontend Service - Video Generator
echo ========================================
echo Starting Frontend Service  
echo ========================================
echo.

if not exist "frontend\package.json" (
    echo ERROR: Please run from project root directory
    pause
    exit /b 1
)

cd frontend

echo Checking dependencies...
if not exist "node_modules" (
    echo Installing dependencies...
    npm install
    if %errorlevel% neq 0 (
        echo ERROR: Failed to install dependencies
        pause
        exit /b 1
    )
)

echo.
echo Frontend starting at: http://localhost:3000
echo.
echo Press Ctrl+C to stop the service
echo.

npm run dev

echo.
echo Frontend service stopped.
pause