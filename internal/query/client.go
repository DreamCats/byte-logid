package query

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/DreamCats/logid/internal/config"
	"github.com/DreamCats/logid/internal/filter"
)

const (
	// defaultScanSpanMin 默认扫描时间范围（分钟）
	defaultScanSpanMin = 10
	// httpTimeout HTTP 请求超时时间
	httpTimeout = 30 * time.Second
	// userAgent 模拟浏览器 UA
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36 Edg/140.0.0.0"
)

// Client 日志查询客户端。
// 负责构造请求、发送查询、提取并过滤日志消息。
type Client struct {
	httpClient   *http.Client
	regionConfig *config.RegionConfig
	sanitizer    *filter.MessageSanitizer
	keyword      *filter.KeywordFilter
	maxLen       int // 消息最大长度，0 表示不截断
}

// NewClient 创建日志查询客户端。
//
// 参数:
//   - regionConfig: 区域配置（包含日志服务 URL 等）
//   - sanitizer: 消息净化器
//   - keyword: 关键词过滤器
//   - maxLen: 消息最大长度，超出截断并标记；0 表示不截断
func NewClient(regionConfig *config.RegionConfig, sanitizer *filter.MessageSanitizer, keyword *filter.KeywordFilter, maxLen int) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	// 从环境变量获取代理
	if proxy := os.Getenv("HTTPS_PROXY"); proxy != "" {
		transport.Proxy = http.ProxyFromEnvironment
	} else if proxy := os.Getenv("HTTP_PROXY"); proxy != "" {
		transport.Proxy = http.ProxyFromEnvironment
	}

	return &Client{
		httpClient: &http.Client{
			Timeout:   httpTimeout,
			Transport: transport,
		},
		regionConfig: regionConfig,
		sanitizer:    sanitizer,
		keyword:      keyword,
		maxLen:       maxLen,
	}
}

// Query 执行日志查询。
//
// 参数:
//   - logid: 要查询的 Log ID
//   - token: JWT 认证 token
//   - psmList: PSM 过滤列表（服务端过滤）
//
// 返回格式化后的 QueryResult，包含经过净化和关键词过滤的日志消息。
func (c *Client) Query(logid string, token string, psmList []string) (*QueryResult, error) {
	if !c.regionConfig.Configured {
		return nil, fmt.Errorf("区域 %s 尚未配置日志服务", c.regionConfig.Region)
	}

	// 构造请求体
	reqBody := LogQueryRequest{
		LogID:         logid,
		PSMList:       psmList,
		ScanSpanInMin: defaultScanSpanMin,
		VRegion:       c.regionConfig.VRegion,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 构造 HTTP 请求
	req, err := http.NewRequest("POST", c.regionConfig.LogServiceURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}

	req.Header.Set("X-Jwt-Token", token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("User-Agent", userAgent)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("日志查询请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("日志查询失败 (HTTP %d): %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var rawResp map[string]json.RawMessage
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}
	if err := json.Unmarshal(respBody, &rawResp); err != nil {
		return nil, fmt.Errorf("解析响应 JSON 失败: %w", err)
	}

	// 提取 data 部分
	var logData LogData
	if dataRaw, ok := rawResp["data"]; ok {
		if err := json.Unmarshal(dataRaw, &logData); err != nil {
			return nil, fmt.Errorf("解析日志数据失败: %w", err)
		}
	}

	// 提取日志消息并应用净化
	messages := c.extractMessages(&logData)
	totalItems := len(messages)

	// 应用关键词过滤（在截断前，用完整文本匹配）
	if c.keyword.IsActive() {
		messages = c.filterByKeyword(messages)
	}

	// 截断超长消息（在关键词过滤之后，避免截断导致关键词丢失）
	c.truncateMessages(messages)

	// 构造结果
	result := &QueryResult{
		LogID:         logid,
		Region:        string(c.regionConfig.Region),
		RegionDisplay: c.regionConfig.Region.DisplayName(),
		TotalItems:    totalItems,
		Messages:      messages,
		Meta:          logData.Meta,
		TagInfos:      logData.TagInfos,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}

	if c.keyword.IsActive() {
		result.FilteredItems = len(messages)
	}

	return result, nil
}

// extractMessages 从 API 响应中提取并净化日志消息。
func (c *Client) extractMessages(data *LogData) []ExtractedLogMessage {
	var messages []ExtractedLogMessage

	for _, item := range data.Items {
		for _, val := range item.Value {
			var values []ExtractedValue
			var location string

			for _, kv := range val.KVList {
				switch kv.Key {
				case "_msg":
					sanitized := c.sanitizer.Sanitize(kv.Value)
					values = append(values, ExtractedValue{
						Key:   kv.Key,
						Value: sanitized,
					})
				case "_location":
					location = kv.Value
				}
			}

			if len(values) > 0 {
				messages = append(messages, ExtractedLogMessage{
					ID:       fmt.Sprintf("%s-%s", item.ID, val.ID),
					Group:    item.Group,
					Values:   values,
					Location: location,
					Level:    val.Level,
				})
			}
		}
	}

	return messages
}

// truncateMessages 对所有消息的 value 进行截断。
func (c *Client) truncateMessages(messages []ExtractedLogMessage) {
	if c.maxLen <= 0 {
		return
	}
	for i := range messages {
		for j := range messages[i].Values {
			v := messages[i].Values[j].Value
			if len(v) > c.maxLen {
				messages[i].Values[j].Value = fmt.Sprintf(
					"%s...[truncated, original: %d chars]", v[:c.maxLen], len(v))
			}
		}
	}
}

// filterByKeyword 使用关键词过滤器筛选日志消息。
func (c *Client) filterByKeyword(messages []ExtractedLogMessage) []ExtractedLogMessage {
	var filtered []ExtractedLogMessage
	for _, msg := range messages {
		for _, v := range msg.Values {
			if c.keyword.Matches(v.Value) {
				filtered = append(filtered, msg)
				break
			}
		}
	}
	return filtered
}
