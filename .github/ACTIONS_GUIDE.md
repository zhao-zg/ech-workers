# GitHub Actions 自动构建配置

本项目使用 GitHub Actions 自动构建 Android APK 和 OpenWrt IPK 软件包。

## 工作流说明

### 1. Android 构建 (build-android.yml)

**触发条件:**
- 推送到 main/master 分支
- 创建以 `v` 开头的标签（如 `v1.0.0`）
- Pull Request 到 main/master 分支
- 手动触发

**构建步骤:**
1. 设置 JDK 17 环境
2. 设置 Go 1.24 环境
3. 安装 Android SDK 和 NDK 26.3.11579264
4. 安装 gomobile 并初始化
5. 编译 Go 库为 Android AAR
6. 使用 Gradle 构建 APK
7. 对于 tag 推送，自动签名 APK（需要配置密钥）
8. 上传 APK 到 Artifacts
9. 对于 tag 推送，自动创建 GitHub Release

**输出文件:**
- `ech-workers-signed.apk` - 已签名版本（需要配置密钥）
- `ech-workers-unsigned.apk` - 未签名版本

### 2. OpenWrt 构建 (build-openwrt.yml)

**触发条件:**
- 推送到 main/master 分支
- 创建以 `v` 开头的标签
- Pull Request 到 main/master 分支
- 手动触发

**构建目标:**
- x86_64 - x86 64位平台
- aarch64_generic - ARM64 平台
- mipsel_24kc - MIPS 平台（常见于路由器）

**构建步骤:**
1. 设置 Go 1.24 环境
2. 下载 OpenWrt SDK
3. 更新 feeds
4. 复制软件包源代码
5. 编译 ech-workers 和 luci-app-ech-workers
6. 上传 IPK 到 Artifacts
7. 对于 tag 推送，自动创建 GitHub Release

**输出文件:**
- `ech-workers_*.ipk` - 核心程序包（针对每个架构）
- `luci-app-ech-workers_*.ipk` - LuCI Web 界面包

### 3. 发布工作流 (release.yml)

**触发条件:**
- 创建以 `v` 开头的标签

**功能:**
- 自动创建 GitHub Release
- 调用 Android 和 OpenWrt 构建工作流
- 将所有构建产物附加到 Release

## 配置 APK 签名（可选）

如果需要发布已签名的 APK，需要在 GitHub 仓库中配置以下 Secrets：

1. 进入仓库的 Settings → Secrets and variables → Actions
2. 添加以下 secrets：

| Secret 名称 | 说明 |
|------------|------|
| `KEYSTORE_BASE64` | Keystore 文件的 Base64 编码 |
| `KEYSTORE_PASSWORD` | Keystore 密码 |
| `KEY_ALIAS` | 密钥别名 |
| `KEY_PASSWORD` | 密钥密码 |

### 生成 Keystore Base64 编码

**Windows PowerShell:**
```powershell
# 生成 keystore（如果还没有）
keytool -genkey -v -keystore release.keystore -alias your_alias -keyalg RSA -keysize 2048 -validity 10000

# 转换为 Base64
$bytes = [System.IO.File]::ReadAllBytes("release.keystore")
[Convert]::ToBase64String($bytes) | Out-File -Encoding ASCII keystore.base64.txt
```

**Linux/macOS:**
```bash
# 生成 keystore（如果还没有）
keytool -genkey -v -keystore release.keystore -alias your_alias -keyalg RSA -keysize 2048 -validity 10000

# 转换为 Base64
base64 release.keystore > keystore.base64.txt
```

然后将 `keystore.base64.txt` 的内容复制到 `KEYSTORE_BASE64` secret 中。

## 发布新版本

### 1. 准备发布

确保所有代码已提交并推送到 main 分支：

```bash
git add .
git commit -m "准备发布 v1.0.0"
git push origin main
```

### 2. 创建并推送标签

```bash
# 创建标签
git tag -a v1.0.0 -m "Release version 1.0.0"

# 推送标签到 GitHub
git push origin v1.0.0
```

### 3. 自动构建

推送标签后，GitHub Actions 会自动：
1. 创建 Release
2. 构建 Android APK（所有架构）
3. 构建 OpenWrt IPK（x86_64, aarch64, mipsel）
4. 将所有文件上传到 Release

### 4. 编辑 Release 说明（可选）

构建完成后，可以在 GitHub Releases 页面编辑发布说明，添加更新日志等信息。

## 手动触发构建

可以在 GitHub 仓库的 Actions 标签页手动触发构建：

1. 进入 Actions 标签
2. 选择要运行的工作流（Build Android APK 或 Build OpenWrt Package）
3. 点击 "Run workflow"
4. 选择分支并点击运行

## 查看构建结果

### Artifacts

每次构建都会生成 Artifacts，保留 30 天：

1. 进入 Actions 标签
2. 点击具体的工作流运行
3. 在页面底部找到 Artifacts 部分
4. 下载需要的文件

### Releases

对于标签推送，构建产物会自动发布到 Releases：

1. 进入仓库的 Releases 页面
2. 找到对应版本
3. 下载 Assets 中的文件

## 构建状态徽章

可以在 README.md 中添加构建状态徽章：

```markdown
![Android Build](https://github.com/zhao-zg/ech-workers/actions/workflows/build-android.yml/badge.svg)
![OpenWrt Build](https://github.com/zhao-zg/ech-workers/actions/workflows/build-openwrt.yml/badge.svg)
```

## 故障排除

### Android 构建失败

1. **NDK 版本问题**: 确保 NDK 版本为 26.3.11579264
2. **Go 版本问题**: 确保使用 Go 1.24+
3. **gomobile 问题**: 检查 gomobile bind 命令输出

### OpenWrt 构建失败

1. **SDK 下载失败**: 检查 SDK URL 是否正确
2. **依赖问题**: 确保所有依赖项都已安装
3. **编译错误**: 查看 V=s 详细输出

### Release 创建失败

1. **权限问题**: 确保 GITHUB_TOKEN 有足够权限
2. **标签格式**: 确保标签以 `v` 开头（如 `v1.0.0`）
3. **重复发布**: 删除旧的同名 Release 或标签

## 本地测试

在推送到 GitHub 之前，可以本地测试构建：

### Android

```bash
cd android
./gradlew assembleRelease
```

### OpenWrt

```bash
cd openwrt
./build.sh
```

## 自定义构建

如需修改构建配置，编辑对应的工作流文件：
- `.github/workflows/build-android.yml`
- `.github/workflows/build-openwrt.yml`
- `.github/workflows/release.yml`
