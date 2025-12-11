# ECH-Workers OpenWrt 软件包

## 概述

ECH-Workers 是一个支持 Encrypted Client Hello (ECH) 的 SOCKS5/HTTP 代理客户端，专为 OpenWrt 路由器优化，具有智能分流功能。

### 主要功能

- ✅ **ECH 支持**: 基于 TLS 1.3 的加密客户端 Hello
- ✅ **多协议**: 支持 SOCKS5 和 HTTP 代理
- ✅ **智能分流**: 三种模式（全局代理/跳过中国大陆/直连）
- ✅ **双栈支持**: IPv4 和 IPv6
- ✅ **DoH 支持**: DNS over HTTPS
- ✅ **自动管理**: 中国 IP 列表自动下载和更新
- ✅ **Web 界面**: 完整的 LuCI 配置界面

## 软件包列表

1. **ech-workers**: 核心程序包
2. **luci-app-ech-workers**: LuCI Web 配置界面

## 安装指南

### 方法 1: 从源码编译（推荐）

#### 准备工作

1. 安装 OpenWrt SDK：
```bash
# 下载对应架构的 SDK
wget https://downloads.openwrt.org/releases/23.05.0/targets/[arch]/[target]/openwrt-sdk-*.tar.xz
tar -xJf openwrt-sdk-*.tar.xz
cd openwrt-sdk-*
```

2. 复制软件包：
```bash
# 复制 ech-workers 包
cp -r /path/to/ech-workers/openwrt package/ech-workers

# 复制 LuCI 应用
cp -r /path/to/ech-workers/openwrt/luci-app-ech-workers package/luci-app-ech-workers
```

#### 编译步骤

```bash
# 更新 feeds
./scripts/feeds update -a
./scripts/feeds install -a

# 配置
make menuconfig
# 在菜单中选择:
# Network -> ech-workers
# LuCI -> Applications -> luci-app-ech-workers

# 编译
make package/ech-workers/compile V=s
make package/luci-app-ech-workers/compile V=s

# 查找生成的 ipk 文件
find bin/packages -name "*ech-workers*.ipk"
```

### 方法 2: 快速构建脚本

```bash
cd /path/to/ech-workers
chmod +x openwrt/build.sh
./openwrt/build.sh
```

### 方法 3: 安装预编译包

```bash
# 上传 ipk 文件到路由器
scp ech-workers_*.ipk root@192.168.1.1:/tmp/
scp luci-app-ech-workers_*.ipk root@192.168.1.1:/tmp/

# SSH 登录路由器
ssh root@192.168.1.1

# 安装
opkg update
opkg install /tmp/ech-workers_*.ipk
opkg install /tmp/luci-app-ech-workers_*.ipk
```

## 配置说明

### 通过 LuCI Web 界面配置

1. 登录 OpenWrt 管理界面
2. 导航到 `服务` -> `ECH-Workers`
3. 配置以下选项：

#### 基本设置

- **启用**: 开启/关闭服务
- **监听地址**: 本地代理监听地址（默认 `0.0.0.0:1080`）
- **服务器地址**: Cloudflare Workers 地址（格式：`域名:端口/路径`）
- **优选 IP**: 可选，指定服务器 IP
- **身份令牌**: 可选，服务端验证令牌

#### ECH 设置

- **DNS 服务器**: DoH 服务器（默认 `dns.alidns.com/dns-query`）
- **ECH 域名**: ECH 公钥查询域名（默认 `cloudflare-ech.com`）

#### 分流设置

- **全局代理**: 所有流量走代理
- **跳过中国大陆**: 中国 IP 直连，其他走代理
- **直连**: 所有流量直连

### 通过命令行配置

编辑配置文件：
```bash
vi /etc/config/ech-workers
```

配置示例：
```
config ech-workers 'main'
    option enabled '1'
    option listen_addr '0.0.0.0:1080'
    option server_addr 'your-worker.workers.dev:443'
    option server_ip ''
    option token 'your-token'
    option dns_server 'dns.alidns.com/dns-query'
    option ech_domain 'cloudflare-ech.com'
    option routing_mode 'bypass_cn'
```

重启服务：
```bash
/etc/init.d/ech-workers restart
```

## 使用场景

