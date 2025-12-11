// +build !linux

package tunnel

import (
	"fmt"
)

// getOriginalDestLinux 在非 Linux 平台上返回不支持错误
func getOriginalDestLinux(fd uintptr) (string, error) {
	return "", fmt.Errorf("transparent proxy is only supported on Linux")
}
