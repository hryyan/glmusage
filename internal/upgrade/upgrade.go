package upgrade

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
)

const repo = "hryyan/glmusage"

// Release 表示 GitHub Release 信息。
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset 表示 Release 附件。
type Asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

// CheckLatest 查询 GitHub 最新 release。
func CheckLatest(ctx context.Context) (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 GitHub API 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("GitHub API 返回 %d: %s", resp.StatusCode, string(body))
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}
	return &release, nil
}

// DoUpgrade 从 GitHub 下载最新版本并替换当前二进制。
func DoUpgrade(ctx context.Context, currentVersion string) error {
	release, err := CheckLatest(ctx)
	if err != nil {
		return err
	}

	if currentVersion == "dev" {
		fmt.Println("开发版本，跳过版本检查直接升级")
	} else if compareVersions(currentVersion, release.TagName) >= 0 {
		fmt.Printf("当前版本 %s >= 远程版本 %s，无需升级\n", currentVersion, release.TagName)
		return nil
	}

	asset := findAsset(release.Assets)
	if asset == nil {
		return fmt.Errorf("未找到适用于 %s/%s 的构建文件", runtime.GOOS, runtime.GOARCH)
	}

	fmt.Printf("正在下载 %s ...\n", asset.Name)

	resp, err := http.Get(asset.URL)
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败: HTTP %d", resp.StatusCode)
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取当前可执行文件路径失败: %w", err)
	}

	tmp := exe + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("写入文件失败: %w", err)
	}
	f.Close()

	if err := os.Rename(tmp, exe); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("替换可执行文件失败: %w (可能需要 sudo)", err)
	}

	fmt.Printf("已升级到 %s\n", release.TagName)
	return nil
}

func compareVersions(v1, v2 string) int {
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")
	p1 := strings.Split(v1, ".")
	p2 := strings.Split(v2, ".")
	maxLen := len(p1)
	if len(p2) > maxLen {
		maxLen = len(p2)
	}
	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(p1) {
			n1, _ = strconv.Atoi(p1[i])
		}
		if i < len(p2) {
			n2, _ = strconv.Atoi(p2[i])
		}
		if n1 != n2 {
			if n1 < n2 {
				return -1
			}
			return 1
		}
	}
	return 0
}

func findAsset(assets []Asset) *Asset {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	suffix := ""
	if goos == "windows" {
		suffix = ".exe"
	}
	target := fmt.Sprintf("glmusage-%s-%s%s", goos, goarch, suffix)

	for i := range assets {
		if strings.EqualFold(assets[i].Name, target) {
			return &assets[i]
		}
	}
	return nil
}
