# Push Local Repository to GitHub
param (
    [Parameter(Mandatory=$false)]
    [string]$RepoUrl
)

$GitExe = "git"
if (Test-Path "C:\Program Files\Git\cmd\git.exe") {
    $GitExe = "C:\Program Files\Git\cmd\git.exe"
} elseif (Test-Path (Join-Path $PSScriptRoot "tools\git\cmd\git.exe")) {
    $GitExe = Join-Path $PSScriptRoot "tools\git\cmd\git.exe"
}

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "  Live Polling Tool - Push to GitHub      " -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

if (-not $RepoUrl) {
    $RepoUrl = Read-Host "`nEnter your GitHub Repository URL (e.g., https://github.com/username/live-polling.git)"
}

if (-not $RepoUrl) {
    Write-Host "No repository URL provided. Aborting." -ForegroundColor Red
    exit 1
}

Write-Host "`n[*] Configuring remote origin: $RepoUrl" -ForegroundColor Yellow
& $GitExe remote remove origin 2>$null
& $GitExe remote add origin $RepoUrl

Write-Host "[*] Pushing 'main' branch to GitHub..." -ForegroundColor Yellow
& $GitExe push -u origin main

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n[OK] Successfully pushed to GitHub!" -ForegroundColor Green
    Write-Host "Your public repository is live at: $RepoUrl" -ForegroundColor Cyan
} else {
    Write-Host "`n[!] Push failed. Please verify that:" -ForegroundColor Red
    Write-Host "1. The repository exists on GitHub and is public." -ForegroundColor White
    Write-Host "2. You are logged into GitHub in your browser/credential manager." -ForegroundColor White
}
