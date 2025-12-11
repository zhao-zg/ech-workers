# OpenWrt Makefile for ECH-Workers
# Copyright (C) 2025 ECH-Workers

include $(TOPDIR)/rules.mk

PKG_NAME:=ech-workers
PKG_VERSION:=1.1.0
PKG_RELEASE:=1

PKG_BUILD_DIR:=$(BUILD_DIR)/$(PKG_NAME)-$(PKG_VERSION)

PKG_LICENSE:=MIT
PKG_LICENSE_FILES:=LICENSE
PKG_MAINTAINER:=zhao-zg

PKG_BUILD_DEPENDS:=golang/host
PKG_BUILD_PARALLEL:=1
PKG_USE_MIPS16:=0

GO_PKG:=github.com/zhao-zg/ech-workers/tunnel
GO_PKG_BUILD_PKG:=$(GO_PKG)
GO_PKG_LDFLAGS_X:=main.version=$(PKG_VERSION)

include $(INCLUDE_DIR)/package.mk
include $(TOPDIR)/feeds/packages/lang/golang/golang-package.mk

define Package/ech-workers
  SECTION:=net
  CATEGORY:=Network
  SUBMENU:=Web Servers/Proxies
  TITLE:=ECH-enabled SOCKS5/HTTP proxy with intelligent routing
  URL:=https://github.com/zhao-zg/ech-workers
  DEPENDS:=$(GO_ARCH_DEPENDS) +ca-bundle +iptables +ip-full +kmod-tun
endef

define Package/ech-workers/description
  ECH-Workers is a SOCKS5/HTTP proxy client with Encrypted Client Hello (ECH)
  support and intelligent routing capabilities. It can automatically route
  traffic based on destination IP addresses (China mainland bypass mode).
  
  Features:
  - ECH (Encrypted Client Hello) support via TLS 1.3
  - SOCKS5 and HTTP proxy protocols
  - Intelligent routing (global/bypass_cn/direct)
  - IPv4 and IPv6 dual stack support
  - DNS over HTTPS (DoH) support
  - Automatic China IP list management
endef

define Package/ech-workers/conffiles
/etc/config/ech-workers
/etc/ech-workers/
endef

define Build/Prepare
	mkdir -p $(PKG_BUILD_DIR)
	$(CP) ./src/* $(PKG_BUILD_DIR)/
endef

define Build/Compile
	$(call GoPackage/Build/Compile)
endef

define Package/ech-workers/install
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(GO_PKG_BUILD_BIN_DIR)/ech-workers $(1)/usr/bin/
	
	$(INSTALL_DIR) $(1)/etc/init.d
	$(INSTALL_BIN) ./files/ech-workers.init $(1)/etc/init.d/ech-workers
	
	$(INSTALL_DIR) $(1)/etc/config
	$(INSTALL_CONF) ./files/ech-workers.config $(1)/etc/config/ech-workers
	
	$(INSTALL_DIR) $(1)/etc/ech-workers
	$(INSTALL_DATA) ./files/china_ip.txt $(1)/etc/ech-workers/chn_ip.txt
	$(INSTALL_DATA) ./files/china_ipv6.txt $(1)/etc/ech-workers/chn_ip_v6.txt
	
	$(INSTALL_DIR) $(1)/etc/hotplug.d/iface
	$(INSTALL_DATA) ./files/ech-workers.hotplug $(1)/etc/hotplug.d/iface/99-ech-workers
endef

$(eval $(call GoBinPackage,ech-workers))
$(eval $(call BuildPackage,ech-workers))
