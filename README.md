# ECH-Workers

![Android Build](https://github.com/zhao-zg/ech-workers/actions/workflows/build-android.yml/badge.svg)
![OpenWrt Build](https://github.com/zhao-zg/ech-workers/actions/workflows/build-openwrt.yml/badge.svg)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

基于 Encrypted Client Hello (ECH) 的智能代理客户端，支持 Android 和 OpenWrt 平台。

## 项目简介

ECH-Workers 是一个支持 ECH（加密客户端 Hello）的 SOCKS5/HTTP 代理客户端，可连接到 Cloudflare Workers 服务端，提供安全、智能的网络代理服务。

### 主要特性

- ✅ **ECH 支持**: 基于 TLS 1.3 的 Encrypted Client Hello
- ✅ **双协议支持**: SOCKS5 和 HTTP 代理
- ✅ **智能分流**: 
  - 全局代理
  - 跳过中国大陆（国内直连，国外代理）
  - 直连模式
- ✅ **双平台支持**: Android 和 OpenWrt
- ✅ **IPv4/IPv6**: 双栈支持
- ✅ **DoH 支持**: DNS over HTTPS
- ✅ **自动管理**: 中国 IP 列表自动下载和更新
- ✅ **版本检查**: 自动检查更新

## 平台支持

### Android 客户端

- VPN 隧道模式
- 分应用代理
- 多配置文件管理
- 自动更新 IP 列表
- 显示出口 IP 和国家
- 最低支持 Android 7.0 (API 24)

[查看 Android 客户端详细说明 →](android/README.md)

### OpenWrt 客户端

- 透明代理
- LuCI Web 配置界面
- UCI 配置系统
- 系统集成（init.d、hotplug）
- 显示出口 IP 和国旗
- 支持所有 OpenWrt 架构

[查看 OpenWrt 客户端详细说明 →](openwrt/README.md)

## 快速开始

### 1. 部署服务端

将 `server/_worker.js` 部署到 Cloudflare Workers。

```bash
# 登录 Cloudflare Dashboard
# 创建一个新的 Worker
# 复制 _worker.js 的内容到 Worker 编辑器
# 保存并部署
```

### 2. 安装客户端

**方式一：从 Release 下载（推荐）**

前往 [Releases](https://github.com/zhao-zg/ech-workers/releases) 页面下载最新版本：

Android:
- `ech-workers-signed.apk` - 已签名版本（推荐）
- `ech-workers-unsigned.apk` - 未签名版本

OpenWrt:
- `ech-workers_*_x86_64.ipk` - x86_64 架构
- `ech-workers_*_aarch64.ipk` - ARM64 架构
- `ech-workers_*_mipsel_24kc.ipk` - MIPS 架构
- `luci-app-ech-workers_*.ipk` - LuCI Web 界面

**方式二：从源码构建**

Android:
```bash
cd android
./gradlew assembleRelease
```

OpenWrt:
```bash
cd openwrt
./build.sh
```

**安装方法:**

Android:
```bash
# 使用 adb 安装
adb install ech-workers-signed.apk

# 或直接在手机上安装 APK
```

OpenWrt:
```bash
# 上传 IPK 文件到路由器
opkg update
opkg install ech-workers_*.ipk
opkg install luci-app-ech-workers_*.ipk
```

### 3. 配置并启动

**Android:**
1. 打开应用
2. 输入服务器地址（如：`your-worker.workers.dev:443`）
3. 选择分流模式
4. 点击启动

**OpenWrt:**
1. 访问 LuCI Web 界面
2. 进入 "服务" → "ECH-Workers"
3. 配置服务器地址和分流模式
4. 启用服务

## 配置说明

### 基本配置

| 配置项 | 说明 | 示例 |
|--------|------|------|
| 服务器地址 | Cloudflare Worker 地址 | `your-worker.workers.dev:443` |
| 优选IP（域名） | 可选，客户端连接Workers的IP | `1.2.3.4` 或 `cf.example.com` 或 `mfa.gov.ua` |
| 反代Host | 可选，Workers连接目标的Host | `1.2.3.4` 或 `example.com` |
| 身份令牌 | 可选，服务端验证 | `your-secret-token` |
| ECH DoH 服务器 | 查询 ECH 公钥的 DoH 服务器 | `dns.alidns.com/dns-query` |
| ECH 查询域名 | ECH 公钥所在域名 | `cloudflare-ech.com` |

**参数说明：**
- **优选IP**：用于客户端直接连接到Cloudflare Workers，绕过DNS解析
- **反代Host**：传递给Workers服务端，用于Workers连接真实目标服务器
  - 如果设置了反代Host，Workers优先使用该Host连接目标
  - 如果反代Host连接失败，自动尝试使用FALLBACK_HOSTS
  - 支持IPv4、IPv6地址或域名

### 分流模式

| 模式 | 说明 | 适用场景 |
|------|------|----------|
| 全局代理 | 所有流量走代理 | 完全隐私保护 |
| 跳过中国大陆 | 国内直连，国外代理 | 国内外分流，节省流量 |
| 直连模式 | 所有流量直连 | 测试或特殊需求 |

## 项目结构

```
ech-workers/
├── .github/         # GitHub Actions 工作流
│   └── workflows/   # 自动构建配置
├── server/          # Cloudflare Workers 服务端
├── android/         # Android 客户端
├── openwrt/         # OpenWrt 客户端和 LuCI 界面
└── tunnel/          # 共享的 Go 代理核心代码
```

## 自动构建

项目使用 GitHub Actions 自动构建。

- **推送代码**: 自动构建并上传 Artifacts
- **创建标签**: 自动创建 Release 并发布软件包

详细说明请查看 [GitHub Actions 配置指南](.github/ACTIONS_GUIDE.md)

### 发布新版本

```bash
# 创建标签
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0

# GitHub Actions 会自动构建并发布
```

## 编译构建

### Android

```bash
cd android
./gradlew assembleRelease
# 输出: android/build/outputs/apk/release/
```

### OpenWrt

```bash
cd openwrt
./build.sh
# 输出: bin/packages/*/
```

详细构建说明请查看各平台的 README。

## 命令行使用

### 查看版本

```bash
# 显示当前版本
./ech-workers -version
```

### 检查更新

```bash
# 手动检查是否有新版本
./ech-workers -check-update
```

**自动更新检查:**
- 程序启动时会自动在后台检查更新（不阻塞启动）
- 发现新版本会自动提示下载地址和更新内容
- 检查失败不影响程序正常运行

### 启动代理

```bash
# 基本使用
./ech-workers -l 0.0.0.0:1080 -f your-worker.workers.dev:443 -token your-token

# 使用优选IP
./ech-workers -l 0.0.0.0:1080 -f your-worker.workers.dev:443 -ip 1.2.3.4

# 使用反代Host
./ech-workers -l 0.0.0.0:1080 -f your-worker.workers.dev:443 -fallback example.com

# 设置分流模式
./ech-workers -l 0.0.0.0:1080 -f your-worker.workers.dev:443 -routing bypass_cn
```

### 命令行参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-version` | 显示版本信息 | - |
| `-check-update` | 检查更新 | - |
| `-l` | 监听地址 | `0.0.0.0:1080` |
| `-f` | 服务器地址 | 必填 |
| `-ip` | 优选IP（域名） | 空 |
| `-fallback` | 反代Host | 空 |
| `-token` | 身份令牌 | 空 |
| `-dns` | DoH 服务器 | `dns.alidns.com/dns-query` |
| `-ech` | ECH 查询域名 | `cloudflare-ech.com` |
| `-routing` | 分流模式 | `bypass_cn` |

## 技术栈

- **服务端**: Cloudflare Workers (JavaScript)
- **Android**: Java + Go (gomobile)
- **OpenWrt**: Go + Lua (LuCI)
- **核心**: Go 1.24+ (ECH 支持)
- **协议**: TLS 1.3, WebSocket, SOCKS5, HTTP

## 性能指标

- **内存占用**: 10-30 MB (含 IP 列表)
- **IP 查询**: < 1ms (二分查找)
- **连接延迟**: 取决于网络
- **并发连接**: 1000+ (取决于设备)

## 常见问题

### Q: 如何检查是否有新版本？
A: 
- **自动检查**: 程序启动时自动在后台检查更新
- **手动检查**: 运行 `./ech-workers -check-update`
- **查看版本**: 运行 `./ech-workers -version`

### Q: ECH 连接失败？
A: 确保使用 Go 1.24+ 编译，并且服务端支持 ECH.

### Q: 跳过中国大陆模式不生效？
A: 检查 IP 列表是否已下载，查看日志确认加载状态。

### Q: 如何更新 IP 列表？
A: Android 会自动每 7 天更新；OpenWrt 可在 LuCI 界面手动更新。

### Q: 支持哪些架构？
A: Android (arm, arm64, x86, x86_64); OpenWrt (所有架构)

## 贡献

欢迎提交 Issue 和 Pull Request。

### 开发指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 开发者指南

### 发布新版本

使用 `release.ps1` 脚本快速发布新版本：

```powershell
# 发布新版本（会自动创建 tag 并触发构建）
.\release.ps1 -Version 1.0.11

# 或者带自定义消息
.\release.ps1 -Version 1.0.11 -Message "修复某某问题"
```

**工作流程：**
1. 脚本会自动创建 git tag（格式：v1.0.11）
2. 推送 tag 到 GitHub
3. GitHub Actions 自动构建所有平台的安装包
4. 版本号会自动注入到所有平台（无需手动修改代码）

**版本号来源：**
- CLI 工具：编译时通过 `-ldflags` 从 git tag 注入
- Android：build.gradle 从 git tag 读取
- OpenWrt：从 git tag 读取

### 手动发布步骤

如果不使用脚本，可以手动操作：

```bash
# 创建并推送 tag
git tag -a v1.0.11 -m "版本说明"
git push origin v1.0.11

# GitHub Actions 会自动构建和发布
```

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 致谢

- [hev-socks5-tunnel](https://github.com/heiher/hev-socks5-tunnel) - Android 隧道实现
- [Cloudflare](https://www.cloudflare.com/) - ECH 支持和 Workers 平台
- [OpenWrt](https://openwrt.org/) - 开源路由器系统

## 联系方式

- 项目主页: https://github.com/zhao-zg/ech-workers
- 问题反馈: https://github.com/zhao-zg/ech-workers/issues

---

**Star ⭐ 本项目以获取更新通知！**