### 场景 1: 路由器全局代理

```bash
# 配置为全局代理模式
uci set ech-workers.main.routing_mode='global'
uci commit ech-workers
/etc/init.d/ech-workers restart

# 配置 iptables 转发（可选）
iptables -t nat -A PREROUTING -p tcp -j REDIRECT --to-ports 1080
```

### 场景 2: 智能分流（国内外分流）

```bash
# 配置为跳过中国大陆模式
uci set ech-workers.main.routing_mode='bypass_cn'
uci commit ech-workers
/etc/init.d/ech-workers restart

# IP 列表会自动下载
# 或手动触发：
wget -O /etc/ech-workers/chn_ip.txt \
  https://raw.githubusercontent.com/mayaxcn/china-ip-list/refs/heads/master/chn_ip.txt
```

### 场景 3: 仅作为 SOCKS5 服务器

```bash
# 配置为直连模式
uci set ech-workers.main.routing_mode='none'
uci set ech-workers.main.listen_addr='0.0.0.0:1080'
uci commit ech-workers
/etc/init.d/ech-workers restart
```

## 管理命令

### 服务控制

```bash
# 启动服务
/etc/init.d/ech-workers start

# 停止服务
/etc/init.d/ech-workers stop

# 重启服务
/etc/init.d/ech-workers restart

# 查看状态
/etc/init.d/ech-workers status

# 开机自启
/etc/init.d/ech-workers enable

# 禁用自启
/etc/init.d/ech-workers disable
```

### 查看日志

```bash
# 实时查看日志
logread -f | grep ech-workers

# 查看历史日志
logread | grep ech-workers
```

### IP 列表管理

```bash
# 手动下载 IPv4 列表
wget -O /etc/ech-workers/chn_ip.txt \
  https://raw.githubusercontent.com/mayaxcn/china-ip-list/refs/heads/master/chn_ip.txt

# 手动下载 IPv6 列表
wget -O /etc/ech-workers/chn_ip_v6.txt \
  https://raw.githubusercontent.com/mayaxcn/china-ip-list/refs/heads/master/chn_ip_v6.txt

# 查看列表状态
ls -lh /etc/ech-workers/
```

## 性能优化

### 内存优化

```bash
# 对于内存较小的路由器（< 64MB），建议：
# 1. 仅使用 IPv4 列表
rm -f /etc/ech-workers/chn_ip_v6.txt

# 2. 使用全局代理模式（不加载 IP 列表）
uci set ech-workers.main.routing_mode='global'
```

### 连接优化

```bash
# 调整系统连接跟踪
sysctl -w net.netfilter.nf_conntrack_max=65536
sysctl -w net.netfilter.nf_conntrack_tcp_timeout_established=7200

# 持久化
cat >> /etc/sysctl.conf << EOF
net.netfilter.nf_conntrack_max=65536
net.netfilter.nf_conntrack_tcp_timeout_established=7200
EOF
```

## 故障排除

### 问题 1: 服务无法启动

**检查步骤**:
```bash
# 1. 检查配置
cat /etc/config/ech-workers

# 2. 检查日志
logread | grep ech-workers

# 3. 手动运行测试
/usr/bin/ech-workers -l 0.0.0.0:1080 -f your-server:443

# 4. 检查端口占用
netstat -lntp | grep 1080
```

### 问题 2: IP 列表下载失败

**解决方案**:
```bash
# 1. 检查网络连接
ping -c 4 raw.githubusercontent.com

# 2. 手动下载
wget https://raw.githubusercontent.com/mayaxcn/china-ip-list/refs/heads/master/chn_ip.txt

# 3. 使用代理下载（如果已配置代理）
export http_proxy=http://127.0.0.1:1080
wget -O /etc/ech-workers/chn_ip.txt \
  https://raw.githubusercontent.com/mayaxcn/china-ip-list/refs/heads/master/chn_ip.txt

# 4. 从其他源下载后上传
scp chn_ip.txt root@192.168.1.1:/etc/ech-workers/
```

### 问题 3: 分流不生效

**检查步骤**:
```bash
# 1. 确认分流模式
uci get ech-workers.main.routing_mode

# 2. 确认 IP 列表已加载
ls -lh /etc/ech-workers/chn_ip.txt

# 3. 查看服务日志
logread | grep -E "分流|bypass|routing"

# 4. 重启服务
/etc/init.d/ech-workers restart
```

