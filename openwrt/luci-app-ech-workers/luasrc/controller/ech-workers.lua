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
	entry({"admin", "services", "ech-workers", "test_proxy"}, call("action_test_proxy"))
	entry({"admin", "services", "ech-workers", "get_logs"}, call("action_get_logs"))
end

function action_status()
	local sys = require "luci.sys"
	local uci = require "luci.model.uci".cursor()
	local status = {
		running = sys.call("pgrep -f /usr/bin/ech-workers >/dev/null") == 0
	}
	
	-- 如果服务运行中，获取IP、国家和延迟
	if status.running then
		-- 获取优选IP用于测延迟
		local server_ip = uci:get("ech-workers", "main", "server_ip") or "mfa.gov.ua"
		
		-- 测试延迟
		local latency_cmd = string.format(
			"timeout 5 sh -c 'start=$(date +%%s%%N); nc -z -w 3 %s 443 && echo $(($(date +%%s%%N)-start))/1000000 | bc' 2>/dev/null",
			server_ip
		)
		local latency_result = sys.exec(latency_cmd):match("^%s*(.-)%s*$")
		if latency_result and latency_result ~= "" then
			status.latency = tonumber(latency_result)
		end
		
		-- 获取出口IP和国家（使用ipapi.co）
		local ip_info = sys.exec("timeout 5 wget -qO- https://ipapi.co/json/ 2>/dev/null")
		if ip_info and ip_info ~= "" then
			local json = require "luci.jsonc"
			local info = json.parse(ip_info)
			if info then
				status.ip = info.ip
				status.country = info.country_code
			end
		end
	end
	
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

function action_test_proxy()
	local sys = require "luci.sys"
	local json = require "luci.jsonc"
	
	-- 获取测试目标
	luci.http.prepare_content("application/json")
	
	local target = luci.http.formvalue("target") or "google"
	local proxy_addr = luci.http.formvalue("proxy") or "127.0.0.1:1080"
	
	-- 检查服务是否运行
	if sys.call("pgrep -f /usr/bin/ech-workers >/dev/null") ~= 0 then
		luci.http.write_json({
			success = false,
			message = "Service not running",
			error = "Please start the service before testing"
		})
		return
	end
	
	-- 测试目标映射
	local test_targets = {
		google = {url = "https://www.google.com", name = "Google"},
		youtube = {url = "https://www.youtube.com", name = "YouTube"},
		openai = {url = "https://chat.openai.com", name = "OpenAI"},
		twitter = {url = "https://twitter.com", name = "Twitter/X"}
	}
	
	local test_info = test_targets[target] or test_targets["google"]
	
	-- 使用curl通过SOCKS5代理测试
	local start_time = os.clock()
	local cmd = string.format(
		"curl -x socks5h://%s -m 10 -s -o /dev/null -w '%%{http_code}|%%{time_total}' '%s' 2>&1",
		proxy_addr,
		test_info.url
	)
	
	local result = sys.exec(cmd)
	local elapsed = (os.clock() - start_time) * 1000
	
	-- 解析结果
	local http_code, time_total = result:match("(%d+)|([%d%.]+)")
	
	local response = {
		target = test_info.name,
		url = test_info.url
	}
	
	if http_code and tonumber(http_code) then
		local code = tonumber(http_code)
		response.success = (code >= 200 and code < 400)
		response.http_code = code
		response.response_time = math.floor(tonumber(time_total) * 1000)
		response.message = response.success and "Connection successful" or "Connection failed"
	else
		response.success = false
		response.message = "Connection timeout or error"
		response.error = result
	end
	
	luci.http.write_json(response)
end

function action_get_logs()
	local sys = require "luci.sys"
	
	luci.http.prepare_content("application/json")
	
	local lines = luci.http.formvalue("lines") or "100"
	
	-- 从 logread 读取日志，过滤 ech-workers 相关
	local cmd = string.format("logread | grep ech-workers | tail -n %s", lines)
	local logs = sys.exec(cmd)
	
	-- 如果没有日志，返回提示信息
	if logs == "" or logs == nil then
		logs = "No logs available. Service may not be running or no events logged yet."
	end
	
	luci.http.write_json({
		success = true,
		logs = logs
	})
end
