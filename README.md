# ECH-Workers

![Android Build](https://github.com/zhao-zg/ech-workers/actions/workflows/build-android.yml/badge.svg)
![OpenWrt Build](https://github.com/zhao-zg/ech-workers/actions/workflows/build-openwrt.yml/badge.svg)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

基于 Encrypted Client Hello (ECH) 的智能代理客户端，支�?Android �?OpenWrt 平台�?

## 项目简�?

ECH-Workers 是一个支�?ECH（加密客户端 Hello）的 SOCKS5/HTTP 代理客户端，可连接到 Cloudflare Workers 服务端，提供安全、智能的网络代理服务�?

### 主要特�?

- �?**ECH 支持**: 基于 TLS 1.3 �?Encrypted Client Hello
- �?**双协议支�?*: SOCKS5 �?HTTP 代理
- �?**智能分流**: 
  - 全局代理
  - 跳过中国大陆（国内直连，国外代理�?
  - 直连模式
- �?**双平台支�?*: Android �?OpenWrt
- �?**IPv4/IPv6**: 双栈支持
- �?**DoH 支持**: DNS over HTTPS
- �?**自动管理**: 中国 IP 列表自动下载和更�?

## 平台支持

### Android 客户�?

- VPN 隧道模式
- 分应用代�?
- 多配置文件管�?
- 自动更新 IP 列表
- 最低支�?Android 7.0 (API 24)

[查看 Android 客户端详细说�?→](android/README.md)

### OpenWrt 客户�?

- 透明代理
- LuCI Web 配置界面
- UCI 配置系统
- 系统集成（init.d、hotplug�?
- 支持所�?OpenWrt 架构

[查看 OpenWrt 客户端详细说�?→](openwrt/README.md)

## 快速开�?

### 1. 部署服务�?

�?`server/_worker.js` 部署�?Cloudflare Workers�?

```bash
# 登录 Cloudflare Dashboard
# 创建一个新�?Worker
# 复制 _worker.js 的内容到 Worker 编辑�?
# 保存并部�?
```

### 2. 安装客户�?

**方式一：从 Release 下载（推荐）**

前往 [Releases](https://github.com/zhao-zg/ech-workers/releases) 页面下载最新版本：

Android:
- `ech-workers-signed.apk` - 已签名版本（推荐�?
- `ech-workers-unsigned.apk` - 未签名版�?

OpenWrt:
- `ech-workers_*_x86_64.ipk` - x86_64 架构
- `ech-workers_*_aarch64.ipk` - ARM64 架构
- `ech-workers_*_mipsel_24kc.ipk` - MIPS 架构
- `luci-app-ech-workers_*.ipk` - LuCI Web 界面

**方式二：从源码构�?*

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

# 或直接在手机上安�?APK
```

OpenWrt:
```bash
# 上传 IPK 文件到路由器
opkg update
opkg install ech-workers_*.ipk
opkg install luci-app-ech-workers_*.ipk
```

### 3. 配置并启�?

**Android:**
1. 打开应用
2. 输入服务器地址（如：`your-worker.workers.dev:443`�?
3. 选择分流模式
4. 点击启动

**OpenWrt:**
1. 访问 LuCI Web 界面
2. 进入 "服务" �?"ECH-Workers"
3. 配置服务器地址和分流模�?
4. 启用服务

## 配置说明

### 基本配置

| 配置�?| 说明 | 示例 |
|--------|------|------|
| 服务器地址 | Cloudflare Worker 地址 | `your-worker.workers.dev:443` |
| 优选IP（域名） | 可选，指定IP或域名绕过DNS | `1.2.3.4` �?`cf.example.com` |
| 身份令牌 | 可选，服务端验�?| `your-secret-token` |
| ECH DoH 服务�?| 查询 ECH 公钥�?DoH 服务�?| `dns.alidns.com/dns-query` |
| ECH 查询域名 | ECH 公钥所在域�?| `cloudflare-ech.com` |

### 分流模式

| 模式 | 说明 | 适用场景 |
|------|------|----------|
| 全局代理 | 所有流量走代理 | 完全隐私保护 |
| 跳过中国大陆 | 国内直连，国外代�?| 国内外分流，节省流量 |
| 直连模式 | 所有流量直�?| 测试或特殊需�?|

## 项目结构

```
ech-workers/
├── .github/         # GitHub Actions 工作�?
�?  └── workflows/   # 自动构建配置
├── server/          # Cloudflare Workers 服务�?
├── android/         # Android 客户�?
├── openwrt/         # OpenWrt 客户端和 LuCI 界面
└── tunnel/          # 共享�?Go 代理核心代码
```

## 自动构建

项目使用 GitHub Actions 自动构建�?

- **推送代�?*: 自动构建并上�?Artifacts
- **创建标签**: 自动创建 Release 并发布软件包

详细说明请查�?[GitHub Actions 配置指南](.github/ACTIONS_GUIDE.md)

### 发布新版�?

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

详细构建说明请查看各平台�?README�?

## 技术栈

- **服务�?*: Cloudflare Workers (JavaScript)
- **Android**: Java + Go (gomobile)
- **OpenWrt**: Go + Lua (LuCI)
- **核心**: Go 1.24+ (ECH 支持)
- **协议**: TLS 1.3, WebSocket, SOCKS5, HTTP

## 性能指标

- **内存占用**: 10-30 MB (�?IP 列表)
- **IP 查询**: < 1ms (二分查找)
- **连接延迟**: 取决于网�?
- **并发连接**: 1000+ (取决于设�?

## 常见问题

### Q: ECH 连接失败�?
A: 确保使用 Go 1.24+ 编译，并且服务端支持 ECH.

### Q: 跳过中国大陆模式不生效？
A: 检�?IP 列表是否已下载，查看日志确认加载状态�?

### Q: 如何更新 IP 列表�?
A: Android 会自动每 7 天更新；OpenWrt 可在 LuCI 界面手动更新�?

### Q: 支持哪些架构�?
A: Android (arm, arm64, x86, x86_64); OpenWrt (所有架�?

## 贡献

欢迎提交 Issue �?Pull Request�?

### 开发指�?

1. Fork 项目
2. 创建特性分�?(`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 许可�?

本项目采�?MIT 许可�?- 详见 [LICENSE](LICENSE) 文件

## 致谢

- [hev-socks5-tunnel](https://github.com/heiher/hev-socks5-tunnel) - Android 隧道实现
- [Cloudflare](https://www.cloudflare.com/) - ECH 支持�?Workers 平台
- [OpenWrt](https://openwrt.org/) - 开源路由器系统

## 联系方式

- GitHub: [@zhao-zg](https://github.com/zhao-zg)
- Issues: [提交问题](https://github.com/zhao-zg/ech-workers/issues)

## 联系方式

- 项目主页: https://github.com/zhao-zg/ech-workers
- 问题反馈: https://github.com/zhao-zg/ech-workers/issues
- 邮箱: your-email@example.com

---

**Star �?本项目以获取更新通知�?*
