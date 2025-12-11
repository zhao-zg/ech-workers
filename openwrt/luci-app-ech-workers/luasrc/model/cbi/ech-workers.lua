-- Copyright (C) 2025 ECH-Workers

local m, s, o
local sys = require "luci.sys"
local uci = require "luci.model.uci".cursor()

-- 主表单
m = Map("ech-workers", translate("ECH-Workers"))
m.description = translate("ECH-enabled SOCKS5/HTTP proxy with intelligent routing. Supports Encrypted Client Hello (ECH) and automatic China mainland bypass.")

-- ==================== 运行状态 ====================
s = m:section(TypedSection, "ech-workers", translate("Running Status"))
s.anonymous = true
s.addremove = false

o = s:option(DummyValue, "_status", translate("Service Status"))
o.template = "ech-workers/status"
o.value = translate("Collecting data...")

o = s:option(DummyValue, "_proxy_test", translate("Connection Test"))
o.template = "ech-workers/proxy_test"

-- ==================== 基本设置 ====================
s = m:section(TypedSection, "ech-workers", translate("Basic Settings"))
s.anonymous = true
s.addremove = false

o = s:option(Flag, "enabled", translate("Enable"))
o.default = "0"
o.rmempty = false

o = s:option(Value, "server_addr", translate("Server Address"))
o.description = translate("Cloudflare Workers address (e.g., your-worker.workers.dev:443)")
o.placeholder = "your-worker.workers.dev:443"
o.rmempty = false

o = s:option(Value, "token", translate("Auth Token"))
o.description = translate("Authentication token (optional)")
o.password = true
o.placeholder = translate("Leave blank if not required")

o = s:option(ListValue, "routing_mode", translate("Routing Mode"))
o.description = translate("Traffic routing strategy")
o:value("bypass_cn", translate("Bypass China Mainland (Recommended)"))
o:value("global", translate("Global Proxy"))
o:value("none", translate("Direct Connection"))
o.default = "bypass_cn"
o.rmempty = false

-- ==================== 高级设置 ====================
s = m:section(TypedSection, "ech-workers", translate("Advanced Settings"))
s.anonymous = true
s.addremove = false

o = s:option(Value, "listen_port", translate("Listen Port"))
o.description = translate("Local proxy listen port")
o.default = "20001"
o.datatype = "port"
o.placeholder = "20001"

o = s:option(Value, "server_ip", translate("Server IP"))
o.description = translate("Preferred IP address or domain to bypass DNS")
o.default = "mfa.gov.ua"
o.placeholder = "mfa.gov.ua"

o = s:option(Value, "fallback_hosts", translate("Fallback Hosts"))
o.description = translate("Fallback target for Workers reverse proxy (IP or domain)")
o.placeholder = "example.com"

o = s:option(Value, "dns_server", translate("DoH Server"))
o.description = translate("DNS over HTTPS server for ECH key query")
o.default = "dns.alidns.com/dns-query"
o.placeholder = "dns.alidns.com/dns-query"

o = s:option(Value, "ech_domain", translate("ECH Domain"))
o.description = translate("Domain for ECH public key lookup")
o.default = "cloudflare-ech.com"
o.placeholder = "cloudflare-ech.com"

-- ==================== IP 列表管理 ====================
s = m:section(TypedSection, "ech-workers", translate("China IP List"))
s.anonymous = true
s.addremove = false
s.description = translate("Required for 'Bypass China Mainland' mode")

o = s:option(DummyValue, "_iplist_status", translate("List Status"))
o.template = "ech-workers/iplist_status"

o = s:option(Button, "_download", translate("Update List"))
o.inputtitle = translate("Download Now")
o.inputstyle = "apply"
o.template = "ech-workers/download_button"

-- ==================== 版本与更新 ====================
s = m:section(TypedSection, "ech-workers", translate("About"))
s.anonymous = true
s.addremove = false

o = s:option(DummyValue, "_version", translate("Version"))
o.rawhtml = true
function o.cfgvalue(self, section)
	local version = sys.exec("ech-workers -version 2>/dev/null | grep -oP 'v[0-9]+\\.[0-9]+\\.[0-9]+' || echo 'Unknown'")
	return string.format('<span style="color:#090;font-weight:bold">%s</span>', version:gsub("\n", ""))
end

o = s:option(Button, "_check_update", translate("Check Update"))
o.inputtitle = translate("Check Now")
o.inputstyle = "reload"
o.write = function()
	luci.http.redirect(luci.dispatcher.build_url("admin", "services", "ech-workers", "check_update"))
end

-- ==================== 日志查看 ====================
s = m:section(TypedSection, "ech-workers", translate("Logs"))
s.anonymous = true
s.addremove = false

o = s:option(DummyValue, "_logs", translate("Service Log"))
o.template = "ech-workers/logs"

return m
