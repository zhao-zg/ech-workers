# 🚀 上传到 GitHub 指南

## ✅ 已完成的工作

1. **清理文档**
   - ✅ 删除 `PROJECT_STRUCTURE.md`
   - ✅ 删除空的 `docs/` 目录
   - ✅ 只保留必要的 README 文件

2. **GitHub Actions 配置**
   - ✅ Android 自动构建 (`.github/workflows/build-android.yml`)
   - ✅ OpenWrt 自动构建 (`.github/workflows/build-openwrt.yml`)
   - ✅ Release 自动发布 (`.github/workflows/release.yml`)

3. **项目文件**
   - ✅ `.gitignore` - Git 忽略配置
   - ✅ `LICENSE` - MIT 许可证
   - ✅ `.github/FUNDING.yml` - 赞助配置
   - ✅ `.github/ACTIONS_GUIDE.md` - 详细的 Actions 使用指南

4. **上传脚本**
   - ✅ `upload.ps1` - Windows PowerShell 脚本
   - ✅ `upload.sh` - Linux/macOS Bash 脚本

## 📋 上传步骤

### 方式一：使用自动脚本（推荐）

**Windows:**
```powershell
.\upload.ps1
```

**Linux/macOS:**
```bash
chmod +x upload.sh
./upload.sh
```

### 方式二：手动上传

#### 1. 初始化 Git 仓库

```bash
git init
git branch -M main
```

#### 2. 在 GitHub 创建仓库

访问 https://github.com/new 创建新仓库（不要初始化 README、.gitignore 或 LICENSE）

#### 3. 添加文件并提交

```bash
git add .
git commit -m "Initial commit"
```

#### 4. 关联远程仓库

```bash
# 替换 YOUR_USERNAME 和 YOUR_REPO
git remote add origin https://github.com/YOUR_USERNAME/YOUR_REPO.git
```

#### 5. 推送到 GitHub

```bash
git push -u origin main
```

## 🏷️ 发布新版本

### 1. 准备发布

确保所有代码已提交：

```bash
git add .
git commit -m "准备发布 v1.0.0"
git push origin main
```

### 2. 创建标签

```bash
# 创建标签（版本号格式：v主版本.次版本.修订号）
git tag -a v1.0.0 -m "Release version 1.0.0"

# 推送标签到 GitHub
git push origin v1.0.0
```

### 3. 自动构建

推送标签后，GitHub Actions 会自动：
- ✅ 创建 GitHub Release
- ✅ 构建 Android APK (所有架构)
- ✅ 构建 OpenWrt IPK (x86_64, aarch64, mipsel)
- ✅ 将所有文件上传到 Release

构建大约需要 15-30 分钟完成。

### 4. 查看构建进度

1. 访问仓库的 Actions 标签页
2. 查看工作流运行状态
3. 构建完成后，在 Releases 页面查看发布的软件包

## 🔧 配置 APK 签名（可选）

如需发布已签名的 APK，在仓库设置中添加以下 Secrets：

### 1. 生成 Keystore

```bash
keytool -genkey -v -keystore release.keystore \
  -alias ech-workers \
  -keyalg RSA -keysize 2048 \
  -validity 10000
```

### 2. 转换为 Base64

**Windows PowerShell:**
```powershell
$bytes = [System.IO.File]::ReadAllBytes("release.keystore")
[Convert]::ToBase64String($bytes) | Out-File -Encoding ASCII keystore.base64.txt
```

**Linux/macOS:**
```bash
base64 release.keystore > keystore.base64.txt
```

### 3. 添加 GitHub Secrets

进入仓库的 `Settings` → `Secrets and variables` → `Actions`，添加：

| Secret 名称 | 说明 |
|------------|------|
| `KEYSTORE_BASE64` | keystore.base64.txt 的内容 |
| `KEYSTORE_PASSWORD` | Keystore 密码 |
| `KEY_ALIAS` | 密钥别名（如 ech-workers） |
| `KEY_PASSWORD` | 密钥密码 |

## 📝 更新 README

上传后，需要更新 README.md 中的仓库地址：

1. 将所有 `YOUR_USERNAME` 替换为你的 GitHub 用户名
2. 将所有 `YOUR_REPO` 替换为你的仓库名

搜索并替换以下内容：
- `YOUR_USERNAME/ech-workers` → `实际用户名/实际仓库名`

## 🎯 构建状态徽章

构建徽章会自动显示在 README 顶部：

- 🟢 绿色：构建成功
- 🔴 红色：构建失败
- 🟡 黄色：构建中

## 📦 下载构建产物

### Artifacts（每次提交）

1. 访问 `Actions` 标签
2. 点击对应的工作流运行
3. 在底部 `Artifacts` 区域下载

保留时间：30 天

### Releases（标签发布）

1. 访问 `Releases` 页面
2. 下载对应版本的文件

永久保留

## 🔍 故障排除

### 构建失败

1. 查看 Actions 日志找出错误原因
2. 修复代码后重新提交
3. 或手动重新运行失败的工作流

### 推送失败

```bash
# 如果远程仓库已有内容，先拉取
git pull origin main --allow-unrelated-histories
git push origin main
```

### 标签冲突

```bash
# 删除本地标签
git tag -d v1.0.0

# 删除远程标签
git push origin :refs/tags/v1.0.0

# 重新创建
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
```

## 📚 更多信息

- [GitHub Actions 配置指南](.github/ACTIONS_GUIDE.md)
- [Android 客户端说明](android/README.md)
- [OpenWrt 客户端说明](openwrt/README.md)

## ✨ 下一步

上传成功后，你可以：

1. ⭐ 为项目添加 Star
2. 📝 完善 README 和 Wiki
3. 🐛 创建 Issue 跟踪 bug
4. 🔧 添加 GitHub Projects 管理任务
5. 🤝 邀请贡献者

祝你使用愉快！🎉
