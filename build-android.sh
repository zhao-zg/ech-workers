#!/bin/bash

# Android APK 构建脚本
# 解决版本号和AAR缺失问题

set -e

echo "========================================="
echo "Android APK 构建脚本"
echo "========================================="

# 1. 检查依赖
echo "检查构建依赖..."

# 检查Go
if ! command -v go &> /dev/null; then
    echo "错误: 未找到Go，请先安装Go"
    exit 1
fi

# 检查Java
if ! command -v java &> /dev/null; then
    echo "错误: 未找到Java，请先安装JDK 17"
    exit 1
fi

# 检查Android SDK
if [ -z "$ANDROID_HOME" ]; then
    echo "错误: 未设置ANDROID_HOME环境变量"
    exit 1
fi

# 2. 设置版本号
VERSION=${1:-$(date +%Y.%m.%d)}
echo "使用版本号: $VERSION"
export APP_VERSION=$VERSION
export GITHUB_REF_NAME=$VERSION

# 3. 安装gomobile (如果需要)
echo "检查gomobile..."
if ! command -v gomobile &> /dev/null; then
    echo "安装gomobile..."
    go install golang.org/x/mobile/cmd/gomobile@latest
    go install golang.org/x/mobile/cmd/gobind@latest
    gomobile init
fi

# 4. 准备Go模块
echo "准备Go模块..."
if [ ! -f go.work ]; then
    go work init
    go work use . ./tunnel ./mobile
fi

cd mobile
go mod tidy
cd ..

# 5. 构建AAR
echo "构建AAR文件..."
cd mobile
gomobile bind -v -target=android -androidapi=24 -o ../android/ech-workers.aar .
cd ..

# 验证AAR文件
if [ ! -f android/ech-workers.aar ]; then
    echo "错误: AAR文件构建失败"
    exit 1
fi

echo "AAR文件构建成功: $(ls -lh android/ech-workers.aar)"

# 6. 构建APK
echo "构建APK..."
cd android
chmod +x gradlew
./gradlew clean
./gradlew assembleDebug -PAPP_VERSION=$VERSION

# 7. 检查输出
echo "构建完成！"
echo "APK文件位置:"
find build/outputs/apk -name "*.apk" -exec ls -lh {} \;

echo "========================================="
echo "构建成功完成！"
echo "版本号: $VERSION"
echo "========================================="