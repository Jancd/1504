# Video Generator Development Environment Startup Script
# PowerShell version for better Windows compatibility

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Starting Video Generator Dev Environment" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if in project root directory
if (-not (Test-Path "go.mod")) {
    Write-Host "ERROR: Please run this script from project root directory" -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}

# Check Go installation
try {
    $null = Get-Command go -ErrorAction Stop
    Write-Host "OK: Go is installed" -ForegroundColor Green
} catch {
    Write-Host "ERROR: Go is not installed. Please install Go 1.21+" -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}

# Check Node.js installation
try {
    $null = Get-Command node -ErrorAction Stop
    Write-Host "OK: Node.js is installed" -ForegroundColor Green
} catch {
    Write-Host "ERROR: Node.js is not installed. Please install Node.js" -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}

# Check npm installation
try {
    $null = Get-Command npm -ErrorAction Stop
    Write-Host "OK: npm is installed" -ForegroundColor Green
} catch {
    Write-Host "ERROR: npm is not installed. Please install npm" -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}

Write-Host ""

# Install frontend dependencies if needed
if (-not (Test-Path "frontend\node_modules")) {
    Write-Host "Installing frontend dependencies..." -ForegroundColor Yellow
    Set-Location frontend
    npm install
    Set-Location ..
    Write-Host "Frontend dependencies installed" -ForegroundColor Green
    Write-Host ""
}

Write-Host "Starting backend service..." -ForegroundColor Yellow
Write-Host "Backend URL: http://localhost:8080"
Write-Host "Health check: http://localhost:8080/health"
Write-Host ""

# Start backend in new window
Start-Process powershell -ArgumentList "-NoExit", "-Command", "go run cmd/server/main.go" -WindowStyle Normal

# Wait for backend to start
Write-Host "Waiting for backend service to start..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# Check if backend started successfully
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -TimeoutSec 5 -ErrorAction Stop
    Write-Host "OK: Backend service started successfully" -ForegroundColor Green
} catch {
    Write-Host "WARNING: Backend service may have failed to start, check backend window" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Starting frontend service..." -ForegroundColor Yellow
Write-Host "Frontend URL: http://localhost:3000"
Write-Host ""

# Start frontend service
Set-Location frontend
Start-Process powershell -ArgumentList "-NoExit", "-Command", "npm run dev" -WindowStyle Normal
Set-Location ..

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Development environment started!" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Service Information:" -ForegroundColor White
Write-Host "  Backend Service: http://localhost:8080" -ForegroundColor Gray
Write-Host "  Frontend UI: http://localhost:3000" -ForegroundColor Gray
Write-Host "  Health Check: http://localhost:8080/health" -ForegroundColor Gray
Write-Host ""
Write-Host "Usage Instructions:" -ForegroundColor White
Write-Host "  1. Open browser and visit http://localhost:3000" -ForegroundColor Gray
Write-Host "  2. Enter novel text in the text box" -ForegroundColor Gray
Write-Host "  3. Click 'Generate Video' to start creation" -ForegroundColor Gray
Write-Host "  4. View task progress on the right side" -ForegroundColor Gray
Write-Host "  5. Click 'Download' when completed" -ForegroundColor Gray
Write-Host ""
Write-Host "Important Notes:" -ForegroundColor White
Write-Host "  - Make sure OpenAI API Key is configured" -ForegroundColor Gray
Write-Host "  - Qiniu mode requires Qiniu API Key" -ForegroundColor Gray
Write-Host "  - Local SD mode requires Stable Diffusion service" -ForegroundColor Gray
Write-Host ""
Write-Host "Opening frontend in browser..." -ForegroundColor Yellow

# Wait a moment then open browser
Start-Sleep -Seconds 3
Start-Process "http://localhost:3000"

Write-Host ""
Write-Host "SUCCESS: Development environment is ready!" -ForegroundColor Green
Write-Host "NOTE: Closing this window will NOT stop the services" -ForegroundColor Yellow
Write-Host "      Please manually close backend and frontend windows when done" -ForegroundColor Yellow
Read-Host "Press Enter to exit this window"