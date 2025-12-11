# ECH-Workers LuCI 应用中文翻译

## 翻译文件位置

`po/zh-cn/ech-workers.po`

## 翻译说明

本文件包含 ECH-Workers LuCI Web 界面的简体中文翻译。

### 主要术语翻译

| 英文 | 中文 |
|------|------|
| ECH-Workers | ECH-Workers（保持不变） |
| Encrypted Client Hello | 加密客户端 Hello |
| SOCKS5/HTTP proxy | SOCKS5/HTTP 代理 |
| Intelligent routing | 智能分流 |
| Server Address | 服务器地址 |
| 优选IP（域名） | 优选IP（域名） |
| Authentication Token | 身份令牌 |
| DNS Server | DNS 服务器 |
| ECH Domain | ECH 域名 |
| Routing Mode | 分流模式 |
| Global Proxy | 全局代理 |
| Bypass China Mainland | 跳过中国大陆 |
| Direct Connection | 直连模式 |
| IP List Management | IP列表管理 |

### 分流模式说明

1. **全局代理（Global Proxy）**
   - 所有流量都通过代理服务器转发
   - 适用于需要完全隐私保护的场景

2. **跳过中国大陆（Bypass China Mainland）**
   - 国内 IP 地址直接连接
   - 国外 IP 地址通过代理转发
   - 需要下载中国 IP 列表
   - 适用于日常使用，可节省流量

3. **直连模式（Direct Connection）**
   - 所有流量都直接连接
   - 不使用代理
   - 适用于测试或临时禁用代理

### 编译翻译文件

翻译文件会在编译 LuCI 应用时自动处理：

```bash
cd openwrt-sdk
make package/luci-app-ech-workers/compile V=s
```

### 测试翻译

1. 安装编译好的 ipk 包
2. 在 LuCI 界面中选择简体中文语言
3. 访问 服务 → ECH-Workers
4. 验证所有文本是否正确显示中文

### 更新翻译

如果需要更新翻译：

1. 编辑 `po/zh-cn/ech-workers.po` 文件
2. 修改对应的 `msgstr` 字段
3. 重新编译 LuCI 应用

### 翻译规范

1. **保持一致性**: 同一术语在整个界面中使用相同翻译
2. **简洁明了**: 翻译应简洁易懂，避免过于冗长
3. **符合习惯**: 使用中文用户习惯的表达方式
4. **技术准确**: 技术术语翻译要准确，必要时可保留英文

### 特殊说明

- **ECH-Workers**: 项目名称，保持英文不翻译
- **优选IP（域名）**: 这是一个特定功能，保持原样
- **DoH**: DNS over HTTPS 的缩写，通常保持不翻译
- **ECH**: Encrypted Client Hello 的缩写，保持不翻译

## 贡献翻译

如果发现翻译错误或需要改进，欢迎提交 Pull Request：

1. Fork 项目
2. 修改 `po/zh-cn/ech-workers.po` 文件
3. 提交 PR 并说明修改原因

## 其他语言

如需添加其他语言翻译：

1. 在 `po/` 目录下创建对应语言代码的子目录
2. 复制 `zh-cn/ech-workers.po` 作为模板
3. 翻译 `msgstr` 字段
4. 提交 PR

支持的语言代码示例：
- `zh-cn`: 简体中文
- `zh-tw`: 繁体中文
- `en`: 英文
- `ja`: 日文
- `ko`: 韩文
- `ru`: 俄文
- `de`: 德文
- `fr`: 法文
- `es`: 西班牙文
