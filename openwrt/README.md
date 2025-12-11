# ECH-Workers OpenWrt 软件包

## 概述

ECH-Workers 是一个支持 Encrypted Client Hello (ECH) 的**透明代理**客户端，专为 OpenWrt 路由器优化，具有智能分流功能。

## 透明代理工作原理

```
客户端设备 → 路由器iptables → redsocks → ech-workers(SOCKS5) → Cloudflare Workers → 目标网站
```

### 流程说明

1. **iptables拦截**：通过 iptables NAT 规则自动拦截局域网内所有TCP流量
2. **redsocks转换**：将透明代理流量转换为SOCKS5协议请求
3. **ech-workers处理**：通过ECH加密连接到Cloudflare Workers
4. **智能分流**：根据配置模式决定哪些流量走代理

### 主要功能

- ✅ **透明代理**: 客户端设备无需配置代理，流量自动转发
- ✅ **ECH 支持**: 基于 TLS 1.3 的加密客户端 Hello
- ✅ **智能分流**: 三种模式（全局代理/跳过中国大陆/直连）
- ✅ **双栈支持**: IPv4 和 IPv6
- ✅ **DoH 支持**: DNS over HTTPS
- ✅ **自动管理**: 中国 IP 列表自动下载和更新
- ✅ **Web 界面**: 完整的 LuCI 配置界面

## 软件包列表

1. **ech-workers**: 核心程序包
2. **luci-app-ech-workers**: LuCI Web 配置界面

## 系统要求

### 依赖包

- `redsocks` - 透明代理转SOCKS5
- `ipset` - IP地址集合管理（用于bypass_cn模式）
- `iptables-mod-nat-extra` - iptables NAT扩展模块

这些依赖会在安装时自动安装。

## 安装指南

### 方法 1: 使用预编译包

