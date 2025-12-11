// +build linux,!android

package tunnel

import (
	"fmt"
	"net"
	"syscall"
	"unsafe"
)

// getOriginalDestLinux 在 Linux 上获取 SO_ORIGINAL_DST (支持 IPv4 和 IPv6)
func getOriginalDestLinux(fd uintptr) (string, error) {
	// 先尝试 IPv6
	addr, err := getOriginalDestIPv6(fd)
	if err == nil {
		return addr, nil
	}

	// 如果 IPv6 失败，尝试 IPv4
	return getOriginalDestIPv4(fd)
}

// getOriginalDestIPv4 获取 IPv4 的原始目标地址
func getOriginalDestIPv4(fd uintptr) (string, error) {
	const SO_ORIGINAL_DST = 80

	// sockaddr_in 结构体 (16 字节)
	var addr [16]byte
	addrLen := uint32(16)

	_, _, errno := syscall.Syscall6(
		syscall.SYS_GETSOCKOPT,
		fd,
		syscall.IPPROTO_IP,
		SO_ORIGINAL_DST,
		uintptr(unsafe.Pointer(&addr[0])),
		uintptr(unsafe.Pointer(&addrLen)),
		0,
	)

	if errno != 0 {
		return "", errno
	}

	// 解析 sockaddr_in 结构
	// 字节布局: [0-1: family, 2-3: port (big-endian), 4-7: IP address, 8-15: padding]
	port := uint16(addr[2])<<8 | uint16(addr[3])
	ip := net.IPv4(addr[4], addr[5], addr[6], addr[7])

	return fmt.Sprintf("%s:%d", ip.String(), port), nil
}

// getOriginalDestIPv6 获取 IPv6 的原始目标地址
func getOriginalDestIPv6(fd uintptr) (string, error) {
	const IP6T_SO_ORIGINAL_DST = 80

	// sockaddr_in6 结构体 (28 字节)
	var addr [28]byte
	addrLen := uint32(28)

	_, _, errno := syscall.Syscall6(
		syscall.SYS_GETSOCKOPT,
		fd,
		syscall.IPPROTO_IPV6,
		IP6T_SO_ORIGINAL_DST,
		uintptr(unsafe.Pointer(&addr[0])),
		uintptr(unsafe.Pointer(&addrLen)),
		0,
	)

	if errno != 0 {
		return "", errno
	}

	// 解析 sockaddr_in6 结构
	// 字节布局: [0-1: family, 2-3: port (big-endian), 4-7: flowinfo, 8-23: IPv6 address, 24-27: scope_id]
	port := uint16(addr[2])<<8 | uint16(addr[3])
	
	// 提取 IPv6 地址 (16 字节)
	var ipv6 [16]byte
	copy(ipv6[:], addr[8:24])
	ip := net.IP(ipv6[:])

	return fmt.Sprintf("[%s]:%d", ip.String(), port), nil
}
