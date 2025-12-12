# Android 构建问题完整解决方案

## 问题总结

你遇到的两个主要问题：

1. **版本号缺失**: GitHub Actions 构建的APK文件名没有版本号
2. **APK白屏**: 安装后打开应用显示白屏，无任何内容

## 根本原因分析

### 版本号问题
- `build.gradle` 中的版本号获取逻辑过于复杂，依赖多个可能不存在的源
- 在某些构建环境中，所有版本号源都获取失败，导致使用默认值

### 白屏问题
- 缺少关键的 `ech-workers.aar` 文件
- 这个AAR包含了Go编译的native库，是应用核心功能
- 没有这个文件，应用无法正常初始化，导致白屏

## 解决方案

### 方案1: 使用自动构建脚本（推荐）

我已经创建了完整的构建脚本：

**Windows用户:**
```cmd
# 使用默认版本号（当前日期）
build-android.bat

# 指定版本号
build-android.bat 1.2.3
```

**Linux/Mac用户:**
```bash
# 给脚本执行权限
chmod +x build-android.sh

# 使用默认版本号
./build-android.sh

# 指定版本号
./build-android.sh 1.2.3
```

### 方案2: 手动步骤

如果你想了解详细过程或脚本执行失败：

#### 1. 安装依赖

**安装Go (如果没有):**
- 下载: https://golang.org/dl/
- 版本要求: >= 1.19

**安装gomobile:**
```bash
go install golang.org/x/mobile/cmd/gomobile@latest
go install golang.org/x/mobile/cmd/gobind@latest
gomobile init
```

**安装Android SDK和NDK:**
- Android Studio 或 command line tools
- NDK 26.3.11579264
- 设置 `ANDROID_HOME` 环境变量

#### 2. 准备Go工作空间

```bash
# 在项目根目录
go work init
go work use . ./tunnel ./mobile
```

#### 3. 构建AAR文件

```bash
cd mobile
go mod tidy
gomobile bind -target=android -androidapi=24 -o ../android/ech-workers.aar .
cd ..
```

验证AAR文件：
```bash
ls -lh android/ech-workers.aar
```

#### 4. 构建APK

```bash
cd android
chmod +x gradlew  # Linux/Mac only

# 清理之前的构建
./gradlew clean

# 构建APK并指定版本号
./gradlew assembleDebug -PAPP_VERSION=1.2.3
```

### 方案3: 修复GitHub Actions

如果你想修复CI构建，需要确保Actions中正确设置了版本号环境变量。

在 `.github/workflows/build-android.yml` 中，确保这部分正确：

```yaml
- name: Extract version from git
  id: version
  run: |
    if [[ "${{ github.ref }}" == refs/tags/* ]]; then
      VERSION=$(echo "${{ github.ref }}" | sed 's/refs\/tags\/v//')
    else
      VERSION=$(git describe --tags --always 2>/dev/null | sed 's/^v//' || echo "$(date +%Y.%m.%d)")
    fi
    echo "version=$VERSION" >> $GITHUB_OUTPUT
    echo "APP_VERSION=$VERSION" >> $GITHUB_ENV
    echo "GITHUB_REF_NAME=$VERSION" >> $GITHUB_ENV
    echo "Extracted version: $VERSION"
```

## 验证构建结果

### 检查版本号
构建完成后，APK文件名应该包含版本号：
```
com.ech.workers-1.2.3-universal.apk
```

### 检查APK内容
验证APK包含必要的文件：
```bash
unzip -l android/build/outputs/apk/debug/*.apk | grep -E '\.so$|classes.dex'
```

应该看到：
- `classes.dex` (Java代码)
- `lib/*/libgojni.so` (Go native库)

### 测试安装
1. 安装APK到设备
2. 打开应用，应该看到正常的UI界面
3. 检查logcat输出，确认没有严重错误

## 常见问题排查

### Q: gomobile 安装失败
```bash
# 设置Go代理
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn

# 重新安装
go install golang.org/x/mobile/cmd/gomobile@latest
```

### Q: AAR构建失败
```bash
# 检查Go模块
cd mobile
go mod tidy
go mod download

# 检查依赖
go list -m all
```

### Q: Gradle构建失败
```bash
# 清理Gradle缓存
cd android
rm -rf .gradle build
./gradlew clean

# 重新构建
./gradlew assembleDebug --stacktrace
```

### Q: 应用仍然白屏
1. 检查logcat输出：`adb logcat | grep -i ech`
2. 确认AAR文件存在且大小正常
3. 检查应用权限设置

## 最终检查清单

- [ ] Go 已安装 (>= 1.19)
- [ ] gomobile 已安装并初始化
- [ ] Android SDK/NDK 已安装
- [ ] `ech-workers.aar` 文件存在
- [ ] APK构建成功且包含版本号
- [ ] APK安装后UI正常显示

按照这个方案，你的Android构建问题应该能够完全解决。