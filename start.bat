@echo off
echo Video Generator Startup
echo.

echo Checking environment...
go version || (echo Go not found && pause && exit)
node --version || (echo Node not found && pause && exit)

echo.
echo Starting backend...
start "Backend" cmd /k "go run cmd/server/main.go"

echo Waiting 5 seconds...
timeout /t 5 /nobreak >nul

echo Starting frontend...
cd frontend
start "Frontend" cmd /k "npm run dev"
cd ..

echo.
echo Services starting...
echo Backend: http://localhost:8080  
echo Frontend: http://localhost:3000
echo.
echo Opening browser in 3 seconds...
timeout /t 3 /nobreak >nul
start http://localhost:3000

echo.
echo Done! Check the opened windows.
pause