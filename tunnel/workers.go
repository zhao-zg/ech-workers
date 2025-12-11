package tunnel

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ======================== 全局参数 ========================

var (
	listenAddr    string
	serverAddr    string
	serverIP      string // 优选IP，用于连接Workers服务器
	fallbackHosts string // 反代Host，传递给Workers用于连接真实目标（支持域名和IP）
	token         string
	dnsServer     string
	echDomain     string
	routingMode   string // 分流模式: "global", "bypass_cn", "none"

	echListMu sync.RWMutex
	echList   []byte

	// 中国IP列表（IPv4）
	chinaIPRangesMu sync.RWMutex
	chinaIPRanges   []ipRange

	// 中国IP列表（IPv6）
	chinaIPV6RangesMu sync.RWMutex
	chinaIPV6Ranges   []ipRangeV6

	proxyListener net.Listener

	// 性能优化：缓冲区池复用
	bufferPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 32768)
		},
	}

	// 性能优化：并发连接限制（防止资源耗尽）
	maxConcurrent = make(chan struct{}, 1000) // 最多1000个并发连接

	// 性能优化：IP查询缓存
	ipCacheMu sync.RWMutex
	ipCache   = make(map[string]bool) // 缓存IP是否在中国
)

// ipRange 表示一个IPv4 IP范围
type ipRange struct {
	start uint32
	end   uint32
}

// ipRangeV6 表示一个IPv6 IP范围
type ipRangeV6 struct {
	start [16]byte
	end   [16]byte
}

// StartSocksProxy 启动 SOCKS5/HTTP 代理，供 Android 调用
// ip: 优选IP，用于客户端连接Workers
// fallback: 反代Host，传递给Workers用于连接真实目标（支持域名和IP）
func StartSocksProxy(host, wsServer, dns, ech, ip, fallback, tkn string) error {
	if wsServer == "" {
		return fmt.Errorf("缺少 wss 服务地址")
	}

	listenAddr = host
	serverAddr = wsServer
	serverIP = ip
	fallbackHosts = fallback
	token = tkn

	dnsServer = dns
	if dnsServer == "" {
		dnsServer = "dns.alidns.com/dns-query"
	}

	echDomain = ech
	if echDomain == "" {
		echDomain = "cloudflare-ech.com"
	}

	// 设置默认分流模式为全局代理
	if routingMode == "" {
		routingMode = "global"
	}

	if err := prepareECH(); err != nil {
		return err
	}

	go runProxyServer(listenAddr)
	return nil
}

// StopSocksProxy 停止代理
func StopSocksProxy() {
	go func() {
		if proxyListener != nil {
			_ = proxyListener.Close()
			proxyListener = nil
		}
	}()
}

// SetRoutingMode 设置分流模式
// mode: "global"(全局代理), "bypass_cn"(跳过中国大陆), "none"(不改变代理)
func SetRoutingMode(mode string) {
	routingMode = mode
	log.Printf("[分流] 模式已设置为: %s", mode)
}

// GetRoutingMode 获取当前分流模式
func GetRoutingMode() string {
	return routingMode
}

// LoadChinaIPList 从数据加载中国IP列表
func LoadChinaIPList(ipv4Data, ipv6Data string) error {
	// 加载IPv4
	if ipv4Data != "" {
		var ranges []ipRange
		lines := strings.Split(ipv4Data, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}

			startIP := net.ParseIP(parts[0])
			endIP := net.ParseIP(parts[1])
			if startIP == nil || endIP == nil {
				continue
			}

			start := ipToUint32(startIP)
			end := ipToUint32(endIP)
			if start > 0 && end > 0 && start <= end {
				ranges = append(ranges, ipRange{start: start, end: end})
			}
		}

		if len(ranges) > 0 {
			// 按起始IP排序
			for i := 0; i < len(ranges)-1; i++ {
				for j := i + 1; j < len(ranges); j++ {
					if ranges[i].start > ranges[j].start {
						ranges[i], ranges[j] = ranges[j], ranges[i]
					}
				}
			}

			chinaIPRangesMu.Lock()
			chinaIPRanges = ranges
			chinaIPRangesMu.Unlock()
			log.Printf("[分流] 已加载 %d 个中国IPv4段", len(ranges))
		}
	}

	// 加载IPv6
	if ipv6Data != "" {
		var ranges []ipRangeV6
		lines := strings.Split(ipv6Data, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}

			startIP := net.ParseIP(parts[0])
			endIP := net.ParseIP(parts[1])
			if startIP == nil || endIP == nil {
				continue
			}

			startBytes := startIP.To16()
			endBytes := endIP.To16()
			if startBytes == nil || endBytes == nil {
				continue
			}

			var start, end [16]byte
			copy(start[:], startBytes)
			copy(end[:], endBytes)

			if compareIPv6(start, end) <= 0 {
				ranges = append(ranges, ipRangeV6{start: start, end: end})
			}
		}

		if len(ranges) > 0 {
			// 按起始IP排序
			for i := 0; i < len(ranges)-1; i++ {
				for j := i + 1; j < len(ranges); j++ {
					if compareIPv6(ranges[i].start, ranges[j].start) > 0 {
						ranges[i], ranges[j] = ranges[j], ranges[i]
					}
				}
			}

			chinaIPV6RangesMu.Lock()
			chinaIPV6Ranges = ranges
			chinaIPV6RangesMu.Unlock()
			log.Printf("[分流] 已加载 %d 个中国IPv6段", len(ranges))
		}
	}

	return nil
}

