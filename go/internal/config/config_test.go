package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.config")

	// テスト用の設定ファイルを作成
	configContent := `[Redmine]
BaseUrl=https://test.example.com
ApiKey=test_api_key_123
FilterUrl=/issues.json?project_id=1&status_id=*

[TitleCleaning]
Pattern1=^\[.*?\]\s*
Pattern2=\s*\(.*?\)$
Pattern3=【重要】
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("設定ファイルの作成に失敗: %v", err)
	}

	// 設定ファイルを読み込み
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig()でエラー: %v", err)
	}

	// Redmine設定の検証
	if cfg.Redmine.BaseURL != "https://test.example.com" {
		t.Errorf("BaseURL = %s; want https://test.example.com", cfg.Redmine.BaseURL)
	}

	if cfg.Redmine.APIKey != "test_api_key_123" {
		t.Errorf("APIKey = %s; want test_api_key_123", cfg.Redmine.APIKey)
	}

	if cfg.Redmine.FilterURL != "/issues.json?project_id=1&status_id=*" {
		t.Errorf("FilterURL = %s; want /issues.json?project_id=1&status_id=*", cfg.Redmine.FilterURL)
	}

	// TitleCleaning設定の検証
	expectedPatterns := []string{`^\[.*?\]\s*`, `\s*\(.*?\)$`, "【重要】"}
	if len(cfg.TitleCleaning.Patterns) != len(expectedPatterns) {
		t.Fatalf("Patterns length = %d; want %d", len(cfg.TitleCleaning.Patterns), len(expectedPatterns))
	}

	for i, pattern := range expectedPatterns {
		if cfg.TitleCleaning.Patterns[i] != pattern {
			t.Errorf("Pattern[%d] = %s; want %s", i, cfg.TitleCleaning.Patterns[i], pattern)
		}
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.ini")
	if err == nil {
		t.Error("存在しないファイルでエラーが発生しなかった")
	}
}

func TestLoadConfigMissingRequiredFields(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.config")

	// BaseUrlが欠けている不正な設定ファイル
	configContent := `[Redmine]
ApiKey=test_api_key
FilterUrl=/issues.json
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("設定ファイルの作成に失敗: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("BaseUrlが欠けている設定でエラーが発生しなかった")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "正常な設定",
			cfg: Config{
				Redmine: RedmineConfig{
					BaseURL:   "https://example.com",
					APIKey:    "key123",
					FilterURL: "/issues.json",
				},
			},
			wantErr: false,
		},
		{
			name: "BaseURLが空",
			cfg: Config{
				Redmine: RedmineConfig{
					BaseURL:   "",
					APIKey:    "key123",
					FilterURL: "/issues.json",
				},
			},
			wantErr: true,
		},
		{
			name: "APIKeyが空",
			cfg: Config{
				Redmine: RedmineConfig{
					BaseURL:   "https://example.com",
					APIKey:    "",
					FilterURL: "/issues.json",
				},
			},
			wantErr: true,
		},
		{
			name: "FilterURLが空",
			cfg: Config{
				Redmine: RedmineConfig{
					BaseURL:   "https://example.com",
					APIKey:    "key123",
					FilterURL: "",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
