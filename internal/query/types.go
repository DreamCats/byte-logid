// Package query 提供日志查询客户端，负责与 StreamLog 日志服务交互。
package query

import "encoding/json"

// LogQueryRequest 日志查询请求体。
type LogQueryRequest struct {
	// LogID 日志 ID
	LogID string `json:"logid"`
	// PSMList PSM 服务名过滤列表
	PSMList []string `json:"psm_list,omitempty"`
	// ScanSpanInMin 扫描时间范围（分钟）
	ScanSpanInMin int `json:"scan_span_in_min"`
	// VRegion 虚拟区域
	VRegion string `json:"vregion"`
}

// LogData 日志服务返回的数据部分。
type LogData struct {
	Items    []LogItem        `json:"items"`
	Meta     *LogMeta         `json:"meta,omitempty"`
	TagInfos []json.RawMessage `json:"tag_infos,omitempty"`
}

// LogItem 单个日志项目。
type LogItem struct {
	ID    string     `json:"id"`
	Group LogGroup   `json:"group"`
	Value []LogValue `json:"value"`
}

// LogGroup 日志分组信息。
type LogGroup struct {
	PSM     string `json:"psm,omitempty"`
	PodName string `json:"pod_name,omitempty"`
	IPv4    string `json:"ipv4,omitempty"`
	Env     string `json:"env,omitempty"`
	VRegion string `json:"vregion,omitempty"`
	IDC     string `json:"idc,omitempty"`
}

// LogValue 日志值条目。
type LogValue struct {
	ID     string  `json:"id"`
	KVList []LogKV `json:"kv_list"`
	Level  string  `json:"level,omitempty"`
}

// LogKV 日志键值对。
type LogKV struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Type      string `json:"type,omitempty"`
	Highlight bool   `json:"highlight,omitempty"`
}

// LogMeta 日志元数据。
type LogMeta struct {
	ScanTimeRange []TimeRange `json:"scan_time_range,omitempty"`
	LevelList     []string    `json:"level_list,omitempty"`
}

// TimeRange 时间范围。
type TimeRange struct {
	Start int64 `json:"start,omitempty"`
	End   int64 `json:"end,omitempty"`
}

// ExtractedLogMessage 从 API 响应中提取的日志消息。
type ExtractedLogMessage struct {
	ID       string           `json:"id"`
	Group    LogGroup         `json:"group"`
	Values   []ExtractedValue `json:"values"`
	Location string           `json:"location,omitempty"`
	Level    string           `json:"level,omitempty"`
}

// ExtractedValue 提取并净化后的日志值。
type ExtractedValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// QueryResult 最终的查询结果，用于 JSON 输出。
type QueryResult struct {
	LogID            string                `json:"logid"`
	Region           string                `json:"region"`
	RegionDisplay    string                `json:"region_display_name"`
	TotalItems       int                   `json:"total_items"`
	FilteredItems    int                   `json:"filtered_items,omitempty"`
	Messages         []ExtractedLogMessage `json:"messages"`
	Meta             *LogMeta              `json:"meta,omitempty"`
	TagInfos         []json.RawMessage     `json:"tag_infos,omitempty"`
	Timestamp        string                `json:"timestamp"`
}