// ======================== 工具函数 ========================

// ipToUint32 将IP地址转换为uint32
func ipToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	if ip == nil {
		return 0
	}
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

// isChinaIP 检查IP是否在中国IP列表中（支持IPv4和IPv6）
func isChinaIP(ipStr string) bool {
	// 性能优化：查询缓存
	ipCacheMu.RLock()
	if cached, ok := ipCache[ipStr]; ok {
		ipCacheMu.RUnlock()
		return cached
	}
	ipCacheMu.RUnlock()

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	var result bool

	// 检查IPv4
	if ip.To4() != nil {
		ipUint32 := ipToUint32(ip)
		if ipUint32 == 0 {
			result = false
		} else {
			chinaIPRangesMu.RLock()
			defer chinaIPRangesMu.RUnlock()

			// 二分查找
			left, right := 0, len(chinaIPRanges)
			for left < right {
				mid := (left + right) / 2
				r := chinaIPRanges[mid]
				if ipUint32 < r.start {
					right = mid
				} else if ipUint32 > r.end {
					left = mid + 1
				} else {
					result = true
					break
				}
			}
		}
	} else {
		// 检查IPv6
		ipBytes := ip.To16()
		if ipBytes == nil {
			result = false
		} else {
			var ipArray [16]byte
			copy(ipArray[:], ipBytes)

			chinaIPV6RangesMu.RLock()
			defer chinaIPV6RangesMu.RUnlock()

			// 二分查找IPv6
			left, right := 0, len(chinaIPV6Ranges)
			for left < right {
				mid := (left + right) / 2
				r := chinaIPV6Ranges[mid]

				// 比较起始IP
				cmpStart := compareIPv6(ipArray, r.start)
				if cmpStart < 0 {
					right = mid
					continue
				}

				// 比较结束IP
				cmpEnd := compareIPv6(ipArray, r.end)
				if cmpEnd > 0 {
					left = mid + 1
					continue
				}

				// 在范围内
				result = true
				break
			}
		}
	}

	// 性能优化：缓存查询结果（限制缓存大小防止内存泄漏）
	ipCacheMu.Lock()
	if len(ipCache) < 10000 { // 最多缓存10000个IP
		ipCache[ipStr] = result
	}
	ipCacheMu.Unlock()

	return result
}

// compareIPv6 比较两个IPv6地址，返回 -1, 0, 或 1
func compareIPv6(a, b [16]byte) int {
	for i := 0; i < 16; i++ {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
	}
	return 0
}

// shouldBypassProxy 根据分流模式判断是否应该绕过代理（直连）
func shouldBypassProxy(targetHost string) bool {
	if routingMode == "none" {
		// "不改变代理"模式：所有流量都直连
		return true
	}
	if routingMode == "global" {
		// "全局代理"模式：所有流量都走代理
		return false
	}
	if routingMode == "bypass_cn" {
		// "跳过中国大陆"模式：检查是否是中国IP
		// 先尝试解析为IP
		if ip := net.ParseIP(targetHost); ip != nil {
			return isChinaIP(targetHost)
		}
		// 如果是域名，先解析IP
		ips, err := net.LookupIP(targetHost)
		if err != nil {
			// 解析失败，默认走代理
			return false
		}
		// 检查所有解析到的IP，如果有一个是中国IP，就直连
		for _, ip := range ips {
			if isChinaIP(ip.String()) {
				return true
			}
		}
		// 都不是中国IP，走代理
		return false
	}
	// 未知模式，默认走代理
	return false
}

func isNormalCloseError(err error) bool {
	if err == nil {
		return false
	}
	if err == io.EOF {
		return true
	}
	errStr := err.Error()
	return strings.Contains(errStr, "use of closed network connection") ||
		strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "connection reset by peer") ||
		strings.Contains(errStr, "normal closure")
}

// ======================== ECH 支持 ========================

const typeHTTPS = 65

func prepareECH() error {
	for {
		log.Printf("[客户端] 使用 DNS 服务器查询 ECH: %s -> %s", dnsServer, echDomain)
		echBase64, err := queryHTTPSRecord(echDomain, dnsServer)
		if err != nil {
			log.Printf("[客户端] DNS 查询失败: %v，2秒后重试...", err)
			time.Sleep(2 * time.Second)
			continue
		}
		if echBase64 == "" {
			log.Printf("[客户端] 未找到 ECH 参数，2秒后重试...")
			time.Sleep(2 * time.Second)
			continue
		}
		raw, err := base64.StdEncoding.DecodeString(echBase64)
		if err != nil {
			log.Printf("[客户端] ECH Base64 解码失败: %v，2秒后重试...", err)
			time.Sleep(2 * time.Second)
			continue
		}
		echListMu.Lock()
		echList = raw
		echListMu.Unlock()
		log.Printf("[客户端] ECHConfigList 长度: %d 字节", len(raw))
		return nil
	}
}

func refreshECH() error {
	log.Printf("[ECH] 刷新 ECH 公钥配置...")
	return prepareECH()
}

func getECHList() ([]byte, error) {
	echListMu.RLock()
	defer echListMu.RUnlock()
	if len(echList) == 0 {
		return nil, errors.New("ECH 配置未加载")
	}
	return echList, nil
}

