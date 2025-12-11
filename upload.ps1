# ECH-Workers Upload to GitHub Script

# Initialize git repository if not exists
if (-not (Test-Path .git)) {
    Write-Host "Initializing git repository..." -ForegroundColor Cyan
    git init
    git branch -M main
}

# Add all files
Write-Host "Adding files to Git..." -ForegroundColor Green
git add .

# Commit
$commitMessage = Read-Host "Enter commit message (default: Update)"
if ([string]::IsNullOrWhiteSpace($commitMessage)) {
    $commitMessage = "Update"
}
git commit -m $commitMessage

# Check remote repository
$remote = git remote get-url origin 2>$null
if (-not $remote) {
    Write-Host "`nNo remote repository configured." -ForegroundColor Yellow
    $repoUrl = Read-Host "Enter GitHub repository URL (e.g., https://github.com/username/ech-workers.git)"
    git remote add origin $repoUrl
}

# Push to GitHub
Write-Host "`nPushing to GitHub..." -ForegroundColor Green
git push -u origin main

Write-Host "`n=== Upload Complete! ===" -ForegroundColor Green
Write-Host "`nTo create a release version, run:" -ForegroundColor Yellow
Write-Host "  git tag -a v1.0.0 -m 'Release version 1.0.0'" -ForegroundColor Cyan
Write-Host "  git push origin v1.0.0" -ForegroundColor Cyan
Write-Host "`nGitHub Actions will automatically build Android APK and OpenWrt IPK packages." -ForegroundColor Yellow
