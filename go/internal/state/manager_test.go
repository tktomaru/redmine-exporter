package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	mgr := NewManager("/tmp/test.state")
	if mgr == nil {
		t.Fatal("NewManager() returned nil")
	}
	if mgr.filePath != "/tmp/test.state" {
		t.Errorf("filePath = %s, want /tmp/test.state", mgr.filePath)
	}
}

func TestManager_SaveAndLoad(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "test.state")

	mgr := NewManager(stateFile)

	// Stateを作成
	state := &State{
		LastRun:        time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
		LastSuccessRun: time.Date(2025, 1, 15, 9, 0, 0, 0, time.UTC),
		Version:        "1.0.0",
		FilterConfig: map[string]string{
			"week":       "last",
			"date_field": "updated_on",
		},
	}

	// 保存
	if err := mgr.Save(state); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// ファイルが存在することを確認
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Fatal("State file was not created")
	}

	// 読み込み
	loaded, err := mgr.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// 検証
	if !loaded.LastRun.Equal(state.LastRun) {
		t.Errorf("LastRun = %v, want %v", loaded.LastRun, state.LastRun)
	}
	if !loaded.LastSuccessRun.Equal(state.LastSuccessRun) {
		t.Errorf("LastSuccessRun = %v, want %v", loaded.LastSuccessRun, state.LastSuccessRun)
	}
	if loaded.Version != state.Version {
		t.Errorf("Version = %s, want %s", loaded.Version, state.Version)
	}
	if loaded.FilterConfig["week"] != "last" {
		t.Errorf("FilterConfig[week] = %s, want last", loaded.FilterConfig["week"])
	}
	if loaded.FilterConfig["date_field"] != "updated_on" {
		t.Errorf("FilterConfig[date_field] = %s, want updated_on", loaded.FilterConfig["date_field"])
	}
}

func TestManager_Load_FileNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "nonexistent.state")

	mgr := NewManager(stateFile)

	// ファイルが存在しない場合、空のStateを返す
	state, err := mgr.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if state == nil {
		t.Fatal("Load() returned nil state")
	}

	if state.FilterConfig == nil {
		t.Error("FilterConfig should be initialized")
	}

	if !state.LastRun.IsZero() {
		t.Error("LastRun should be zero")
	}
}

func TestManager_Load_CorruptedFile(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "corrupted.state")

	// 不正なJSONファイルを作成
	os.WriteFile(stateFile, []byte("invalid json content"), 0644)

	mgr := NewManager(stateFile)

	// 破損したファイルの場合、警告を返して空のStateを返す
	state, err := mgr.Load()
	if err == nil {
		t.Error("Load() should return error for corrupted file")
	}

	if state == nil {
		t.Fatal("Load() should return empty state even if corrupted")
	}

	if state.FilterConfig == nil {
		t.Error("FilterConfig should be initialized")
	}
}

func TestManager_UpdateLastRun(t *testing.T) {
	mgr := NewManager("/tmp/test.state")
	state := &State{}

	before := time.Now()
	mgr.UpdateLastRun(state)
	after := time.Now()

	if state.LastRun.Before(before) || state.LastRun.After(after) {
		t.Errorf("LastRun = %v, should be between %v and %v", state.LastRun, before, after)
	}
}

func TestManager_UpdateLastSuccessRun(t *testing.T) {
	mgr := NewManager("/tmp/test.state")
	state := &State{}

	before := time.Now()
	mgr.UpdateLastSuccessRun(state)
	after := time.Now()

	if state.LastSuccessRun.Before(before) || state.LastSuccessRun.After(after) {
		t.Errorf("LastSuccessRun = %v, should be between %v and %v", state.LastSuccessRun, before, after)
	}
}

func TestManager_SetAndGetFilterConfig(t *testing.T) {
	mgr := NewManager("/tmp/test.state")
	state := &State{}

	// 設定
	mgr.SetFilterConfig(state, "test_key", "test_value")

	// 取得
	value := mgr.GetFilterConfig(state, "test_key")
	if value != "test_value" {
		t.Errorf("GetFilterConfig() = %s, want test_value", value)
	}

	// 存在しないキー
	value = mgr.GetFilterConfig(state, "nonexistent")
	if value != "" {
		t.Errorf("GetFilterConfig() for nonexistent key = %s, want empty", value)
	}
}

func TestManager_SetFilterConfig_NilMap(t *testing.T) {
	mgr := NewManager("/tmp/test.state")
	state := &State{
		FilterConfig: nil,
	}

	// nilマップでも正常に動作すること
	mgr.SetFilterConfig(state, "key", "value")

	if state.FilterConfig == nil {
		t.Fatal("FilterConfig should be initialized")
	}

	if state.FilterConfig["key"] != "value" {
		t.Errorf("FilterConfig[key] = %s, want value", state.FilterConfig["key"])
	}
}

func TestManager_Save_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "atomic.state")

	mgr := NewManager(stateFile)
	state := &State{
		Version: "1.0.0",
		FilterConfig: make(map[string]string),
	}

	// 保存
	if err := mgr.Save(state); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// 一時ファイルが残っていないことを確認
	tmpFile := stateFile + ".tmp"
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Error("Temporary file should be removed after save")
	}

	// 本番ファイルが存在することを確認
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Error("State file should exist after save")
	}
}