func buildTLSConfigWithECH(serverName string, echList []byte) (*tls.Config, error) {
	roots, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("加载系统根证书失败: %w", err)
	}

	if echList == nil || len(echList) == 0 {
		return nil, errors.New("ECH 配置为空，这是必需功能")
	}

	config := &tls.Config{
		MinVersion: tls.VersionTLS13,
		ServerName: serverName,
		RootCAs:    roots,
	}

	// 使用反射设置 ECH 字段（ECH 是核心功能，必须设置成功）
	if err := setECHConfig(config, echList); err != nil {
		return nil, fmt.Errorf("设置 ECH 配置失败（需要 Go 1.23+ 或支持 ECH 的版本）: %w", err)
	}

	return config, nil
}

// setECHConfig 使用反射设置 ECH 配置（ECH 是核心功能，必须成功）
func setECHConfig(config *tls.Config, echList []byte) error {
	configValue := reflect.ValueOf(config).Elem()

	// 设置 EncryptedClientHelloConfigList（必需）
	field1 := configValue.FieldByName("EncryptedClientHelloConfigList")
	if !field1.IsValid() || !field1.CanSet() {
		return fmt.Errorf("EncryptedClientHelloConfigList 字段不可用，需要 Go 1.23+ 版本")
	}
	field1.Set(reflect.ValueOf(echList))

	// 设置 EncryptedClientHelloRejectionVerify（必需）
	field2 := configValue.FieldByName("EncryptedClientHelloRejectionVerify")
	if !field2.IsValid() || !field2.CanSet() {
		return fmt.Errorf("EncryptedClientHelloRejectionVerify 字段不可用，需要 Go 1.23+ 版本")
	}
	rejectionFunc := func(cs tls.ConnectionState) error {
		return errors.New("服务器拒绝 ECH")
	}
	field2.Set(reflect.ValueOf(rejectionFunc))

	return nil
}

// queryHTTPSRecord 通过 DoH 查询 HTTPS 记录
func queryHTTPSRecord(domain, server string) (string, error) {
	dohURL := server
	if !strings.HasPrefix(dohURL, "https://") && !strings.HasPrefix(dohURL, "http://") {
		dohURL = "https://" + dohURL
	}
	return queryDoH(domain, dohURL)
}

// queryDoH 执行 DoH 查询（用于获取 ECH 配置）
func queryDoH(domain, dohURL string) (string, error) {
	// 备用 DoH 服务器列表(格式: URL|IP)
	dohServers := []struct {
		url string
		ip  string
	}{
		{dohURL, ""}, // 优先使用用户配置的服务器
		{"https://dns.alidns.com/dns-query", "223.5.5.5"},
		{"https://doh.pub/dns-query", "1.12.12.12"},
		{"https://1.1.1.1/dns-query", "1.1.1.1"},
		{"https://dns.google/dns-query", "8.8.8.8"},
	}

	var lastErr error
	for i, server := range dohServers {
		if i > 0 {
			log.Printf("[DoH] 尝试备用服务器: %s", server.url)
		}
		
		result, err := doSingleDoHQuery(domain, server.url, server.ip)
		if err == nil && result != "" {
			return result, nil
		}
		lastErr = err
		if i > 0 {
			log.Printf("[DoH] 备用服务器 %s 失败: %v", server.url, err)
		}
	}

	return "", fmt.Errorf("所有 DoH 服务器都失败，最后错误: %v", lastErr)
}

// doSingleDoHQuery 执行单次 DoH 查询
func doSingleDoHQuery(domain, dohURL, preferIP string) (string, error) {
	u, err := url.Parse(dohURL)
	if err != nil {
		return "", fmt.Errorf("无效的 DoH URL: %v", err)
	}

	dnsQuery := buildDNSQuery(domain, typeHTTPS)
	dnsBase64 := base64.RawURLEncoding.EncodeToString(dnsQuery)

	q := u.Query()
	q.Set("dns", dnsBase64)
	u.RawQuery = q.Encode()

	// 保存原始主机名用于 SNI 和 Host 头
	originalHost := u.Host
	
	// 如果提供了 IP,替换 URL 中的主机名
	if preferIP != "" {
		// 移除端口号(如果有)
		host := originalHost
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}
		u.Host = preferIP
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}
	
	// 设置正确的 Host 头(使用原始域名,不是IP)
	req.Host = originalHost
	req.Header.Set("Accept", "application/dns-message")
	req.Header.Set("Content-Type", "application/dns-message")

	// 创建自定义 Transport,支持 SNI
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			ServerName: strings.Split(originalHost, ":")[0], // 使用域名作为 SNI
		},
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// 如果使用了 IP,直接连接
			if preferIP != "" {
				addr = net.JoinHostPort(preferIP, "443")
			}
			d := &net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}
			return d.DialContext(ctx, network, addr)
		},
	}

	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("DoH 请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("DoH 服务器返回错误: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取 DoH 响应失败: %v", err)
	}

	return parseDNSResponse(body)
}

func buildDNSQuery(domain string, qtype uint16) []byte {
	query := make([]byte, 0, 512)
	query = append(query, 0x00, 0x01, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00)
	for _, label := range strings.Split(domain, ".") {
		query = append(query, byte(len(label)))
		query = append(query, []byte(label)...)
	}
	query = append(query, 0x00, byte(qtype>>8), byte(qtype), 0x00, 0x01)
	return query
}