### 问题 4: LuCI 界面无法访问

**解决方案**:
```bash
# 1. 重启 uhttpd
/etc/init.d/uhttpd restart

# 2. 清除浏览器缓存

# 3. 检查 LuCI 权限
ls -l /usr/share/rpcd/acl.d/luci-app-ech-workers.json

# 4. 重新安装 LuCI 应用
opkg remove luci-app-ech-workers
opkg install /tmp/luci-app-ech-workers_*.ipk
```

## 高级功能

### 透明代理配置

创建透明代理规则：
```bash
#!/bin/sh
# /etc/init.d/ech-workers-transparent

# 创建新链
iptables -t nat -N ECH_WORKERS

# 绕过局域网
iptables -t nat -A ECH_WORKERS -d 0.0.0.0/8 -j RETURN
iptables -t nat -A ECH_WORKERS -d 10.0.0.0/8 -j RETURN
iptables -t nat -A ECH_WORKERS -d 127.0.0.0/8 -j RETURN
iptables -t nat -A ECH_WORKERS -d 169.254.0.0/16 -j RETURN
iptables -t nat -A ECH_WORKERS -d 172.16.0.0/12 -j RETURN
iptables -t nat -A ECH_WORKERS -d 192.168.0.0/16 -j RETURN
iptables -t nat -A ECH_WORKERS -d 224.0.0.0/4 -j RETURN
iptables -t nat -A ECH_WORKERS -d 240.0.0.0/4 -j RETURN

# 重定向到代理
iptables -t nat -A ECH_WORKERS -p tcp -j REDIRECT --to-ports 1080

# 应用规则
iptables -t nat -A PREROUTING -p tcp -j ECH_WORKERS
iptables -t nat -A OUTPUT -p tcp -j ECH_WORKERS
```

### 多节点配置

可以运行多个实例：
```bash
# 创建第二个配置
cp /etc/config/ech-workers /etc/config/ech-workers2

# 修改端口和服务器
uci -c /etc/config set ech-workers2.main.listen_addr='0.0.0.0:1081'
uci -c /etc/config set ech-workers2.main.server_addr='backup-server:443'

# 手动运行第二个实例
/usr/bin/ech-workers -l 0.0.0.0:1081 -f backup-server:443 &
```

## 系统要求

### 最低要求

- **CPU**: 任意架构（ARM/MIPS/x86）
- **内存**: 32MB RAM（全局模式）/ 64MB RAM（分流模式）
- **存储**: 10MB 可用空间
- **OpenWrt**: 19.07 或更高版本

### 推荐配置

- **内存**: 128MB+ RAM
- **存储**: 20MB+ 可用空间
- **OpenWrt**: 21.02 或更高版本
- **架构**: ARM Cortex-A7 或更高

### 兼容性

支持的架构：
- ✅ ARM (armv7, armv8)
- ✅ MIPS (mips, mipsel)
- ✅ x86 (i386, amd64)
- ✅ RISC-V

测试过的设备：
- ✅ Xiaomi R4A Gigabit
- ✅ Newifi D2
- ✅ Raspberry Pi 3/4
- ✅ x86_64 软路由

## 更新日志

### v1.1.0 (2025-12-11)

新增功能：
- ✅ OpenWrt 软件包支持
- ✅ LuCI Web 配置界面
- ✅ 智能分流功能
- ✅ 中国 IP 列表自动管理
- ✅ 多架构支持

### v1.0.0

初始版本：
- ECH 支持
- SOCKS5/HTTP 代理
- 基础功能

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

GPL-3.0 License

## 相关链接

- [项目主页](https://github.com/your-repo/ech-workers)
- [OpenWrt 官方文档](https://openwrt.org/docs/start)
- [LuCI 开发文档](https://github.com/openwrt/luci/wiki)
- [中国 IP 列表](https://github.com/mayaxcn/china-ip-list)

## 支持

如有问题，请：
1. 查看本文档的故障排除部分
2. 查看项目 Issues
3. 提交新的 Issue

---

文档版本: 1.1.0  
最后更新: 2025年12月11日
