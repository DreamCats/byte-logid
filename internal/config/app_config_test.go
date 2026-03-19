package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAppConfigEnsureFilters(t *testing.T) {
	// 使用临时目录模拟
	tmpDir := t.TempDir()
	ac := &AppConfig{ConfigDir: tmpDir}

	// 首次调用应创建文件
	if err := ac.EnsureFilters(); err != nil {
		t.Fatalf("EnsureFilters() error: %v", err)
	}

	path := filepath.Join(tmpDir, "filters.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("EnsureFilters() should create filters.json")
	}

	// 加载并验证内容是默认规则
	fc, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(fc.MsgFilters) != len(DefaultFilters()) {
		t.Errorf("default filters count = %d, want %d", len(fc.MsgFilters), len(DefaultFilters()))
	}

	// 再次调用不应覆盖
	fc.AddFilter("custom")
	if err := fc.Save(path); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	if err := ac.EnsureFilters(); err != nil {
		t.Fatalf("EnsureFilters() second call error: %v", err)
	}

	fc2, _ := Load(path)
	if len(fc2.MsgFilters) != len(DefaultFilters())+1 {
		t.Error("EnsureFilters() should not overwrite existing file")
	}
}

func TestAppConfigFiltersPath(t *testing.T) {
	ac := &AppConfig{ConfigDir: "/tmp/test-logid"}
	want := filepath.Join("/tmp/test-logid", "filters.json")
	if got := ac.FiltersPath(); got != want {
		t.Errorf("FiltersPath() = %q, want %q", got, want)
	}
}