func parseDNSResponse(response []byte) (string, error) {
	if len(response) < 12 {
		return "", errors.New("响应过短")
	}
	ancount := binary.BigEndian.Uint16(response[6:8])
	if ancount == 0 {
		return "", errors.New("无应答记录")
	}

	offset := 12
	for offset < len(response) && response[offset] != 0 {
		offset += int(response[offset]) + 1
	}
	offset += 5

	for i := 0; i < int(ancount); i++ {
		if offset >= len(response) {
			break
		}
		if response[offset]&0xC0 == 0xC0 {
			offset += 2
		} else {
			for offset < len(response) && response[offset] != 0 {
				offset += int(response[offset]) + 1
			}
			offset++
		}
		if offset+10 > len(response) {
			break
		}
		rrType := binary.BigEndian.Uint16(response[offset : offset+2])
		offset += 8
		dataLen := binary.BigEndian.Uint16(response[offset : offset+2])
		offset += 2
		if offset+int(dataLen) > len(response) {
			break
		}
		data := response[offset : offset+int(dataLen)]
		offset += int(dataLen)

		if rrType == typeHTTPS {
			if ech := parseHTTPSRecord(data); ech != "" {
				return ech, nil
			}
		}
	}
	return "", nil
}

func parseHTTPSRecord(data []byte) string {
	if len(data) < 2 {
		return ""
	}
	offset := 2
	if offset < len(data) && data[offset] == 0 {
		offset++
	} else {
		for offset < len(data) && data[offset] != 0 {
			offset += int(data[offset]) + 1
		}
		offset++
	}
	for offset+4 <= len(data) {
		key := binary.BigEndian.Uint16(data[offset : offset+2])
		length := binary.BigEndian.Uint16(data[offset+2 : offset+4])
		offset += 4
		if offset+int(length) > len(data) {
			break
		}
		value := data[offset : offset+int(length)]
		offset += int(length)
		if key == 5 {
			return base64.StdEncoding.EncodeToString(value)
		}
	}
	return ""
}

// ======================== DoH 代理支持 ========================

// queryDoHForProxy 通过 ECH 转发 DNS 查询到 Cloudflare DoH
func queryDoHForProxy(dnsQuery []byte) ([]byte, error) {
	_, port, _, err := parseServerAddr(serverAddr)
	if err != nil {
		return nil, err
	}

	// 构建 DoH URL
	dohURL := fmt.Sprintf("https://cloudflare-dns.com:%s/dns-query", port)

	echBytes, err := getECHList()
	if err != nil {
		return nil, fmt.Errorf("获取 ECH 配置失败: %w", err)
	}

	tlsCfg, err := buildTLSConfigWithECH("cloudflare-dns.com", echBytes)
	if err != nil {
		return nil, fmt.Errorf("构建 TLS 配置失败: %w", err)
	}

	// 创建 HTTP 客户端
	transport := &http.Transport{
		TLSClientConfig: tlsCfg,
	}

	// 如果指定了 IP，使用自定义 Dialer
	if serverIP != "" {
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			_, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			dialer := &net.Dialer{
				Timeout: 10 * time.Second,
			}
			return dialer.DialContext(ctx, network, net.JoinHostPort(serverIP, port))
		}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	// 发送 DoH 请求
	req, err := http.NewRequest("POST", dohURL, bytes.NewReader(dnsQuery))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/dns-message")
	req.Header.Set("Accept", "application/dns-message")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("DoH 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DoH 响应错误: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// ======================== WebSocket 客户端 ========================

func parseServerAddr(addr string) (host, port, path string, err error) {
	path = "/"
	slashIdx := strings.Index(addr, "/")
	if slashIdx != -1 {
		path = addr[slashIdx:]
		addr = addr[:slashIdx]
	}

	host, port, err = net.SplitHostPort(addr)
	if err != nil {
		return "", "", "", fmt.Errorf("无效的服务器地址格式: %v", err)
	}

	return host, port, path, nil
}

func dialWebSocketWithECH(maxRetries int) (*websocket.Conn, error) {
	host, port, path, err := parseServerAddr(serverAddr)
	if err != nil {
		return nil, err
	}

	wsURL := fmt.Sprintf("wss://%s:%s%s", host, port, path)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		echBytes, echErr := getECHList()
		if echErr != nil {
			if attempt < maxRetries {
				refreshECH()
				continue
			}
			return nil, echErr
		}

		tlsCfg, tlsErr := buildTLSConfigWithECH(host, echBytes)
		if tlsErr != nil {
			return nil, tlsErr
		}

		dialer := websocket.Dialer{
			TLSClientConfig: tlsCfg,
			Subprotocols: func() []string {
				if token == "" {
					return nil
				}
				return []string{token}
			}(),
			HandshakeTimeout: 10 * time.Second,
			ReadBufferSize:   65536,
			WriteBufferSize:  65536,
		}

		if serverIP != "" {
			dialer.NetDial = func(network, address string) (net.Conn, error) {
				_, port, err := net.SplitHostPort(address)
				if err != nil {
					return nil, err
				}
				return net.DialTimeout(network, net.JoinHostPort(serverIP, port), 10*time.Second)
			}
		}

		wsConn, _, dialErr := dialer.Dial(wsURL, nil)
		if dialErr != nil {
			if strings.Contains(strings.ToLower(dialErr.Error()), "ech") && attempt < maxRetries {
				log.Printf("[ECH] 连接失败，尝试刷新配置 (%d/%d)", attempt, maxRetries)
				refreshECH()
				time.Sleep(time.Second)
				continue
			}
			return nil, dialErr
		}

		return wsConn, nil
	}

	return nil, errors.New("连接失败，已达最大重试次数")
}

// ======================== 统一代理服务器 ========================

func runProxyServer(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("[代理] 监听失败: %v", err)
		return
	}
	proxyListener = listener

	log.Printf("[代理] 服务器启动: %s (支持 SOCKS5 和 HTTP)", addr)
	log.Printf("[代理] 后端服务器: %s", serverAddr)
	if serverIP != "" {
		log.Printf("[代理] 使用固定 IP: %s", serverIP)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) || strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// 性能优化：并发连接限制
	maxConcurrent <- struct{}{}
	defer func() { <-maxConcurrent }()

	// 延迟优化：禁用Nagle算法，减少小包延迟
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)
	}

	clientAddr := conn.RemoteAddr().String()
	conn.SetDeadline(time.Now().Add(30 * time.Second))

	// 尝试获取透明代理的原始目标地址
	if origDest, err := getOriginalDest(conn); err == nil {
		// 透明代理模式：直接转发原始流量
		handleTransparentProxy(conn, clientAddr, origDest)
		return
	}

	// 非透明代理模式：读取第一个字节判断协议
	buf := make([]byte, 1)
	n, err := conn.Read(buf)
	if err != nil || n == 0 {
		return
	}

	firstByte := buf[0]

	// 使用 switch 判断协议类型
	switch firstByte {
	case 0x05:
		// SOCKS5 协议
		handleSOCKS5(conn, clientAddr, firstByte)
	case 'C', 'G', 'P', 'H', 'D', 'O', 'T':
		// HTTP 协议 (CONNECT, GET, POST, HEAD, DELETE, OPTIONS, TRACE, PUT, PATCH)
		handleHTTP(conn, clientAddr, firstByte)
	default:
		log.Printf("[代理] %s 未知协议: 0x%02x", clientAddr, firstByte)
	}
}

