// Package filter 提供日志消息的过滤功能，包括消息净化和关键词筛选。
package filter

import (
	"fmt"
	"regexp"
	"strings"
)

// MessageSanitizer 消息净化器。
// 使用预编译的正则表达式对日志消息进行净化，剔除合规日志、足迹数据等冗余内容。
type MessageSanitizer struct {
	// filters 预编译的正则过滤器
	filters []*regexp.Regexp
}

// NewMessageSanitizer 从过滤规则字符串列表创建消息净化器。
// 每条规则为一个正则表达式，无效的正则会返回错误。
func NewMessageSanitizer(patterns []string) (*MessageSanitizer, error) {
	filters := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, fmt.Errorf("无效的过滤正则表达式 '%s': %w", p, err)
		}
		filters = append(filters, re)
	}
	return &MessageSanitizer{filters: filters}, nil
}

// Sanitize 对消息内容进行净化，依次应用所有过滤规则。
// 返回净化后的文本。
func (s *MessageSanitizer) Sanitize(message string) string {
	result := message
	for _, re := range s.filters {
		result = re.ReplaceAllString(result, "")
	}

	// 清理多余空格
	spaceRe := regexp.MustCompile(`[ \t]{2,}`)
	result = spaceRe.ReplaceAllString(result, " ")

	// 清理多余空行
	newlineRe := regexp.MustCompile(`\n\s*\n\s*\n`)
	result = newlineRe.ReplaceAllString(result, "\n\n")

	return strings.TrimSpace(result)
}
