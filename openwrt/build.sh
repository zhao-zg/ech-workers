#!/bin/bash
# ECH-Workers OpenWrt 包构建脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}ECH-Workers OpenWrt 包构建脚本${NC}"
echo -e "${GREEN}================================${NC}"

# 检查必要工具
echo -e "\n${YELLOW}检查必要工具...${NC}"
for cmd in go git wget; do
    if ! command -v $cmd &> /dev/null; then
        echo -e "${RED}错误: 未找到 $cmd${NC}"
        exit 1
    fi
done
echo -e "${GREEN}✓ 工具检查完成${NC}"

# 创建工作目录
WORK_DIR=$(pwd)
BUILD_DIR="${WORK_DIR}/build"
OPENWRT_DIR="${BUILD_DIR}/openwrt"

echo -e "\n${YELLOW}创建构建目录...${NC}"
mkdir -p "${BUILD_DIR}"
mkdir -p "${OPENWRT_DIR}"

# 下载中国 IP 列表
echo -e "\n${YELLOW}下载中国 IP 列表...${NC}"
IPV4_URL="https://raw.githubusercontent.com/mayaxcn/china-ip-list/refs/heads/master/chn_ip.txt"
IPV6_URL="https://raw.githubusercontent.com/mayaxcn/china-ip-list/refs/heads/master/chn_ip_v6.txt"

wget -q -O "${WORK_DIR}/openwrt/files/china_ip.txt" "$IPV4_URL" || {
    echo -e "${YELLOW}警告: 下载 IPv4 列表失败${NC}"
    touch "${WORK_DIR}/openwrt/files/china_ip.txt"
}

wget -q -O "${WORK_DIR}/openwrt/files/china_ipv6.txt" "$IPV6_URL" || {
    echo -e "${YELLOW}警告: 下载 IPv6 列表失败${NC}"
    touch "${WORK_DIR}/openwrt/files/china_ipv6.txt"
}

echo -e "${GREEN}✓ IP 列表下载完成${NC}"

# 编译 Go 程序（多架构）
echo -e "\n${YELLOW}编译 Go 程序...${NC}"

ARCHS=("amd64" "arm" "arm64" "mipsle" "mips")
GOARMS=("" "7" "" "" "")

for i in "${!ARCHS[@]}"; do
    ARCH="${ARCHS[$i]}"
    GOARM="${GOARMS[$i]}"
    
    echo -e "${YELLOW}编译 ${ARCH}...${NC}"
    
    OUTPUT_DIR="${BUILD_DIR}/bin/${ARCH}"
    mkdir -p "${OUTPUT_DIR}"
    
    if [ "$ARCH" = "arm" ]; then
        GOOS=linux GOARCH=$ARCH GOARM=$GOARM CGO_ENABLED=0 \
            go build -ldflags="-s -w" -o "${OUTPUT_DIR}/ech-workers" ./cmd/ech-workers
    else
        GOOS=linux GOARCH=$ARCH CGO_ENABLED=0 \
            go build -ldflags="-s -w" -o "${OUTPUT_DIR}/ech-workers" ./cmd/ech-workers
    fi
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ ${ARCH} 编译成功${NC}"
        
        # 压缩二进制文件
        if command -v upx &> /dev/null; then
            upx -9 "${OUTPUT_DIR}/ech-workers" 2>/dev/null || true
        fi
    else
        echo -e "${RED}✗ ${ARCH} 编译失败${NC}"
    fi
done

# 复制文件到 OpenWrt 包目录
echo -e "\n${YELLOW}准备 OpenWrt 包...${NC}"

PKG_DIR="${OPENWRT_DIR}/ech-workers"
mkdir -p "${PKG_DIR}"

cp -r openwrt/* "${PKG_DIR}/"

# 创建 ipk 包信息
echo -e "\n${YELLOW}生成包信息...${NC}"
cat > "${BUILD_DIR}/build_info.txt" << EOF
ECH-Workers OpenWrt Package Build Info
======================================

Build Date: $(date)
Version: 1.1.0

Architectures:
$(for arch in "${ARCHS[@]}"; do
    if [ -f "${BUILD_DIR}/bin/${arch}/ech-workers" ]; then
        size=$(ls -lh "${BUILD_DIR}/bin/${arch}/ech-workers" | awk '{print $5}')
        echo "  - ${arch}: ${size}"
    fi
done)

Files:
  - Makefile
  - Init script
  - Config file
  - LuCI web interface
  - China IP lists (IPv4 + IPv6)

Installation:
  1. Copy ech-workers package to OpenWrt build system
  2. Run: make package/ech-workers/compile V=s
  3. Install the generated ipk file

EOF

cat "${BUILD_DIR}/build_info.txt"

echo -e "\n${GREEN}================================${NC}"
echo -e "${GREEN}构建完成！${NC}"
echo -e "${GREEN}================================${NC}"
echo -e "\n输出目录: ${BUILD_DIR}"
echo -e "OpenWrt 包目录: ${PKG_DIR}"
echo -e "\n${YELLOW}下一步:${NC}"
echo -e "1. 将 ${PKG_DIR} 复制到 OpenWrt SDK 的 package/ech-workers 目录"
echo -e "2. 运行: make package/ech-workers/compile V=s"
echo -e "3. 在 bin/packages/*/packages/ 找到生成的 ipk 文件"
