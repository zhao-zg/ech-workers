# OpenWrt Makefile for ECH-Workers
# Copyright (C) 2025 ECH-Workers

include $(TOPDIR)/rules.mk

PKG_NAME:=ech-workers
# 优先使用环境变量中的版本，否则使用git tag，最后使用默认值
PKG_VERSION:=$(if $(PKG_VERSION_OVERRIDE),$(PKG_VERSION_OVERRIDE),$(shell git describe --tags --always 2>/dev/null | sed 's/^v//' || echo "1.0.12"))
PKG_RELEASE:=

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
  DEPENDS:=+ipset +iptables +kmod-ipt-nat
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

define Package/ech-workers/postinst
#!/bin/sh
[ -n "$${IPKG_INSTROOT}" ] || {
	# 如果服务正在运行,先停止(这会清理防火墙规则)
	if /etc/init.d/ech-workers running 2>/dev/null; then
		/etc/init.d/ech-workers stop 2>/dev/null
	fi
	
	# 如果存在-opkg备份文件，说明配置文件有冲突
	if [ -f /etc/config/ech-workers-opkg ]; then
		# 检查旧配置文件是否存在fallback_hosts选项
		if ! grep -q "option fallback_hosts" /etc/config/ech-workers 2>/dev/null; then
			# 旧配置文件缺少新选项，添加它
			sed -i "/option server_ip/a\	option fallback_hosts ''" /etc/config/ech-workers
		fi
		# 删除-opkg备份文件
		rm -f /etc/config/ech-workers-opkg
	fi
	
	# 确保服务默认不启动(用户需要手动启用)
	/etc/init.d/ech-workers disable 2>/dev/null
	
	echo "ECH-Workers installed. Use 'uci set ech-workers.main.enabled=1; uci commit' to enable."
	exit 0
}
endef

define Package/ech-workers/prerm
#!/bin/sh
[ -n "$${IPKG_INSTROOT}" ] || {
	# 停止服务
	/etc/init.d/ech-workers stop 2>/dev/null
	# 禁用开机自启
	/etc/init.d/ech-workers disable 2>/dev/null
	exit 0
}
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
