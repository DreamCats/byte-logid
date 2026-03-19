package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// FilterConfig 消息净化过滤配置。
// 从 ~/.config/logid/filters.json 读取，配置文件为唯一规则来源。
type FilterConfig struct {
	// MsgFilters 正则过滤规则列表
	MsgFilters []string `json:"msg_filters"`
}

// DefaultFilters 返回默认的消息净化过滤规则。
// 用于首次初始化和 reset 命令。
func DefaultFilters() []string {
	return []string{
		"_compliance_nlp_log",
		"_compliance_whitelist_log",
		"_compliance_source=footprint",
		`(?s)"user_extra":\s*"\{.*?\}"`,
		`(?m)"LogID":\s*"[^"]*"`,
		`(?m)"Addr":\s*"[^"]*"`,
		`(?m)"Client":\s*"[^"]*"`,
		// RPC 远程 IP，被通配后无信息量，group 里已有 ipv4
		`\{\{rip=[^}]*\}\}`,
	}
}

// Load 从指定路径加载过滤配置。
func Load(path string) (*FilterConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取过滤配置文件 %s 失败: %w", path, err)
	}

	var fc FilterConfig
	if err := json.Unmarshal(data, &fc); err != nil {
		return nil, fmt.Errorf("解析过滤配置文件 %s 失败: %w", path, err)
	}

	return &fc, nil
}

// Save 将过滤配置保存到指定路径。
// 文件权限设置为 0600。
func (fc *FilterConfig) Save(path string) error {
	data, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化过滤配置失败: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("写入过滤配置文件 %s 失败: %w", path, err)
	}

	return nil
}

// AddFilter 添加一条过滤规则。
func (fc *FilterConfig) AddFilter(pattern string) {
	fc.MsgFilters = append(fc.MsgFilters, pattern)
}

// RemoveFilter 按索引删除一条过滤规则。
// 返回被删除的规则内容，索引越界时返回错误。
func (fc *FilterConfig) RemoveFilter(index int) (string, error) {
	if index < 0 || index >= len(fc.MsgFilters) {
		return "", fmt.Errorf("过滤规则索引 %d 越界，当前共 %d 条规则", index, len(fc.MsgFilters))
	}

	removed := fc.MsgFilters[index]
	fc.MsgFilters = append(fc.MsgFilters[:index], fc.MsgFilters[index+1:]...)
	return removed, nil
}

// Reset 重置为默认过滤规则。
func (fc *FilterConfig) Reset() {
	fc.MsgFilters = DefaultFilters()
}
