package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// AppConfig 管理 ~/.config/byte-logid/ 下的应用配置。
type AppConfig struct {
	// ConfigDir 配置目录路径
	ConfigDir string
}

// NewAppConfig 创建配置管理器，确保配置目录存在。
// 目录权限设置为 0700。
func NewAppConfig() (*AppConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("获取用户主目录失败: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "byte-logid")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, fmt.Errorf("创建配置目录 %s 失败: %w", configDir, err)
	}

	return &AppConfig{ConfigDir: configDir}, nil
}

// FiltersPath 返回 filters.json 的完整路径。
func (c *AppConfig) FiltersPath() string {
	return filepath.Join(c.ConfigDir, "filters.json")
}

// EnsureFilters 确保 filters.json 存在。
// 如果文件不存在，则生成包含默认规则的配置文件，权限设置为 0600。
func (c *AppConfig) EnsureFilters() error {
	path := c.FiltersPath()
	if _, err := os.Stat(path); err == nil {
		return nil // 文件已存在
	}

	// 生成默认配置
	fc := &FilterConfig{MsgFilters: DefaultFilters()}
	return fc.Save(path)
}
