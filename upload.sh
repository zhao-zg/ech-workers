#!/bin/bash

# ECH-Workers 上传到 GitHub

# 初始化 git（如果还没有）
if [ ! -d .git ]; then
    git init
    git branch -M main
fi

# 添加所有文件
echo "添加文件到 Git..."
git add .

# 提交
read -p "请输入提交信息 (默认: Update): " commit_message
commit_message=${commit_message:-Update}
git commit -m "$commit_message"

# 设置远程仓库（如果还没有）
remote=$(git remote get-url origin 2>/dev/null)
if [ -z "$remote" ]; then
    read -p "请输入 GitHub 仓库地址 (如: https://github.com/username/ech-workers.git): " repo_url
    git remote add origin "$repo_url"
fi

# 推送
echo "推送到 GitHub..."
git push -u origin main

echo ""
echo "上传完成！"
echo ""
echo "如需创建 Release 版本，请执行:"
echo "  git tag -a v1.0.0 -m 'Release version 1.0.0'"
echo "  git push origin v1.0.0"
echo ""
echo "GitHub Actions 将自动构建 Android APK 和 OpenWrt IPK 包"
