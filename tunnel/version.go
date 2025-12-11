package tunnel

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// 当前版本
	Version = "v1.0.7"
	
	// GitHub Release API
	githubAPIURL = "https://api.github.com/repos/zhao-zg/ech-workers/releases/latest"
)

// GitHubRelease GitHub Release 响应结构
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	PublishedAt time.Time `json:"published_at"`
	Body        string    `json:"body"`
	HTMLURL     string    `json:"html_url"`
}

// CompareVersion 比较版本号
// 返回值: 1 表示 v1 > v2, -1 表示 v1 < v2, 0 表示相等
func CompareVersion(v1, v2 string) int {
	// 移除 'v' 前缀
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")
	
	// 分割版本号
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")
	
	// 比较每个部分
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}
	
	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		
		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &n1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &n2)
		}
		
		if n1 > n2 {
			return 1
		} else if n1 < n2 {
			return -1
		}
	}
	
	return 0
}

// CheckUpdate 检查是否有新版本
func CheckUpdate() (*GitHubRelease, bool, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	req, err := http.NewRequest("GET", githubAPIURL, nil)
	if err != nil {
		return nil, false, fmt.Errorf("创建请求失败: %v", err)
	}
	
	// 添加 User-Agent 避免被 GitHub 限制
	req.Header.Set("User-Agent", "ECH-Workers/"+Version)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("GitHub API 返回错误: %s", resp.Status)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, fmt.Errorf("读取响应失败: %v", err)
	}
	
	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, false, fmt.Errorf("解析响应失败: %v", err)
	}
	
	// 比较版本号
	hasUpdate := CompareVersion(release.TagName, Version) > 0
	
	return &release, hasUpdate, nil
}

// PrintUpdateInfo 打印更新信息
func PrintUpdateInfo(release *GitHubRelease) {
	fmt.Printf("\n========================================\n")
	fmt.Printf("🎉 发现新版本!\n")
	fmt.Printf("========================================\n")
	fmt.Printf("当前版本: %s\n", Version)
	fmt.Printf("最新版本: %s\n", release.TagName)
	fmt.Printf("发布时间: %s\n", release.PublishedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("\n更新内容:\n%s\n", release.Body)
	fmt.Printf("\n下载地址: %s\n", release.HTMLURL)
	fmt.Printf("========================================\n\n")
}

// CheckUpdateAsync 异步检查更新(不阻塞启动)
func CheckUpdateAsync() {
	go func() {
		// 延迟 2 秒再检查,避免影响启动速度
		time.Sleep(2 * time.Second)
		
		release, hasUpdate, err := CheckUpdate()
		if err != nil {
			// 静默失败,不影响程序运行
			return
		}
		
		if hasUpdate {
			PrintUpdateInfo(release)
		}
	}()
}
