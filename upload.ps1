# ECH-Workers 上传到 GitHub

# 初始化 git（如果还没有）
if (-not (Test-Path .git)) {
    git init
    git branch -M main
}

# 添加所有文件
Write-Host "添加文件到 Git..." -ForegroundColor Green
git add .

# 提交
$commitMessage = Read-Host "请输入提交信息 (默认: Update)"
if ([string]::IsNullOrWhiteSpace($commitMessage)) {
    $commitMessage = "Update"
}
git commit -m $commitMessage

# 设置远程仓库（如果还没有）
$remote = git remote get-url origin 2>$null
if (-not $remote) {
    $repoUrl = Read-Host "请输入 GitHub 仓库地址 (如: https://github.com/username/ech-workers.git)"
    git remote add origin $repoUrl
}

# 推送
Write-Host "推送到 GitHub..." -ForegroundColor Green
git push -u origin main

Write-Host "`n上传完成！" -ForegroundColor Green
Write-Host "`n如需创建 Release 版本，请执行:" -ForegroundColor Yellow
Write-Host "  git tag -a v1.0.0 -m 'Release version 1.0.0'" -ForegroundColor Cyan
Write-Host "  git push origin v1.0.0" -ForegroundColor Cyan
Write-Host "`nGitHub Actions 将自动构建 Android APK 和 OpenWrt IPK 包" -ForegroundColor Yellow
