package config

import "testing"

func TestParseRegion(t *testing.T) {
	tests := []struct {
		input   string
		want    Region
		wantErr bool
	}{
		{"us", RegionUS, false},
		{"US", RegionUS, false},
		{"i18n", RegionI18n, false},
		{"I18N", RegionI18n, false},
		{"eu", RegionEU, false},
		{"cn", RegionCN, false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseRegion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRegion(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseRegion(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestRegionDisplayName(t *testing.T) {
	tests := []struct {
		region Region
		want   string
	}{
		{RegionUS, "美区"},
		{RegionI18n, "国际化区域（新加坡）"},
		{RegionEU, "欧洲区"},
		{RegionCN, "中国区"},
	}

	for _, tt := range tests {
		t.Run(string(tt.region), func(t *testing.T) {
			if got := tt.region.DisplayName(); got != tt.want {
				t.Errorf("Region(%q).DisplayName() = %q, want %q", tt.region, got, tt.want)
			}
		})
	}
}

func TestGetRegionConfig(t *testing.T) {
	// US 区域应该已配置
	us := GetRegionConfig(RegionUS)
	if us == nil || !us.Configured {
		t.Error("US region should be configured")
	}
	if us.LogServiceURL == "" {
		t.Error("US region should have a log service URL")
	}

	// CN 区域应该未配置
	cn := GetRegionConfig(RegionCN)
	if cn == nil || cn.Configured {
		t.Error("CN region should not be configured")
	}
}