从 [Releases](https://github.com/zhao-zg/ech-workers/releases) 下载对应架构的 IPK 文件：

```bash
# 上传到路由器后安装
opkg update
opkg install ech-workers_*.ipk
opkg install luci-app-ech-workers_*.ipk

# 重启LuCI
/etc/init.d/uhttpd restart
```

### 方法 2: 从源码编译

#### 准备工作

1. 安装 OpenWrt SDK

```bash
# 下载对应架构的 SDK
wget https://downloads.openwrt.org/releases/23.05.2/targets/x86/64/openwrt-sdk-23.05.2-x86-64_gcc-12.3.0_musl.Linux-x86_64.tar.xz

# 解压
tar xJf openwrt-sdk-*.tar.xz
cd openwrt-sdk-*/
```

2. 添加本项目源码

```bash
# 克隆整个项目
git clone https://github.com/zhao-zg/ech-workers.git package/ech-workers

# 或者只复制 openwrt 目录
mkdir -p package/ech-workers
cp -r /path/to/ech-workers/openwrt/* package/ech-workers/
```

#### 编译步骤

```bash
# 更新feeds
./scripts/feeds update -a
./scripts/feeds install -a

# 配置
make menuconfig
# 在 Network -> Web Servers/Proxies 中选择：
# [M] ech-workers
# [M] luci-app-ech-workers

# 编译
make package/ech-workers/compile V=s
make package/luci-app-ech-workers/compile V=s

# 编译好的包在 bin/packages/*/base/ 目录下
```

## 配置说明

### Web界面配置

访问 `http://路由器IP/cgi-bin/luci/admin/services/ech-workers`

#### 服务器配置

- **启用服务**: 开启/关闭透明代理
- **服务器地址**: Cloudflare Workers 地址（格式：`your-worker.workers.dev:443`）
- **监听地址**: 本地SOCKS5监听地址（默认：`0.0.0.0:20001`）
- **优选IP**: （可选）指定服务器IP或域名以绕过DNS解析
- **反代Host**: （可选）指定Workers连接目标服务器的反代Host
- **身份令牌**: （可选）用于服务器身份验证的令牌

#### 路由配置

- **全局代理**: 所有流量走代理（除局域网和保留地址）
- **跳过中国大陆**: 只代理国外流量，国内直连（需要下载IP列表）
- **直连模式**: 不进行透明代理（仅启动SOCKS5监听）

#### 高级设置

- **DNS服务器**: DoH服务器地址（默认：`dns.alidns.com/dns-query`）
- **ECH域名**: ECH查询域名（默认：`cloudflare-ech.com`）

### 命令行配置

编辑 `/etc/config/ech-workers`:

```bash
config ech-workers 'main'
    option enabled '1'
    option listen_addr '0.0.0.0:20001'
    option server_addr 'your-worker.workers.dev:443'
    option server_ip 'mfa.gov.ua'
    option fallback_hosts ''
    option token ''
    option dns_server 'dns.alidns.com/dns-query'
    option ech_domain 'cloudflare-ech.com'
    option routing_mode 'bypass_cn'
```

重启服务：

```bash
/etc/init.d/ech-workers restart
```

### 查看状态

```bash
# 查看服务状态
/etc/init.d/ech-workers status

# 查看日志
logread | grep ech-workers

# 查看iptables规则
iptables -t nat -L ECH_WORKERS -n -v

# 查看ipset（bypass_cn模式）
ipset list chn_ip | head -20

# 查看redsocks进程
ps | grep redsocks
```

## 分流模式详解

### 1. 全局代理 (global)

**流量路径**：
```
所有TCP → iptables → redsocks → ech-workers → Cloudflare Workers → 目标
```

**适用场景**：
- 需要所有流量加密
- 访问被封锁的网站
- 隐私保护

**排除规则**：
- 局域网地址（192.168.0.0/16, 10.0.0.0/8, 172.16.0.0/12）
- 本地回环（127.0.0.0/8）
- 保留地址

### 2. 跳过中国大陆 (bypass_cn)

**流量路径**：
```
国外TCP → iptables → redsocks → ech-workers → Workers → 目标
国内TCP → 直连
```

**工作原理**：
1. 首次启用时自动下载中国IP列表（约4000条CIDR）
2. 加载IP列表到ipset集合
3. iptables匹配ipset，国内IP直连，其他走代理

**适用场景**：
- 国内网站需要高速访问
- 节省代理流量
- 避免国内服务检测到代理

**IP列表来源**：
- https://github.com/mayaxcn/china-ip-list

### 3. 直连模式 (none)

**流量路径**：
```
所有TCP → 直连（不经过代理）
```

**说明**：
- 不设置iptables规则
- ech-workers仍然监听SOCKS5端口
- 可手动配置客户端使用SOCKS5代理

**适用场景**：
- 仅在特定设备上使用代理
- 需要手动控制代理

## 透明代理技术细节

### iptables规则结构

```bash
# 创建自定义链
iptables -t nat -N ECH_WORKERS

# 跳过局域网和保留地址
iptables -t nat -A ECH_WORKERS -d 192.168.0.0/16 -j RETURN
iptables -t nat -A ECH_WORKERS -d 10.0.0.0/8 -j RETURN
# ... 其他保留地址 ...

# bypass_cn模式：跳过中国IP
iptables -t nat -A ECH_WORKERS -m set --match-set chn_ip dst -j RETURN

# 重定向到redsocks端口（12345）
iptables -t nat -A ECH_WORKERS -p tcp -j REDIRECT --to-ports 12345

# 应用到PREROUTING（转发流量）和OUTPUT（本机流量）
iptables -t nat -A PREROUTING -p tcp -j ECH_WORKERS
iptables -t nat -A OUTPUT -p tcp -j ECH_WORKERS
```

### redsocks配置

```
base {
    log_debug = off;
    log_info = on;
    log = "syslog:daemon";
    daemon = on;
    redirector = iptables;
}

redsocks {
    local_ip = 0.0.0.0;
    local_port = 12345;      # 接收iptables重定向的流量
    ip = 127.0.0.1;
    port = 20001;            # ech-workers SOCKS5端口
    type = socks5;
}
```

### 数据流转换

1. **原始TCP连接**：客户端 → 目标服务器:443
2. **iptables拦截**：目标端口改为 → 路由器:12345
3. **redsocks接收**：解析原始目标地址
4. **SOCKS5封装**：连接127.0.0.1:20001，发送SOCKS5请求（原始目标）
5. **ech-workers处理**：通过ECH连接到Cloudflare Workers
6. **Workers转发**：连接到原始目标服务器:443

## 故障排查

### 代理不生效

**现象**：启用后流量仍直连

**检查步骤**：

1. 确认服务运行
```bash
ps | grep ech-workers
ps | grep redsocks
```

2. 检查iptables规则
```bash
iptables -t nat -L ECH_WORKERS -n -v
# 应该看到 REDIRECT tcp -- 0.0.0.0/0 0.0.0.0/0 redir ports 12345
```

3. 检查redsocks配置
```bash
cat /var/etc/ech-workers-redsocks.conf
```

4. 查看日志
```bash
logread | grep -E "ech-workers|redsocks"
```

**常见原因**：
- redsocks未安装：`opkg install redsocks`
- iptables规则未生效：重启服务
- 防火墙冲突：检查其他代理软件

### bypass_cn模式不生效

**现象**：国内网站也走代理

**检查步骤**：

1. 确认IP列表已下载
```bash
ls -lh /etc/ech-workers/chn_ip.txt
# 应该有约200KB大小
```

2. 检查ipset
```bash
ipset list chn_ip | wc -l
# 应该有约4000条规则
```

3. 手动下载IP列表
```bash
wget -O /etc/ech-workers/chn_ip.txt https://raw.githubusercontent.com/mayaxcn/china-ip-list/refs/heads/master/chn_ip.txt
/etc/init.d/ech-workers restart
```

### 连接超时

**现象**：部分网站无法访问

**可能原因**：
- Workers地址错误
- 优选IP不可达
- ECH配置问题

**解决方法**：
1. 在Web界面测试连接
2. 检查Workers是否正常运行
3. 尝试更换优选IP

### 性能问题

**现象**：网速慢、延迟高

**优化建议**：

1. 使用bypass_cn模式（国内直连）
2. 更换更快的优选IP
3. 升级路由器硬件（CPU和内存）
4. 检查带宽限制

## 卸载

```bash
# 停止服务
/etc/init.d/ech-workers stop
/etc/init.d/ech-workers disable

# 卸载软件包
opkg remove luci-app-ech-workers
opkg remove ech-workers

# 清理配置文件（可选）
rm -rf /etc/config/ech-workers
rm -rf /etc/ech-workers/

# 重启LuCI
/etc/init.d/uhttpd restart
```

## 开发指南

### 构建环境

需要完整的 OpenWrt 构建系统或 SDK。

### 目录结构

```
openwrt/
├── Makefile                    # OpenWrt包定义
├── files/
│   ├── ech-workers.init        # init.d启动脚本（含透明代理逻辑）
│   ├── ech-workers.config      # UCI配置文件
│   └── ech-workers.hotplug     # 网络热插拔脚本
└── luci-app-ech-workers/       # LuCI应用
    ├── Makefile
    ├── luasrc/
    │   ├── controller/         # 控制器
    │   ├── model/cbi/          # 配置模型
    │   └── view/               # 视图模板
    └── po/                     # 翻译文件
        ├── zh_Hans/
        └── zh-cn/
```

### 修改init脚本

编辑 `files/ech-workers.init` 后重新编译：

```bash
make package/ech-workers/clean
make package/ech-workers/compile V=s
```

### 修改LuCI界面

编辑 `luci-app-ech-workers/` 下的文件后：

```bash
make package/luci-app-ech-workers/clean
make package/luci-app-ech-workers/compile V=s
```

## 许可证

MIT License - 详见 [LICENSE](../LICENSE)

## 贡献

欢迎提交 Issue 和 Pull Request！

## 相关链接

- 项目主页: https://github.com/zhao-zg/ech-workers
- Cloudflare Workers文档: https://developers.cloudflare.com/workers/
- ECH标准: https://datatracker.ietf.org/doc/draft-ietf-tls-esni/
- OpenWrt开发文档: https://openwrt.org/docs/guide-developer/start
