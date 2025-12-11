# ECH-Workers Android 客户端

ECH-Workers 的 Android VPN 客户端，支持智能分流和多配置管理。

## 功能特性

- ✅ VPN 隧道模式（基于 hev-socks5-tunnel）
- ✅ SOCKS5/HTTP 代理支持
- ✅ ECH (Encrypted Client Hello) 支持
- ✅ 智能分流（全局/跳过中国大陆/直连）
- ✅ 分应用代理
- ✅ 多配置文件管理
- ✅ 自动更新中国 IP 列表
- ✅ IPv4/IPv6 双栈支持

## 系统要求

- Android 7.0 (API 24) 或更高版本
- 支持架构：armeabi-v7a, arm64-v8a, x86, x86_64

## 安装

### 方式一：下载 APK
```bash
# 下载预编译的 APK
wget https://github.com/your-repo/ech-workers/releases/latest/download/ech-workers.apk

# 安装
adb install ech-workers.apk
```

### 方式二：从源码编译
```bash
# 克隆项目
git clone https://github.com/your-repo/ech-workers.git
cd ech-workers/android

# 编译
./gradlew assembleRelease

# 输出位置
# build/outputs/apk/release/ech-workers-release.apk
```

## 配置说明

### 基本设置

1. **服务器地址**: Cloudflare Workers 地址
   - 格式：`your-worker.workers.dev:443`
   - 支持自定义路径：`your-worker.workers.dev:443/path`

2. **本地代理端口**: 默认 1080
   - 范围：1024-65535
   - 用于 VPN 隧道内部通信

3. **优选IP（域名）**: 可选
   - 可以是 IP 地址：`1.2.3.4`
   - 也可以是域名：`cf.example.com`
   - 用于绕过 DNS 解析或使用备用地址

4. **身份令牌**: 可选
   - 用于服务端验证
   - 与服务端配置一致

### ECH 配置

1. **ECH DoH 服务器**: 
   - 默认：`dns.alidns.com/dns-query`
   - 用于查询 ECH 公钥

2. **ECH 查询域名**:
   - 默认：`cloudflare-ech.com`
   - ECH 公钥所在的域名

### 分流模式

| 模式 | 说明 | 使用场景 |
|------|------|----------|
| 全局代理 | 所有流量通过代理 | 完全隐私保护 |
| 跳过中国大陆 | 国内直连，国外代理 | 日常使用，节省流量 |
| 直连模式 | 所有流量直连 | 测试或禁用代理 |

### 代理模式

- **全局代理**: 所有应用的流量都走代理
- **分应用代理**: 仅选中的应用走代理

## 使用指南

### 首次使用

1. 打开应用并授予 VPN 权限
2. 配置服务器地址
3. 选择分流模式
4. 点击"启动"按钮

### 配置管理

#### 添加配置
1. 点击"添加"按钮
2. 输入配置名称
3. 配置参数
4. 点击"保存"

#### 切换配置
1. 在配置下拉列表中选择
2. 自动加载对应配置

#### 删除配置
1. 选择要删除的配置
2. 点击"删除"按钮
3. 确认删除（至少保留一个配置）

### 分应用代理

1. 取消勾选"全局代理"
2. 点击"分应用代理"按钮
3. 选择需要代理的应用
4. 保存并启动

## 中国 IP 列表

### 自动管理

应用会自动：
- 首次使用时下载 IP 列表
- 缓存到本地存储
- 每 7 天自动检查更新
- 失败时自动重试

### 手动更新

暂不支持手动触发更新，应用会在后台自动处理。

### 数据来源

- IPv4: https://github.com/mayaxcn/china-ip-list
- IPv6: https://github.com/mayaxcn/china-ip-list

## 故障排除

### 连接失败

**症状**: 显示"代理失败"或无法连接

**解决方案**:
1. 检查服务器地址是否正确
2. 确认网络连接正常
3. 尝试配置优选IP（域名）
4. 检查身份令牌是否匹配
5. 查看系统日志

### VPN 权限被拒绝

**症状**: 启动时提示权限错误

