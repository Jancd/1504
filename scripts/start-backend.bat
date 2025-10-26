@echo off
title Backend Service - Video Generator
echo ========================================
echo Starting Backend Service
echo ========================================
echo.

if not exist "go.mod" (
    echo ERROR: Please run from project root directory
    pause
    exit /b 1
)

echo Backend starting at: http://localhost:8080
echo Health check: http://localhost:8080/health
echo.
echo Press Ctrl+C to stop the service
echo.

go run cmd/server/main.go

echo.
echo Backend service stopped.
pause