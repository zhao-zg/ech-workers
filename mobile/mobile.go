package mobile

// Package mobile provides gomobile-compatible bindings for ECH-Workers

import (
	"github.com/zhao-zg/ech-workers/tunnel"
)

// Proxy represents the proxy client
type Proxy struct {
	running bool
}

// NewProxy creates a new proxy instance
func NewProxy() *Proxy {
	return &Proxy{running: false}
}

// Start starts the SOCKS proxy
// Returns empty string on success, error message on failure
func (p *Proxy) Start(host, wsServer, dns, ech, ip, token string) string {
	err := tunnel.StartSocksProxy(host, wsServer, dns, ech, ip, token)
	if err != nil {
		return err.Error()
	}
	p.running = true
	return ""
}

// Stop stops the proxy
func (p *Proxy) Stop() {
	tunnel.StopSocksProxy()
	p.running = false
}

// SetRoutingMode sets the routing mode
// mode: "global", "bypass_cn", or "none"
func (p *Proxy) SetRoutingMode(mode string) {
	tunnel.SetRoutingMode(mode)
}

// GetRoutingMode gets current routing mode
func (p *Proxy) GetRoutingMode() string {
	return tunnel.GetRoutingMode()
}

// LoadChinaIPList loads China IP list
// Returns empty string on success, error message on failure
func (p *Proxy) LoadChinaIPList(ipv4Data, ipv6Data string) string {
	err := tunnel.LoadChinaIPList(ipv4Data, ipv6Data)
	if err != nil {
		return err.Error()
	}
	return ""
}

// IsRunning returns true if proxy is running
func (p *Proxy) IsRunning() bool {
	return p.running
}
