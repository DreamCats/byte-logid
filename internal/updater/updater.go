// Package updater 提供 logid 的自更新功能。
// 从 GitHub Releases 下载最新版本并替换当前二进制。
package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	releaseURL = "https://api.github.com/repos/DreamCats/logid/releases/latest"
	httpTimeout = 30 * time.Second
)

// githubRelease GitHub Release API 响应。
type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

// githubAsset GitHub Release 的附件。
type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Update 执行自更新流程。
//
// 参数:
//   - currentVersion: 当前版本号
//   - checkOnly: 仅检查是否有新版本
//   - force: 强制更新
func Update(currentVersion string, checkOnly bool, force bool) error {
	fmt.Printf("当前版本: %s\n", currentVersion)

	release, err := getLatestRelease()
	if err != nil {
		return err
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentClean := strings.TrimPrefix(currentVersion, "v")
	fmt.Printf("最新版本: %s\n", latestVersion)

	if !force && currentClean >= latestVersion {
		fmt.Println("当前已是最新版本")
		return nil
	}

	if checkOnly {
		if currentClean < latestVersion {
			fmt.Println("有新版本可用，运行 'logid update' 进行更新")
		}
		return nil
	}

	// 查找当前平台的资源
	asset, err := findPlatformAsset(release)
	if err != nil {
		return err
	}

	fmt.Printf("下载文件: %s\n", asset.Name)

	// 获取当前可执行文件路径
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取当前可执行文件路径失败: %w", err)
	}
	currentExe, err = filepath.EvalSymlinks(currentExe)
	if err != nil {
		return fmt.Errorf("解析符号链接失败: %w", err)
	}

	// 下载新版本
	tmpFile, err := downloadAsset(asset)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	// 备份当前版本
	backupPath := currentExe + ".backup"
	if err := copyFile(currentExe, backupPath); err != nil {
		return fmt.Errorf("备份当前版本失败: %w", err)
	}

	// 替换二进制
	if err := copyFile(tmpFile, currentExe); err != nil {
		// 回滚
		_ = copyFile(backupPath, currentExe)
		return fmt.Errorf("替换文件失败: %w", err)
	}

	// 设置权限
	if err := os.Chmod(currentExe, 0755); err != nil {
		return fmt.Errorf("设置文件权限失败: %w", err)
	}

	// 清理备份
	_ = os.Remove(backupPath)

	fmt.Println("更新完成！运行 'logid --version' 验证新版本")
	return nil
}

// getLatestRelease 获取 GitHub 上的最新 Release 信息。
func getLatestRelease() (*githubRelease, error) {
	client := &http.Client{Timeout: httpTimeout}

	req, err := http.NewRequest("GET", releaseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "logid-update")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取最新版本信息失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取最新版本失败 (HTTP %d)", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("解析版本信息失败: %w", err)
	}

	return &release, nil
}

// findPlatformAsset 查找匹配当前平台的下载资源。
func findPlatformAsset(release *githubRelease) (*githubAsset, error) {
	platform := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)

	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, platform) && strings.Contains(asset.Name, "logid") {
			return &asset, nil
		}
	}

	return nil, fmt.Errorf("找不到适用于 %s 平台的发布文件", platform)
}

// downloadAsset 下载资源到临时文件，返回临时文件路径。
func downloadAsset(asset *githubAsset) (string, error) {
	resp, err := http.Get(asset.BrowserDownloadURL)
	if err != nil {
		return "", fmt.Errorf("下载失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载失败 (HTTP %d)", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "logid-update-*")
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("写入临时文件失败: %w", err)
	}

	return tmpFile.Name(), nil
}

// copyFile 复制文件。
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
