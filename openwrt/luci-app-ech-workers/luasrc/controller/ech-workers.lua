-- Copyright (C) 2025 ECH-Workers

module("luci.controller.ech-workers", package.seeall)

function index()
	if not nixio.fs.access("/etc/config/ech-workers") then
		return
	end

	local page
	page = entry({"admin", "services", "ech-workers"}, cbi("ech-workers"), _("ECH-Workers"), 60)
	page.dependent = true
	page.acl_depends = { "luci-app-ech-workers" }
	
	entry({"admin", "services", "ech-workers", "status"}, call("action_status"))
	entry({"admin", "services", "ech-workers", "download_iplist"}, call("action_download_iplist"))
end

function action_status()
	local sys = require "luci.sys"
	local status = {
		running = sys.call("pgrep -f /usr/bin/ech-workers >/dev/null") == 0
	}
	
	-- 检查 IP 列表状态
	local ipv4_file = "/etc/ech-workers/chn_ip.txt"
	local ipv6_file = "/etc/ech-workers/chn_ip_v6.txt"
	
	status.ipv4_loaded = nixio.fs.stat(ipv4_file) and nixio.fs.stat(ipv4_file).size > 0
	status.ipv6_loaded = nixio.fs.stat(ipv6_file) and nixio.fs.stat(ipv6_file).size > 0
	
	if status.ipv4_loaded then
		local mtime = nixio.fs.stat(ipv4_file).mtime
		status.ipv4_update_time = os.date("%Y-%m-%d %H:%M:%S", mtime)
	end
	
	if status.ipv6_loaded then
		local mtime = nixio.fs.stat(ipv6_file).mtime
		status.ipv6_update_time = os.date("%Y-%m-%d %H:%M:%S", mtime)
	end
	
	luci.http.prepare_content("application/json")
	luci.http.write_json(status)
end

function action_download_iplist()
	local sys = require "luci.sys"
	
	-- 执行下载脚本
	local result = sys.exec([[
		ipv4_url="https://raw.githubusercontent.com/mayaxcn/china-ip-list/refs/heads/master/chn_ip.txt"
		ipv6_url="https://raw.githubusercontent.com/mayaxcn/china-ip-list/refs/heads/master/chn_ip_v6.txt"
		
		mkdir -p /etc/ech-workers
		
		success=true
		
		if wget -q -O /etc/ech-workers/chn_ip.txt.tmp "$ipv4_url" 2>/dev/null; then
			mv /etc/ech-workers/chn_ip.txt.tmp /etc/ech-workers/chn_ip.txt
		else
			success=false
			rm -f /etc/ech-workers/chn_ip.txt.tmp
		fi
		
		if wget -q -O /etc/ech-workers/chn_ip_v6.txt.tmp "$ipv6_url" 2>/dev/null; then
			mv /etc/ech-workers/chn_ip_v6.txt.tmp /etc/ech-workers/chn_ip_v6.txt
		else
			rm -f /etc/ech-workers/chn_ip_v6.txt.tmp
		fi
		
		if [ "$success" = "true" ]; then
			echo "success"
		else
			echo "failed"
		fi
	]])
	
	local response = {
		success = result:match("success") ~= nil,
		message = result:match("success") and "IP list downloaded successfully" or "Failed to download IP list"
	}
	
	luci.http.prepare_content("application/json")
	luci.http.write_json(response)
end
