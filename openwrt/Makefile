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

PKG_BUILD_PARALLEL:=1

include $(INCLUDE_DIR)/package.mk

define Package/ech-workers
  SECTION:=net
  CATEGORY:=Network
  SUBMENU:=Web Servers/Proxies
  TITLE:=ECH-enabled SOCKS5/HTTP proxy with intelligent routing
  URL:=https://github.com/zhao-zg/ech-workers
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
	# Go binary will be pre-compiled by GitHub Actions
endef

define Build/Compile
	# Binary already compiled, skip
endef

define Package/ech-workers/install
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) ./ech-workers $(1)/usr/bin/ech-workers
	
	$(INSTALL_DIR) $(1)/etc/init.d
	$(INSTALL_BIN) ./files/ech-workers.init $(1)/etc/init.d/ech-workers
	
	$(INSTALL_DIR) $(1)/etc/config
	$(INSTALL_CONF) ./files/ech-workers.config $(1)/etc/config/ech-workers
	
	$(INSTALL_DIR) $(1)/etc/ech-workers
	
	$(INSTALL_DIR) $(1)/etc/hotplug.d/iface
	$(INSTALL_DATA) ./files/ech-workers.hotplug $(1)/etc/hotplug.d/iface/99-ech-workers
endef

$(eval $(call BuildPackage,ech-workers))
