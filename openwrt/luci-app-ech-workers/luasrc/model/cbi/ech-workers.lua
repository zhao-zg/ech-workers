-- Copyright (C) 2025 ECH-Workers

local m, s, o
local sys = require "luci.sys"
local uci = require "luci.model.uci".cursor()

m = Map("ech-workers", translate("ECH-Workers"), 
	translate("ECH-enabled SOCKS5/HTTP proxy with intelligent routing. " ..
	"Supports Encrypted Client Hello (ECH) and automatic China mainland bypass."))

-- 状态区域
s = m:section(TypedSection, "ech-workers", translate("Service Status"))
s.anonymous = true
s.addremove = false

o = s:option(DummyValue, "_status", translate("Running Status"))
o.template = "ech-workers/status"
o.value = translate("Collecting data...")

-- 基本设置
s = m:section(TypedSection, "ech-workers", translate("Basic Settings"))
s.anonymous = true
s.addremove = false

o = s:option(Flag, "enabled", translate("Enable"))
o.rmempty = false

o = s:option(Value, "listen_addr", translate("Listen Address"),
	translate("Local proxy listen address and port (e.g., 0.0.0.0:1080)"))
o.default = "0.0.0.0:1080"
o.placeholder = "0.0.0.0:1080"
o.rmempty = false

o = s:option(Value, "server_addr", translate("Server Address"),
	translate("Cloudflare Workers server address (format: domain:port/path)"))
o.placeholder = "your-worker.workers.dev:443"
o.rmempty = false

o = s:option(Value, "server_ip", translate("Preferred IP (domain name)"),
	translate("Optional: Specify the server IP or domain name to bypass DNS resolution"))
o.placeholder = "1.2.3.4 or cf.example.com"

o = s:option(Value, "token", translate("Authentication Token"),
	translate("Optional: Token for server authentication"))
o.password = true

-- ECH 设置
s = m:section(TypedSection, "ech-workers", translate("ECH Settings"))
s.anonymous = true
s.addremove = false

o = s:option(Value, "dns_server", translate("DNS Server"),
	translate("DNS over HTTPS (DoH) server for ECH key query"))
o.default = "dns.alidns.com/dns-query"
o.placeholder = "dns.alidns.com/dns-query"

o = s:option(Value, "ech_domain", translate("ECH Domain"),
	translate("Domain name for ECH public key query"))
o.default = "cloudflare-ech.com"
o.placeholder = "cloudflare-ech.com"

-- 分流设置
s = m:section(TypedSection, "ech-workers", translate("Routing Settings"))
s.anonymous = true
s.addremove = false

o = s:option(ListValue, "routing_mode", translate("Routing Mode"),
	translate("Select traffic routing strategy"))
o:value("global", translate("Global Proxy (all traffic via proxy)"))
o:value("bypass_cn", translate("Bypass China Mainland (China IPs direct, others via proxy)"))
o:value("none", translate("Direct Connection (no proxy)"))
o.default = "global"
o.rmempty = false

-- IP 列表管理
s = m:section(TypedSection, "ech-workers", translate("China IP List Management"))
s.anonymous = true
s.addremove = false
s.description = translate("Required for 'Bypass China Mainland' mode. " ..
	"IP lists will be downloaded automatically on first use.")

o = s:option(DummyValue, "_iplist_status", translate("IP List Status"))
o.template = "ech-workers/iplist_status"

o = s:option(Button, "_download", translate("Download/Update IP List"))
o.inputtitle = translate("Download Now")
o.inputstyle = "apply"
o.template = "ech-workers/download_button"

return m
