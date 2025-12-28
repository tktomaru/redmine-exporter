package state

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// State は実行状態を保存する構造体
type State struct {
	LastRun        time.Time         `json:"last_run"`         // 最後の実行日時
	LastSuccessRun time.Time         `json:"last_success_run"` // 最後の成功実行日時
	Version        string            `json:"version"`          // アプリケーションバージョン
	FilterConfig   map[string]string `json:"filter_config"`    // フィルタ設定のスナップショット
}

// Manager はStateの管理を行う
type Manager struct {
	filePath string
}

// NewManager は新しいManagerを作成
func NewManager(filePath string) *Manager {
	return &Manager{
		filePath: filePath,
	}
}

// Load はStateファイルを読み込む
// ファイルが存在しない場合は空のStateを返す
func (m *Manager) Load() (*State, error) {
	// ファイルが存在しない場合は空のStateを返す
	if _, err := os.Stat(m.filePath); os.IsNotExist(err) {
		return &State{
			FilterConfig: make(map[string]string),
		}, nil
	}

	// ファイルを読み込む
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return nil, fmt.Errorf("Stateファイル読み込みエラー: %w", err)
	}

	// JSONをパース
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		// JSONパースエラーの場合、警告を返して空のStateを返す
		return &State{
			FilterConfig: make(map[string]string),
		}, fmt.Errorf("State破損（新規作成します）: %w", err)
	}

	// FilterConfigがnilの場合は初期化
	if state.FilterConfig == nil {
		state.FilterConfig = make(map[string]string)
	}

	return &state, nil
}

// Save はStateをファイルに保存
func (m *Manager) Save(state *State) error {
	// JSONに変換（インデント付き）
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("State JSONエラー: %w", err)
	}

	// ファイルに書き込み（一時ファイル経由で安全に保存）
	tmpFile := m.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("State保存エラー: %w", err)
	}

	// 一時ファイルを本番ファイルにリネーム（アトミック操作）
	if err := os.Rename(tmpFile, m.filePath); err != nil {
		os.Remove(tmpFile) // クリーンアップ
		return fmt.Errorf("Stateファイル更新エラー: %w", err)
	}

	return nil
}

// UpdateLastRun は最後の実行日時を更新
func (m *Manager) UpdateLastRun(state *State) {
	state.LastRun = time.Now()
}

// UpdateLastSuccessRun は最後の成功実行日時を更新
func (m *Manager) UpdateLastSuccessRun(state *State) {
	state.LastSuccessRun = time.Now()
}

// SetFilterConfig はフィルタ設定を保存
func (m *Manager) SetFilterConfig(state *State, key, value string) {
	if state.FilterConfig == nil {
		state.FilterConfig = make(map[string]string)
	}
	state.FilterConfig[key] = value
}

// GetFilterConfig はフィルタ設定を取得
func (m *Manager) GetFilterConfig(state *State, key string) string {
	if state.FilterConfig == nil {
		return ""
	}
	return state.FilterConfig[key]
}