// ======================== 透明代理支持 ========================

// getOriginalDest 获取透明代理的原始目标地址（Linux SO_ORIGINAL_DST）
func getOriginalDest(conn net.Conn) (string, error) {
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return "", fmt.Errorf("not a TCP connection")
	}

	// 获取底层文件描述符
	rawConn, err := tcpConn.SyscallConn()
	if err != nil {
		return "", err
	}

	var addr string
	var getErr error
	
	// 在 Linux 上使用 SO_ORIGINAL_DST
	err = rawConn.Control(func(fd uintptr) {
		addr, getErr = getOriginalDestFromFd(fd)
	})
	
	if err != nil {
		return "", err
	}
	if getErr != nil {
		return "", getErr
	}
	
	return addr, nil
}

// getOriginalDestFromFd 从文件描述符获取原始目标地址
// 此函数仅在 Linux 上有效，在其他平台会返回错误
func getOriginalDestFromFd(fd uintptr) (string, error) {
	// 仅在 Linux 上支持透明代理
	// 在 Windows/Mac 上会编译但返回错误
	return getOriginalDestLinux(fd)
}

// handleTransparentProxy 处理透明代理连接
func handleTransparentProxy(conn net.Conn, clientAddr, target string) {
	log.Printf("[透明代理] %s -> %s", clientAddr, target)

	// 直接使用 handleTunnel，mode 设为特殊值表示透明代理
	// 透明代理模式下，客户端发送的第一帧就是应用数据，不需要读取 SOCKS5 握手
	if err := handleTunnel(conn, target, clientAddr, modeTransparent, ""); err != nil {
		log.Printf("[透明代理] %s 代理失败: %v", clientAddr, err)
	}
}

// ======================== SOCKS5 处理 ========================

