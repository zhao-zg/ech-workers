-- Copyright (C) 2025 ECH-Workers

local m, s, o
local sys = require "luci.sys"
local uci = require "luci.model.uci".cursor()

-- 使用 Map + TabContainer 实现分页
m = Map("ech-workers", translate("ECH-Workers"))
m.description = translate("ECH-enabled SOCKS5/HTTP proxy with intelligent routing. Supports Encrypted Client Hello (ECH) and automatic China mainland bypass.")

-- ==================== 状态页 ====================
s = m:section(TypedSection, "ech-workers")
s.anonymous = true
s.addremove = false

-- 创建标签页容器
s:tab("status", translate("Status"))
s:tab("basic", translate("Basic Settings"))
s:tab("advanced", translate("Advanced"))
s:tab("iplist", translate("IP List"))
s:tab("logs", translate("Logs"))

-- 状态标签页
o = s:taboption("status", DummyValue, "_status", translate("Service Status"))
o.template = "ech-workers/status"
o.value = translate("Collecting data...")

o = s:taboption("status", DummyValue, "_proxy_test", translate("Connection Test"))
o.template = "ech-workers/proxy_test"

o = s:taboption("status", DummyValue, "_version", translate("Version"))
o.rawhtml = true
function o.cfgvalue(self, section)
	local version = sys.exec("ech-workers -version 2>/dev/null | grep -oP 'v[0-9]+\\.[0-9]+\\.[0-9]+' || echo 'Unknown'")
	return string.format('<span style="color:#090;font-weight:bold">%s</span>', version:gsub("\n", ""))
end

o = s:taboption("status", Button, "_check_update", translate("Check Update"))
o.inputtitle = translate("Check Now")
o.inputstyle = "reload"
o.write = function()
	luci.http.redirect(luci.dispatcher.build_url("admin", "services", "ech-workers", "check_update"))
end

-- 基本设置标签页
o = s:taboption("basic", Flag, "enabled", translate("Enable Service"))
o.default = "0"
o.rmempty = false

o = s:taboption("basic", Value, "server_addr", translate("Server Address"))
o.description = translate("Cloudflare Workers address (e.g., your-worker.workers.dev:443)")
o.placeholder = "your-worker.workers.dev:443"
o.rmempty = false

o = s:taboption("basic", Value, "token", translate("Auth Token"))
o.description = translate("Authentication token (optional)")
o.password = true
o.placeholder = translate("Leave blank if not required")

o = s:taboption("basic", ListValue, "routing_mode", translate("Routing Mode"))
o.description = translate("Traffic routing strategy")
o:value("bypass_cn", translate("Bypass China Mainland (Recommended)"))
o:value("global", translate("Global Proxy"))
o:value("none", translate("Direct Connection"))
o.default = "bypass_cn"
o.rmempty = false

-- 高级设置标签页
o = s:taboption("advanced", Value, "listen_port", translate("Listen Port"))
o.description = translate("Local proxy listen port")
o.default = "20001"
o.datatype = "port"
o.placeholder = "20001"

o = s:taboption("advanced", Value, "server_ip", translate("Server IP"))
o.description = translate("Preferred IP address or domain to bypass DNS")
o.default = "mfa.gov.ua"
o.placeholder = "mfa.gov.ua"

o = s:taboption("advanced", Value, "fallback_hosts", translate("Fallback Hosts"))
o.description = translate("Fallback target for Workers reverse proxy (IP or domain)")
o.placeholder = "example.com"

o = s:taboption("advanced", Value, "dns_server", translate("DoH Server"))
o.description = translate("DNS over HTTPS server for ECH key query")
o.default = "dns.alidns.com/dns-query"
o.placeholder = "dns.alidns.com/dns-query"

o = s:taboption("advanced", Value, "ech_domain", translate("ECH Domain"))
o.description = translate("Domain for ECH public key lookup")
o.default = "cloudflare-ech.com"
o.placeholder = "cloudflare-ech.com"

-- IP 列表标签页
o = s:taboption("iplist", DummyValue, "_iplist_desc", translate("Description"))
o.rawhtml = true
o.value = translate("Required for 'Bypass China Mainland' mode")

o = s:taboption("iplist", DummyValue, "_iplist_status", translate("List Status"))
o.template = "ech-workers/iplist_status"

o = s:taboption("iplist", Button, "_download", translate("Update List"))
o.inputtitle = translate("Download Now")
o.inputstyle = "apply"
o.template = "ech-workers/download_button"

-- 日志标签页
o = s:taboption("logs", DummyValue, "_logs", translate("Service Log"))
o.template = "ech-workers/logs"

return m
