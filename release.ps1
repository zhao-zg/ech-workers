# 发布新版本的脚本
param(
    [Parameter(Mandatory=$true)]
    [string]$Version,
    
    [string]$Message = ""
)

# 确保版本号格式正确
if (-not $Version.StartsWith("v")) {
    $Version = "v$Version"
}

# 验证版本号格式 (v1.2.3)
if ($Version -notmatch '^v\d+\.\d+\.\d+$') {
    Write-Error "版本号格式错误！应该是 v1.2.3 格式"
    exit 1
}

Write-Host "准备发布版本: $Version" -ForegroundColor Green

# 检查是否有未提交的更改
$gitStatus = git status --porcelain
if ($gitStatus) {
    Write-Host "发现未提交的更改:" -ForegroundColor Yellow
    Write-Host $gitStatus
    $response = Read-Host "是否继续？(y/n)"
    if ($response -ne "y") {
        exit 0
    }
}

# 如果没有提供消息，使用默认消息
if ([string]::IsNullOrWhiteSpace($Message)) {
    $Message = "$Version 版本发布"
}

# 创建并推送 tag
Write-Host "`n创建 git tag..." -ForegroundColor Cyan
git tag -a $Version -m $Message

if ($LASTEXITCODE -ne 0) {
    Write-Error "创建 tag 失败！"
    Write-Host "如果 tag 已存在，可以使用以下命令删除：" -ForegroundColor Yellow
    Write-Host "  git tag -d $Version" -ForegroundColor Yellow
    Write-Host "  git push origin :refs/tags/$Version" -ForegroundColor Yellow
    exit 1
}

Write-Host "`n推送 tag 到 GitHub..." -ForegroundColor Cyan
git push origin $Version

if ($LASTEXITCODE -ne 0) {
    Write-Error "推送 tag 失败！"
    exit 1
}

Write-Host "`n✅ 版本 $Version 发布成功！" -ForegroundColor Green
Write-Host "`nGitHub Actions 将自动构建以下内容：" -ForegroundColor Cyan
Write-Host "  - Android APK (多架构)"
Write-Host "  - OpenWrt IPK 包"
Write-Host "`n查看构建进度：" -ForegroundColor Cyan
Write-Host "  https://github.com/zhao-zg/ech-workers/actions" -ForegroundColor Blue
Write-Host "`n查看 Release：" -ForegroundColor Cyan
Write-Host "  https://github.com/zhao-zg/ech-workers/releases/tag/$Version" -ForegroundColor Blue