func handleSOCKS5(conn net.Conn, clientAddr string, firstByte byte) {
	// 验证版本
	if firstByte != 0x05 {
		log.Printf("[SOCKS5] %s 版本错误: 0x%02x", clientAddr, firstByte)
		return
	}

	// 读取认证方法数量
	buf := make([]byte, 1)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return
	}

	nmethods := buf[0]
	methods := make([]byte, nmethods)
	if _, err := io.ReadFull(conn, methods); err != nil {
		return
	}

	// 响应无需认证
	if _, err := conn.Write([]byte{0x05, 0x00}); err != nil {
		return
	}

	// 读取请求
	buf = make([]byte, 4)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return
	}

	if buf[0] != 5 {
		return
	}

	command := buf[1]
	atyp := buf[3]

	var host string
	switch atyp {
	case 0x01: // IPv4
		buf = make([]byte, 4)
		if _, err := io.ReadFull(conn, buf); err != nil {
			return
		}
		host = net.IP(buf).String()

	case 0x03: // 域名
		buf = make([]byte, 1)
		if _, err := io.ReadFull(conn, buf); err != nil {
			return
		}
		domainBuf := make([]byte, buf[0])
		if _, err := io.ReadFull(conn, domainBuf); err != nil {
			return
		}
		host = string(domainBuf)

	case 0x04: // IPv6
		buf = make([]byte, 16)
		if _, err := io.ReadFull(conn, buf); err != nil {
			return
		}
		host = net.IP(buf).String()

	default:
		conn.Write([]byte{0x05, 0x08, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		return
	}

	// 读取端口
	buf = make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return
	}
	port := int(buf[0])<<8 | int(buf[1])

	switch command {
	case 0x01: // CONNECT
		var target string
		if atyp == 0x04 {
			target = fmt.Sprintf("[%s]:%d", host, port)
		} else {
			target = fmt.Sprintf("%s:%d", host, port)
		}

		log.Printf("[SOCKS5] %s -> %s", clientAddr, target)

		if err := handleTunnel(conn, target, clientAddr, modeSOCKS5, ""); err != nil {
			if !isNormalCloseError(err) {
				log.Printf("[SOCKS5] %s 代理失败: %v", clientAddr, err)
			}
		}

	case 0x03: // UDP ASSOCIATE
		handleUDPAssociate(conn, clientAddr)

	default:
		conn.Write([]byte{0x05, 0x07, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		return
	}
}

func handleUDPAssociate(tcpConn net.Conn, clientAddr string) {
	// 创建 UDP 监听器
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		log.Printf("[UDP] %s 解析地址失败: %v", clientAddr, err)
		tcpConn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		return
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Printf("[UDP] %s 监听失败: %v", clientAddr, err)
		tcpConn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		return
	}

	// 获取实际监听的端口
	localAddr := udpConn.LocalAddr().(*net.UDPAddr)
	port := localAddr.Port

	log.Printf("[UDP] %s UDP ASSOCIATE 监听端口: %d", clientAddr, port)

	// 发送成功响应
	response := []byte{0x05, 0x00, 0x00, 0x01}
	response = append(response, 127, 0, 0, 1) // 127.0.0.1
	response = append(response, byte(port>>8), byte(port&0xff))

	if _, err := tcpConn.Write(response); err != nil {
		udpConn.Close()
		return
	}

	// 启动 UDP 处理
	stopChan := make(chan struct{})
	go handleUDPRelay(udpConn, clientAddr, stopChan)

	// 保持 TCP 连接，直到客户端关闭
	buf := make([]byte, 1)
	tcpConn.Read(buf)

	close(stopChan)
	udpConn.Close()
	log.Printf("[UDP] %s UDP ASSOCIATE 连接关闭", clientAddr)
}

func handleUDPRelay(udpConn *net.UDPConn, clientAddr string, stopChan chan struct{}) {
	buf := make([]byte, 65535)
	for {
		select {
		case <-stopChan:
			return
		default:
		}

		udpConn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, addr, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return
		}

		// 解析 SOCKS5 UDP 请求头
		if n < 10 {
			continue
		}

		data := buf[:n]

		if data[2] != 0x00 { // FRAG 必须为 0
			continue
		}

		atyp := data[3]
		var headerLen int
		var dstHost string
		var dstPort int

		switch atyp {
		case 0x01: // IPv4
			if n < 10 {
				continue
			}
			dstHost = net.IP(data[4:8]).String()
			dstPort = int(data[8])<<8 | int(data[9])
			headerLen = 10

		case 0x03: // 域名
			if n < 5 {
				continue
			}
			domainLen := int(data[4])
			if n < 7+domainLen {
				continue
			}
			dstHost = string(data[5 : 5+domainLen])
			dstPort = int(data[5+domainLen])<<8 | int(data[6+domainLen])
			headerLen = 7 + domainLen

		case 0x04: // IPv6
			if n < 22 {
				continue
			}
			dstHost = net.IP(data[4:20]).String()
			dstPort = int(data[20])<<8 | int(data[21])
			headerLen = 22

		default:
			continue
		}

		udpData := data[headerLen:]
		target := fmt.Sprintf("%s:%d", dstHost, dstPort)

		// 检查是否是 DNS 查询（端口 53）
		if dstPort == 53 {
			log.Printf("[UDP-DNS] %s -> %s (DoH 查询)", clientAddr, target)
			go handleDNSQuery(udpConn, addr, udpData, data[:headerLen])
		} else {
			log.Printf("[UDP] %s -> %s (暂不支持非 DNS UDP)", clientAddr, target)
		}
	}
}

func handleDNSQuery(udpConn *net.UDPConn, clientAddr *net.UDPAddr, dnsQuery []byte, socks5Header []byte) {
	// 通过 DoH 查询
	dnsResponse, err := queryDoHForProxy(dnsQuery)
	if err != nil {
		log.Printf("[UDP-DNS] DoH 查询失败: %v", err)
		return
	}

	// 构建 SOCKS5 UDP 响应
	response := make([]byte, 0, len(socks5Header)+len(dnsResponse))
	response = append(response, socks5Header...)
	response = append(response, dnsResponse...)

	// 发送响应
	_, err = udpConn.WriteToUDP(response, clientAddr)
	if err != nil {
		log.Printf("[UDP-DNS] 发送响应失败: %v", err)
		return
	}

	log.Printf("[UDP-DNS] DoH 查询成功，响应 %d 字节", len(dnsResponse))
}

// ======================== HTTP 处理 ========================

