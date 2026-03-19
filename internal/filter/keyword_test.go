package filter

import "testing"

func TestKeywordFilter_NotActive(t *testing.T) {
	f := NewKeywordFilter(nil)
	if f.IsActive() {
		t.Error("empty keyword filter should not be active")
	}
	// 不活跃时应匹配所有文本
	if !f.Matches("anything") {
		t.Error("inactive filter should match everything")
	}
}

func TestKeywordFilter_CaseInsensitive(t *testing.T) {
	f := NewKeywordFilter([]string{"ForLive"})
	if !f.IsActive() {
		t.Error("filter with keywords should be active")
	}

	tests := []struct {
		text string
		want bool
	}{
		{"this has ForLive in it", true},
		{"this has forlive lowercase", true},
		{"this has FORLIVE uppercase", true},
		{"no match here", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := f.Matches(tt.text); got != tt.want {
			t.Errorf("Matches(%q) = %v, want %v", tt.text, got, tt.want)
		}
	}
}

func TestKeywordFilter_ORLogic(t *testing.T) {
	f := NewKeywordFilter([]string{"error", "warning"})

	tests := []struct {
		text string
		want bool
	}{
		{"this is an error message", true},
		{"this is a warning message", true},
		{"this has error and warning", true},
		{"this is normal", false},
	}

	for _, tt := range tests {
		if got := f.Matches(tt.text); got != tt.want {
			t.Errorf("Matches(%q) = %v, want %v", tt.text, got, tt.want)
		}
	}
}
