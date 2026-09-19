# Run Local Development Environment for Live Polling Tool
$ErrorActionPreference = "Continue"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "  Live Polling Tool — Local Launcher      " -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

$WorkspaceRoot = $PSScriptRoot
$GoPath = Join-Path $WorkspaceRoot "tools\go\bin"
$RedisPath = Join-Path $WorkspaceRoot "tools\redis"
$MongoPath = Join-Path $WorkspaceRoot "tools\mongo"

# Add Go to PATH
if (Test-Path $GoPath) {
    $env:PATH = "$GoPath;$env:PATH"
    Write-Host "[✓] Go toolchain configured" -ForegroundColor Green
}

# 1. Start Redis
$redisRunning = $false
try {
    $res = & "$RedisPath\redis-cli.exe" ping 2>$null
    if ($res -match "PONG") { $redisRunning = $true }
} catch {}

if (-not $redisRunning -and (Test-Path "$RedisPath\redis-server.exe")) {
    Write-Host "[*] Starting Redis Server on port 6379..." -ForegroundColor Yellow
    Start-Process -FilePath "$RedisPath\redis-server.exe" -ArgumentList "--port 6379" -WindowStyle Hidden
    Start-Sleep -Seconds 1
} else {
    Write-Host "[✓] Redis is already active on port 6379" -ForegroundColor Green
}

# 2. Start MongoDB
$mongoRunning = $false
try {
    $tcp = New-Object System.Net.Sockets.TcpClient
    $tcp.Connect("127.0.0.1", 27017)
    $mongoRunning = $tcp.Connected
    $tcp.Close()
} catch {}

if (-not $mongoRunning -and (Test-Path "$MongoPath\bin\mongod.exe")) {
    Write-Host "[*] Starting MongoDB Daemon on port 27017..." -ForegroundColor Yellow
    $dataPath = Join-Path $MongoPath "data"
    if (-not (Test-Path $dataPath)) { New-Item -ItemType Directory -Force -Path $dataPath | Out-Null }
    Start-Process -FilePath "$MongoPath\bin\mongod.exe" -ArgumentList "--dbpath `"$dataPath`" --port 27017 --bind_ip 127.0.0.1" -WindowStyle Hidden
    Start-Sleep -Seconds 2
} else {
    Write-Host "[✓] MongoDB is already active on port 27017" -ForegroundColor Green
}

# 3. Start Backend Server
Write-Host "[*] Starting Go + Gin Backend on http://localhost:8080..." -ForegroundColor Cyan
$backendDir = Join-Path $WorkspaceRoot "backend"
Start-Process powershell -ArgumentList "-NoExit", "-Command", "`$env:PATH = '$GoPath;' + `$env:PATH; cd '$backendDir'; ./server.exe" -WindowStyle Normal

# 4. Start Frontend
Write-Host "[*] Starting React Frontend on http://localhost:5173..." -ForegroundColor Cyan
$frontendDir = Join-Path $WorkspaceRoot "frontend"
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$frontendDir'; npm.cmd run dev" -WindowStyle Normal

Write-Host "`nAll services launched successfully!" -ForegroundColor Green
Write-Host "Frontend: http://localhost:5173" -ForegroundColor White
Write-Host "Backend API: http://localhost:8080/api" -ForegroundColor White
Write-Host "WebSocket endpoint: ws://localhost:8080/ws/polls/:id" -ForegroundColor White