func handleHTTP(conn net.Conn, clientAddr string, firstByte byte) {
	// 将第一个字节放回缓冲区
	reader := bufio.NewReader(io.MultiReader(
		strings.NewReader(string(firstByte)),
		conn,
	))

	// 读取 HTTP 请求行
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	parts := strings.Fields(requestLine)
	if len(parts) < 3 {
		return
	}

	method := parts[0]
	requestURL := parts[1]
	httpVersion := parts[2]

	// 读取所有 headers
	headers := make(map[string]string)
	var headerLines []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		headerLines = append(headerLines, line)
		if idx := strings.Index(line, ":"); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			headers[strings.ToLower(key)] = value
		}
	}

	switch method {
	case "CONNECT":
		// HTTPS 隧道代理 - 需要发送 200 响应
		log.Printf("[HTTP-CONNECT] %s -> %s", clientAddr, requestURL)
		if err := handleTunnel(conn, requestURL, clientAddr, modeHTTPConnect, ""); err != nil {
			if !isNormalCloseError(err) {
				log.Printf("[HTTP-CONNECT] %s 代理失败: %v", clientAddr, err)
			}
		}

	case "GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH", "TRACE":
		// HTTP 代理 - 直接转发，不发送 200 响应
		log.Printf("[HTTP-%s] %s -> %s", method, clientAddr, requestURL)

		var target string
		var path string

		if strings.HasPrefix(requestURL, "http://") {
			// 解析完整 URL
			urlWithoutScheme := strings.TrimPrefix(requestURL, "http://")
			idx := strings.Index(urlWithoutScheme, "/")
			if idx > 0 {
				target = urlWithoutScheme[:idx]
				path = urlWithoutScheme[idx:]
			} else {
				target = urlWithoutScheme
				path = "/"
			}
		} else {
			// 相对路径，从 Host header 获取
			target = headers["host"]
			path = requestURL
		}

		if target == "" {
			conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
			return
		}

		// 添加默认端口
		if !strings.Contains(target, ":") {
			target += ":80"
		}

		// 重构 HTTP 请求（去掉完整 URL，使用相对路径）
		var requestBuilder strings.Builder
		requestBuilder.WriteString(fmt.Sprintf("%s %s %s\r\n", method, path, httpVersion))

		// 写入 headers（过滤掉 Proxy-Connection）
		for _, line := range headerLines {
			key := strings.Split(line, ":")[0]
			keyLower := strings.ToLower(strings.TrimSpace(key))
			if keyLower != "proxy-connection" && keyLower != "proxy-authorization" {
				requestBuilder.WriteString(line)
				requestBuilder.WriteString("\r\n")
			}
		}
		requestBuilder.WriteString("\r\n")

		// 如果有请求体，需要读取并附加
		if contentLength := headers["content-length"]; contentLength != "" {
			var length int
			fmt.Sscanf(contentLength, "%d", &length)
			if length > 0 && length < 10*1024*1024 { // 限制 10MB
				body := make([]byte, length)
				if _, err := io.ReadFull(reader, body); err == nil {
					requestBuilder.Write(body)
				}
			}
		}

		firstFrame := requestBuilder.String()

		// 使用 modeHTTPProxy 模式（不发送 200 响应）
		if err := handleTunnel(conn, target, clientAddr, modeHTTPProxy, firstFrame); err != nil {
			if !isNormalCloseError(err) {
				log.Printf("[HTTP-%s] %s 代理失败: %v", method, clientAddr, err)
			}
		}

	default:
		log.Printf("[HTTP] %s 不支持的方法: %s", clientAddr, method)
		conn.Write([]byte("HTTP/1.1 405 Method Not Allowed\r\n\r\n"))
	}
}

// ======================== 通用隧道处理 ========================

// 代理模式常量
const (
	modeSOCKS5      = 1 // SOCKS5 代理
	modeHTTPConnect = 2 // HTTP CONNECT 隧道
	modeHTTPProxy   = 3 // HTTP 普通代理（GET/POST等）
	modeTransparent = 4 // 透明代理（iptables REDIRECT）
)

