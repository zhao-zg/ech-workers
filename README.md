# ECH-Workers

![Android Build](https://github.com/YOUR_USERNAME/ech-workers/actions/workflows/build-android.yml/badge.svg)
![OpenWrt Build](https://github.com/YOUR_USERNAME/ech-workers/actions/workflows/build-openwrt.yml/badge.svg)
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

## 平台支持

### Android 客户端

- VPN 隧道模式
- 分应用代理
- 多配置文件管理
- 自动更新 IP 列表
- 最低支持 Android 7.0 (API 24)

[查看 Android 客户端详细说明 →](android/README.md)

### OpenWrt 客户端

- 透明代理
- LuCI Web 配置界面
- UCI 配置系统
- 系统集成（init.d、hotplug）
- 支持所有 OpenWrt 架构

[查看 OpenWrt 客户端详细说明 →](openwrt/README.md)

## 快速开始

### 1. 部署服务端

将 `server/_worker.js` 部署到 Cloudflare Workers：

```bash
# 登录 Cloudflare Dashboard
# 创建一个新的 Worker
# 复制 _worker.js 的内容到 Worker 编辑器
# 保存并部署
```

### 2. 安装客户端

**方式一：从 Release 下载（推荐）**

前往 [Releases](https://github.com/YOUR_USERNAME/ech-workers/releases) 页面下载最新版本：

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
| 优选IP（域名） | 可选，指定IP或域名绕过DNS | `1.2.3.4` 或 `cf.example.com` |
| 身份令牌 | 可选，服务端验证 | `your-secret-token` |
| ECH DoH 服务器 | 查询 ECH 公钥的 DoH 服务器 | `dns.alidns.com/dns-query` |
| ECH 查询域名 | ECH 公钥所在域名 | `cloudflare-ech.com` |

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

项目使用 GitHub Actions 自动构建：

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

## 技术栈

- **前端**: Cloudflare Workers (JavaScript)
- **Android**: Java + Go (gomobile)
- **OpenWrt**: Go + Lua (LuCI)
- **核心**: Go 1.23+ (ECH 支持)
- **协议**: TLS 1.3, WebSocket, SOCKS5, HTTP

## 性能指标

- **内存占用**: 10-30 MB (含 IP 列表)
- **IP 查询**: < 1ms (二分查找)
- **连接延迟**: 取决于网络
- **并发连接**: 1000+ (取决于设备)

## 常见问题

### Q: ECH 连接失败？
A: 确保使用 Go 1.23+ 编译，并且服务端支持 ECH。

### Q: 跳过中国大陆模式不生效？
A: 检查 IP 列表是否已下载，查看日志确认加载状态。

### Q: 如何更新 IP 列表？
A: Android 会自动每 7 天更新；OpenWrt 可在 LuCI 界面手动更新。

### Q: 支持哪些架构？
A: Android (arm, arm64, x86, x86_64); OpenWrt (所有架构)

更多问题请查看 [Wiki](https://github.com/your-repo/ech-workers/wiki)

## 贡献

欢迎提交 Issue 和 Pull Request！

### 开发指南

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 许可证

本项目采用 GPL-3.0 许可证 - 详见 [LICENSE](LICENSE) 文件

## 致谢

- [hev-socks5-tunnel](https://github.com/heiher/hev-socks5-tunnel) - Android VPN 隧道实现
- [mayaxcn/china-ip-list](https://github.com/mayaxcn/china-ip-list) - 中国 IP 列表数据源
- [Cloudflare](https://www.cloudflare.com/) - Workers 平台和 ECH 支持
- [OpenWrt](https://openwrt.org/) - 路由器系统

## 免责声明

本项目仅供学习和研究使用。请遵守当地法律法规，不得用于非法用途。

## 联系方式

- 项目主页: https://github.com/your-repo/ech-workers
- 问题反馈: https://github.com/your-repo/ech-workers/issues
- 邮箱: your-email@example.com

---

**Star ⭐ 本项目以获取更新通知！**
