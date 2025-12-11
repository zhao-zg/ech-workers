# ============================================
# ECH-Workers GitHub Upload Script (CN)
# ============================================

Write-Host @"

 _____ ____ _   _      __        __         _
| ____/ ___| | | |     \ \      / /__  _ __| | _____ _ __ ___
|  _|| |   | |_| |_____\ \ /\ / / _ \| '__| |/ / _ \ '__/ __|
| |__| |___|  _  |______\ V  V / (_) | |  |   <  __/ |  \__ \
|_____\____|_| |_|       \_/\_/ \___/|_|  |_|\_\___|_|  |___/

GitHub Upload Tool
"@ -ForegroundColor Cyan

# Step 1: Initialize git
if (-not (Test-Path .git)) {
    Write-Host "[1/5] Chu shi hua Git..." -ForegroundColor Yellow
    git init
    git branch -M main
} else {
    Write-Host "[1/5] Git yi cun zai" -ForegroundColor Green
}

# Step 2: Add files
Write-Host "[2/5] Tian jia wen jian..." -ForegroundColor Yellow
git add .

# Step 3: Commit
Write-Host "[3/5] Ti jiao wen jian..." -ForegroundColor Yellow
$commitMessage = Read-Host "Qing shu ru ti jiao xin xi (mo ren: Update)"
if ([string]::IsNullOrWhiteSpace($commitMessage)) {
    $commitMessage = "Update"
}
git commit -m $commitMessage

# Step 4: Check remote
Write-Host "[4/5] Jian cha yuan cheng cang ku..." -ForegroundColor Yellow
$remote = git remote get-url origin 2>$null
if (-not $remote) {
    Write-Host "Wei zhao dao yuan cheng cang ku!" -ForegroundColor Red
    $repoUrl = Read-Host "Qing shu ru GitHub di zhi (li: https://github.com/username/ech-workers.git)"
    git remote add origin $repoUrl
} else {
    Write-Host "Yuan cheng cang ku: $remote" -ForegroundColor Green
}

# Step 5: Push
Write-Host "[5/5] Shang chuan dao GitHub..." -ForegroundColor Yellow
git push -u origin main

Write-Host @"

================================
    Shang chuan wan cheng!
================================

"@ -ForegroundColor Green

Write-Host "Fa bu xin ban ben:" -ForegroundColor Cyan
Write-Host "  git tag -a v1.0.0 -m 'Release version 1.0.0'" -ForegroundColor White
Write-Host "  git push origin v1.0.0" -ForegroundColor White
Write-Host ""
Write-Host "GitHub Actions hui zi dong gou jian:" -ForegroundColor Cyan
Write-Host "  - Android APK" -ForegroundColor White
Write-Host "  - OpenWrt IPK" -ForegroundColor White
Write-Host ""