func handleTunnel(conn net.Conn, target, clientAddr string, mode int, firstFrame string) error {
	// 解析目标地址
	targetHost, _, err := net.SplitHostPort(target)
	if err != nil {
		targetHost = target
	}

	// 检查是否应该绕过代理（直连）
	if shouldBypassProxy(targetHost) {
		log.Printf("[分流] %s -> %s (直连，绕过代理)", clientAddr, target)
		return handleDirectConnection(conn, target, clientAddr, mode, firstFrame)
	}

	// 走代理
	log.Printf("[分流] %s -> %s (通过代理)", clientAddr, target)
	wsConn, err := dialWebSocketWithECH(2)
	if err != nil {
		sendErrorResponse(conn, mode)
		return err
	}
	defer wsConn.Close()

	var mu sync.Mutex

	// 保活（性能优化：减少ping频率）
	stopPing := make(chan bool)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				mu.Lock()
				wsConn.WriteMessage(websocket.PingMessage, nil)
				mu.Unlock()
			case <-stopPing:
				return
			}
		}
	}()
	defer close(stopPing)

	conn.SetDeadline(time.Time{})

	// 如果没有预设的 firstFrame，尝试读取第一帧数据
	// SOCKS5 模式会尝试读取，透明代理模式也需要读取第一帧应用数据
	if firstFrame == "" && (mode == modeSOCKS5 || mode == modeTransparent) {
		_ = conn.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
		buffer := make([]byte, 32768)
		n, _ := conn.Read(buffer)
		_ = conn.SetReadDeadline(time.Time{})
		if n > 0 {
			firstFrame = string(buffer[:n])
		}
	}

	// 发送连接请求，如果设置了fallbackHosts，则传递给服务端用于反代连接
	var connectMsg string
	if fallbackHosts != "" {
		connectMsg = fmt.Sprintf("CONNECT:%s|%s#%s", target, firstFrame, fallbackHosts)
	} else {
		connectMsg = fmt.Sprintf("CONNECT:%s|%s", target, firstFrame)
	}
	mu.Lock()
	err = wsConn.WriteMessage(websocket.TextMessage, []byte(connectMsg))
	mu.Unlock()
	if err != nil {
		sendErrorResponse(conn, mode)
		return err
	}

	// 等待响应
	_, msg, err := wsConn.ReadMessage()
	if err != nil {
		sendErrorResponse(conn, mode)
		return err
	}

	response := string(msg)
	if strings.HasPrefix(response, "ERROR:") {
		sendErrorResponse(conn, mode)
		return errors.New(response)
	}
	if response != "CONNECTED" {
		sendErrorResponse(conn, mode)
		return fmt.Errorf("意外响应: %s", response)
	}

	// 发送成功响应（根据模式不同而不同）
	if err := sendSuccessResponse(conn, mode); err != nil {
		return err
	}

	log.Printf("[代理] %s 已连接: %s", clientAddr, target)

	// 双向转发（性能优化：使用缓冲区池）
	done := make(chan bool, 2)

	// Client -> Server
	go func() {
		buf := bufferPool.Get().([]byte)
		defer bufferPool.Put(buf)
		
		for {
			n, err := conn.Read(buf)
			if err != nil {
				mu.Lock()
				wsConn.WriteMessage(websocket.TextMessage, []byte("CLOSE"))
				mu.Unlock()
				done <- true
				return
			}

			mu.Lock()
			err = wsConn.WriteMessage(websocket.BinaryMessage, buf[:n])
			mu.Unlock()
			if err != nil {
				done <- true
				return
			}
		}
	}()

	// Server -> Client
	go func() {
		for {
			mt, msg, err := wsConn.ReadMessage()
			if err != nil {
				done <- true
				return
			}

			if mt == websocket.TextMessage {
				if string(msg) == "CLOSE" {
					done <- true
					return
				}
			}

			if _, err := conn.Write(msg); err != nil {
				done <- true
				return
			}
		}
	}()

	<-done
	log.Printf("[代理] %s 已断开: %s", clientAddr, target)
	return nil
}

// ======================== 直连处理 ========================

// handleDirectConnection 处理直连（绕过代理）
func handleDirectConnection(conn net.Conn, target, clientAddr string, mode int, firstFrame string) error {
	// 解析目标地址
	host, port, err := net.SplitHostPort(target)
	if err != nil {
		// 如果没有端口，根据模式添加默认端口
		host = target
		if mode == modeHTTPConnect || mode == modeHTTPProxy {
			port = "443"
		} else {
			port = "80"
		}
		target = net.JoinHostPort(host, port)
	}

	// 直接连接到目标
	targetConn, err := net.DialTimeout("tcp", target, 10*time.Second)
	if err != nil {
		sendErrorResponse(conn, mode)
		return fmt.Errorf("直连失败: %w", err)
	}
	defer targetConn.Close()

	// 延迟优化：禁用Nagle算法
	if tcpConn, ok := targetConn.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)
	}

	// 发送成功响应
	if err := sendSuccessResponse(conn, mode); err != nil {
		return err
	}

	// 如果有预设的第一帧数据，先发送
	if firstFrame != "" {
		if _, err := targetConn.Write([]byte(firstFrame)); err != nil {
			return err
		}
	}

	// 双向转发
	done := make(chan bool, 2)

	// Client -> Target
	go func() {
		io.Copy(targetConn, conn)
		done <- true
	}()

	// Target -> Client
	go func() {
		io.Copy(conn, targetConn)
		done <- true
	}()

	<-done
	log.Printf("[分流] %s 直连已断开: %s", clientAddr, target)
	return nil
}

// ======================== 响应辅助函数 ========================

func sendErrorResponse(conn net.Conn, mode int) {
	switch mode {
	case modeSOCKS5:
		conn.Write([]byte{0x05, 0x04, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	case modeHTTPConnect, modeHTTPProxy:
		conn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
	case modeTransparent:
		// 透明代理模式下，直接关闭连接，不发送任何响应
		// 因为客户端不知道被代理，发送任何数据都会导致协议错误
	}
}

func sendSuccessResponse(conn net.Conn, mode int) error {
	switch mode {
	case modeSOCKS5:
		// SOCKS5 成功响应
		_, err := conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		return err
	case modeHTTPConnect:
		// HTTP CONNECT 需要发送 200 响应
		_, err := conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
		return err
	case modeHTTPProxy:
		// HTTP GET/POST 等不需要发送响应，直接转发目标服务器的响应
		return nil
	case modeTransparent:
		// 透明代理模式不需要发送任何响应，直接开始转发数据
		return nil
	}
	return nil
}
