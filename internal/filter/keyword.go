package filter

import "strings"

// KeywordFilter 关键词过滤器。
// 在日志条目级别进行筛选，只保留包含指定关键词的条目。
// 多个关键词之间为 OR 关系（包含任意一个即保留），匹配大小写不敏感。
type KeywordFilter struct {
	// keywords 关键词列表（已转小写）
	keywords []string
}

// NewKeywordFilter 创建关键词过滤器。
// 空的关键词列表表示不启用过滤（保留所有条目）。
func NewKeywordFilter(keywords []string) *KeywordFilter {
	lower := make([]string, len(keywords))
	for i, k := range keywords {
		lower[i] = strings.ToLower(k)
	}
	return &KeywordFilter{keywords: lower}
}

// IsActive 返回过滤器是否启用（关键词列表非空）。
func (f *KeywordFilter) IsActive() bool {
	return len(f.keywords) > 0
}

// Matches 检查文本是否匹配任一关键词。
// 大小写不敏感，纯字符串匹配（非正则）。
func (f *KeywordFilter) Matches(text string) bool {
	if !f.IsActive() {
		return true
	}
	lower := strings.ToLower(text)
	for _, k := range f.keywords {
		if strings.Contains(lower, k) {
			return true
		}
	}
	return false
}
