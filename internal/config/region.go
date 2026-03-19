// Package config 提供 logid 的配置管理功能。
package config

import "fmt"

// Region 表示支持的区域标识。
type Region string

const (
	RegionCN   Region = "cn"
	RegionI18n Region = "i18n"
	RegionUS   Region = "us"
	RegionEU   Region = "eu"
)

// AllRegions 返回所有支持的区域列表。
func AllRegions() []Region {
	return []Region{RegionUS, RegionI18n, RegionEU, RegionCN}
}

// ParseRegion 将字符串解析为 Region，不区分大小写。
// 返回错误如果字符串不是有效的区域标识。
func ParseRegion(s string) (Region, error) {
	switch s {
	case "cn", "CN":
		return RegionCN, nil
	case "i18n", "I18N":
		return RegionI18n, nil
	case "us", "US":
		return RegionUS, nil
	case "eu", "EU":
		return RegionEU, nil
	default:
		return "", fmt.Errorf("不支持的区域: %s，支持的区域: us, i18n, eu, cn", s)
	}
}

// String 返回区域的字符串表示。
func (r Region) String() string {
	return string(r)
}

// DisplayName 返回区域的中文显示名称。
func (r Region) DisplayName() string {
	switch r {
	case RegionCN:
		return "中国区"
	case RegionI18n:
		return "国际化区域（新加坡）"
	case RegionUS:
		return "美区"
	case RegionEU:
		return "欧洲区"
	default:
		return string(r)
	}
}

// RegionConfig 包含区域相关的服务配置。
type RegionConfig struct {
	// Region 区域标识
	Region Region
	// LogServiceURL 日志服务 API 地址
	LogServiceURL string
	// VRegion 虚拟区域列表（逗号分隔）
	VRegion string
	// Configured 标记该区域是否已配置
	Configured bool
}

// GetRegionConfig 根据区域标识获取对应的服务配置。
func GetRegionConfig(region Region) *RegionConfig {
	switch region {
	case RegionUS:
		return &RegionConfig{
			Region:        RegionUS,
			LogServiceURL: "https://logservice-tx.tiktok-us.org/streamlog/platform/microservice/v1/query/trace",
			VRegion:       "US-TTP,US-TTP2",
			Configured:    true,
		}
	case RegionI18n:
		return &RegionConfig{
			Region:        RegionI18n,
			LogServiceURL: "https://logservice-sg.tiktok-row.org/streamlog/platform/microservice/v1/query/trace",
			VRegion:       "Singapore-Common,US-East,Singapore-Central",
			Configured:    true,
		}
	case RegionEU:
		return &RegionConfig{
			Region:        RegionEU,
			LogServiceURL: "https://logservice-eu-ttp.tiktok-eu.org/streamlog/platform/microservice/v1/query/trace",
			VRegion:       "US-EastRed,EU-TTP2,EU-TTP-PPE,EU-TTP",
			Configured:    true,
		}
	case RegionCN:
		return &RegionConfig{
			Region:     RegionCN,
			Configured: false,
		}
	default:
		return nil
	}
}
