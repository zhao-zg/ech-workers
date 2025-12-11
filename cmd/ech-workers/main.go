// +build !android

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	
	"github.com/zhao-zg/ech-workers/tunnel"
)

var (
	listenAddr    string
	serverAddr    string
	serverIP      string
	fallbackHosts string
	token         string
	dnsServer     string
	echDomain     string
	routingMode   string
)

func init() {
	flag.StringVar(&listenAddr, "l", "0.0.0.0:1080", "代理监听地址 (支持 SOCKS5 和 HTTP)")
	flag.StringVar(&serverAddr, "f", "", "服务端地址 (格式: x.x.workers.dev:443)")
	flag.StringVar(&serverIP, "ip", "", "优选IP（域名）- 指定服务端 IP 或域名以绕过 DNS 解析")
	flag.StringVar(&fallbackHosts, "fallback", "", "反代Host - 传递给Workers用于连接真实目标（支持域名和IP）")
	flag.StringVar(&token, "token", "", "身份验证令牌")
	flag.StringVar(&dnsServer, "dns", "dns.alidns.com/dns-query", "ECH 查询 DoH 服务器")
	flag.StringVar(&echDomain, "ech", "cloudflare-ech.com", "ECH 查询域名")
	flag.StringVar(&routingMode, "routing", "global", "分流模式: global(全局代理), bypass_cn(跳过中国大陆), none(不改变代理)")
}

func main() {
	flag.Parse()

	if serverAddr == "" {
		log.Fatal("必须指定服务端地址 -f\n\n示例:\n  ./ech-workers -l 0.0.0.0:1080 -f your-worker.workers.dev:443 -token your-token")
	}

	log.Printf("[启动] ECH-Workers v1.1.0")
	log.Printf("[启动] 监听地址: %s", listenAddr)
	log.Printf("[启动] 服务器地址: %s", serverAddr)
	if serverIP != "" {
		log.Printf("[启动] 优选IP（域名）: %s", serverIP)
	}
	if fallbackHosts != "" {
		log.Printf("[启动] 反代Host: %s", fallbackHosts)
	}
	log.Printf("[启动] 分流模式: %s", routingMode)

	// 设置分流模式
	tunnel.SetRoutingMode(routingMode)

	// 如果是 bypass_cn 模式，加载中国 IP 列表
	if routingMode == "bypass_cn" {
		if err := loadChinaIPListFromFiles(); err != nil {
			log.Printf("[警告] 加载中国 IP 列表失败: %v", err)
		}
	}

	// 启动代理服务
	if err := tunnel.StartSocksProxy(listenAddr, serverAddr, dnsServer, echDomain, serverIP, fallbackHosts, token); err != nil {
		log.Fatalf("[启动] 启动失败: %v", err)
	}

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	log.Printf("[启动] 服务已启动，按 Ctrl+C 停止")
	<-sigChan
	
	log.Printf("[停止] 正在停止服务...")
	tunnel.StopSocksProxy()
	log.Printf("[停止] 服务已停止")
}

func loadChinaIPListFromFiles() error {
	var ipv4Data, ipv6Data string
	
	// 尝试从多个可能的位置加载
	paths := []string{
		"/etc/ech-workers/chn_ip.txt",
		"./chn_ip.txt",
		"/usr/share/ech-workers/chn_ip.txt",
	}
	
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err == nil {
			ipv4Data = string(data)
			log.Printf("[启动] 已加载 IPv4 列表: %s (%d 字节)", path, len(data))
			break
		}
	}
	
	// 加载 IPv6（可选）
	pathsV6 := []string{
		"/etc/ech-workers/chn_ip_v6.txt",
		"./chn_ip_v6.txt",
		"/usr/share/ech-workers/chn_ip_v6.txt",
	}
	
	for _, path := range pathsV6 {
		data, err := os.ReadFile(path)
		if err == nil {
			ipv6Data = string(data)
			log.Printf("[启动] 已加载 IPv6 列表: %s (%d 字节)", path, len(data))
			break
		}
	}
	
	if ipv4Data == "" {
		return os.ErrNotExist
	}
	
	return tunnel.LoadChinaIPList(ipv4Data, ipv6Data)
}
