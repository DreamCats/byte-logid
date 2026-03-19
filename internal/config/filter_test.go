package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultFilters(t *testing.T) {
	filters := DefaultFilters()
	if len(filters) != 8 {
		t.Errorf("DefaultFilters() returned %d filters, want 8", len(filters))
	}
}

func TestFilterConfigSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "filters.json")

	// 保存
	fc := &FilterConfig{MsgFilters: []string{"pattern1", "pattern2"}}
	if err := fc.Save(path); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// 检查文件权限
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("file permission = %o, want 0600", perm)
	}

	// 加载
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(loaded.MsgFilters) != 2 {
		t.Errorf("loaded %d filters, want 2", len(loaded.MsgFilters))
	}
	if loaded.MsgFilters[0] != "pattern1" {
		t.Errorf("loaded filter[0] = %q, want %q", loaded.MsgFilters[0], "pattern1")
	}
}

func TestFilterConfigAddFilter(t *testing.T) {
	fc := &FilterConfig{MsgFilters: []string{"a"}}
	fc.AddFilter("b")
	if len(fc.MsgFilters) != 2 {
		t.Errorf("AddFilter: len = %d, want 2", len(fc.MsgFilters))
	}
	if fc.MsgFilters[1] != "b" {
		t.Errorf("AddFilter: [1] = %q, want %q", fc.MsgFilters[1], "b")
	}
}

func TestFilterConfigRemoveFilter(t *testing.T) {
	fc := &FilterConfig{MsgFilters: []string{"a", "b", "c"}}

	// 正常删除
	removed, err := fc.RemoveFilter(1)
	if err != nil {
		t.Fatalf("RemoveFilter(1) error: %v", err)
	}
	if removed != "b" {
		t.Errorf("RemoveFilter(1) = %q, want %q", removed, "b")
	}
	if len(fc.MsgFilters) != 2 {
		t.Errorf("after remove: len = %d, want 2", len(fc.MsgFilters))
	}

	// 越界
	_, err = fc.RemoveFilter(10)
	if err == nil {
		t.Error("RemoveFilter(10) should return error for out of bounds")
	}

	// 负索引
	_, err = fc.RemoveFilter(-1)
	if err == nil {
		t.Error("RemoveFilter(-1) should return error")
	}
}

func TestFilterConfigReset(t *testing.T) {
	fc := &FilterConfig{MsgFilters: []string{"custom"}}
	fc.Reset()
	if len(fc.MsgFilters) != len(DefaultFilters()) {
		t.Errorf("Reset: len = %d, want %d", len(fc.MsgFilters), len(DefaultFilters()))
	}
}