**解决方案**:
1. 卸载其他 VPN 应用
2. 重启应用并重新授权
3. 检查系统设置中的 VPN 权限

### IP 列表下载失败

**症状**: 日志显示"下载 IP 列表失败"

**解决方案**:
1. 检查网络连接
2. 确认可以访问 GitHub
3. 临时切换到"全局代理"模式
4. 等待自动重试

### 分流不生效

**症状**: 国内网站也走代理

**解决方案**:
1. 确认选择了"跳过中国大陆"模式
2. 检查 IP 列表是否已加载
3. 重启应用
4. 查看日志确认分流逻辑

## 查看日志

使用 adb 查看日志：

```bash
# 查看所有日志
adb logcat | grep -E "TProxyService|Tunnel|ChinaIpListManager"

# 仅查看错误
adb logcat *:E | grep -E "TProxyService|Tunnel"

# 保存日志到文件
adb logcat -d > ech-workers.log
```

## 开发构建

### 环境要求

- Android SDK 34
- NDK 26.3.11579264
- Go 1.24+ (支持 ECH)
- Gradle 8.0+
- JDK 17

### 构建步骤

```bash
# 1. 克隆项目
git clone https://github.com/your-repo/ech-workers.git
cd ech-workers/android

# 2. 编译 Go 库（可选，已预编译）
cd ../tunnel
gomobile bind -target=android -o ../android/libs/tunnel.aar .

# 3. 编译 Android 应用
cd ../android
./gradlew assembleDebug      # 调试版本
./gradlew assembleRelease    # 发布版本

# 4. 输出位置
# build/outputs/apk/debug/ech-workers-debug.apk
# build/outputs/apk/release/ech-workers-release.apk
```

### 签名配置

创建 `store.properties`：

```properties
storeFile=your-keystore.jks
storePassword=your-store-password
keyAlias=your-key-alias
keyPassword=your-key-password
```

## 项目结构

```
android/
├── build.gradle              # Gradle 构建配置
├── src/main/
│   ├── AndroidManifest.xml   # 应用清单
│   ├── java/com/ech/workers/ # Java 源代码
│   │   ├── MainActivity.java
│   │   ├── TProxyService.java
│   │   ├── Preferences.java
│   │   ├── ChinaIpListManager.java
│   │   ├── AppListActivity.java
│   │   └── ServiceReceiver.java
│   ├── jni/                  # JNI 本地代码
│   │   ├── Android.mk
│   │   └── hev-socks5-tunnel/
│   └── res/                  # 资源文件
│       ├── layout/main.xml
│       └── values/strings.xml
└── libs/                     # 第三方库
    └── tunnel.aar            # Go 编译的库
```

## 性能数据

- **APK 大小**: ~5-8 MB
- **运行内存**: 15-30 MB
- **IP 列表**: 1-2 MB
- **电池消耗**: 中等（取决于使用量）
- **CPU 占用**: < 5%（空闲时）

## 隐私说明

应用收集的数据：
- ✅ 配置信息（仅本地存储）
- ✅ IP 列表缓存（仅本地存储）
- ❌ 不收集个人信息
- ❌ 不上传浏览数据
- ❌ 不追踪用户行为

## 更新日志

### v1.1.0 (2025-12-11)
- ✅ 添加智能分流功能
- ✅ 支持跳过中国大陆模式
- ✅ 自动管理中国 IP 列表
- ✅ 优化 ECH 配置方式
- ✅ 改进 UI 界面

### v1.0.0
- ✅ 初始版本
- ✅ ECH 支持
- ✅ VPN 隧道
- ✅ 多配置管理

## 许可证

GPL-3.0 License - 详见 [LICENSE](../LICENSE)

## 相关资源

- [项目主页](https://github.com/your-repo/ech-workers)
- [OpenWrt 客户端](../openwrt/README.md)
- [问题反馈](https://github.com/your-repo/ech-workers/issues)

## 致谢

- [hev-socks5-tunnel](https://github.com/heiher/hev-socks5-tunnel)
- [mayaxcn/china-ip-list](https://github.com/mayaxcn/china-ip-list)
