# LuCI Makefile for ECH-Workers
# Copyright (C) 2025 ECH-Workers

include $(TOPDIR)/rules.mk

PKG_NAME:=luci-app-ech-workers
PKG_VERSION:=$(shell git -C "$(CURDIR)/../.." describe --tags --always 2>/dev/null | sed 's/^v//' || echo "1.0.0")
PKG_RELEASE:=

PKG_LICENSE:=GPL-3.0
PKG_MAINTAINER:=zhao-zg

include $(INCLUDE_DIR)/package.mk

define Package/luci-app-ech-workers
  SECTION:=luci
  CATEGORY:=LuCI
  SUBMENU:=3. Applications
  TITLE:=LuCI support for ECH-Workers
  DEPENDS:=+ech-workers
  PKGARCH:=all
endef

define Package/luci-app-ech-workers/description
  LuCI web interface for ECH-Workers proxy client.
  Provides easy configuration and management through OpenWrt's web interface.
endef

define Build/Compile
endef

define Package/luci-app-ech-workers/install
	$(INSTALL_DIR) $(1)/usr/lib/lua/luci/controller
	$(INSTALL_DATA) ./luasrc/controller/ech-workers.lua $(1)/usr/lib/lua/luci/controller/
	
	$(INSTALL_DIR) $(1)/usr/lib/lua/luci/model/cbi
	$(INSTALL_DATA) ./luasrc/model/cbi/ech-workers.lua $(1)/usr/lib/lua/luci/model/cbi/
	
	$(INSTALL_DIR) $(1)/usr/lib/lua/luci/view/ech-workers
	$(INSTALL_DATA) ./luasrc/view/ech-workers/*.htm $(1)/usr/lib/lua/luci/view/ech-workers/
	
	$(INSTALL_DIR) $(1)/usr/share/rpcd/acl.d
	$(CP) ./root/usr/share/rpcd/acl.d/*.json $(1)/usr/share/rpcd/acl.d/ 2>/dev/null || true
endef

$(eval $(call BuildPackage,luci-app-ech-workers))
