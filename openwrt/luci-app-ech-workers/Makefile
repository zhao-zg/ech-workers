# LuCI Makefile for ECH-Workers
# Copyright (C) 2025 ECH-Workers

include $(TOPDIR)/rules.mk

PKG_NAME:=luci-app-ech-workers
PKG_VERSION:=1.1.0
PKG_RELEASE:=1

PKG_LICENSE:=GPL-3.0
PKG_MAINTAINER:=ECH Workers Team

LUCI_TITLE:=LuCI support for ECH-Workers
LUCI_DEPENDS:=+ech-workers +luci-base
LUCI_PKGARCH:=all
LUCI_DESCRIPTION:=ECH-enabled SOCKS5/HTTP proxy with intelligent routing

PKG_BUILD_DEPENDS:=luci-base/host

include $(TOPDIR)/feeds/luci/luci.mk

# call BuildPackage - OpenWrt buildroot signature
