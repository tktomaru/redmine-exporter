package config

import (
	"fmt"

	"gopkg.in/ini.v1"
)

// Config はアプリケーション設定を保持
type Config struct {
	Redmine       RedmineConfig
	TitleCleaning TitleCleaningConfig
}

// RedmineConfig はRedmine接続設定
type RedmineConfig struct {
	BaseURL   string
	APIKey    string
	FilterURL string
}

// TitleCleaningConfig はタイトルクリーニング設定
type TitleCleaningConfig struct {
	Patterns []string
}

// LoadConfig は指定されたパスから設定ファイルを読み込む
func LoadConfig(path string) (*Config, error) {
	cfg, err := ini.Load(path)
	if err != nil {
		return nil, fmt.Errorf("設定ファイルの読み込みに失敗: %w", err)
	}

	config := &Config{}

	// [Redmine]セクション
	config.Redmine.BaseURL = cfg.Section("Redmine").Key("BaseUrl").String()
	config.Redmine.APIKey = cfg.Section("Redmine").Key("ApiKey").String()
	config.Redmine.FilterURL = cfg.Section("Redmine").Key("FilterUrl").String()

	// [TitleCleaning]セクション - Pattern1, Pattern2, ... を動的に読み込む
	section := cfg.Section("TitleCleaning")
	patterns := []string{}
	for i := 1; ; i++ {
		key := fmt.Sprintf("Pattern%d", i)
		if !section.HasKey(key) {
			break
		}
		pattern := section.Key(key).String()
		if pattern != "" {
			patterns = append(patterns, pattern)
		}
	}
	config.TitleCleaning.Patterns = patterns

	// バリデーション
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate は設定値の妥当性をチェック
func (c *Config) Validate() error {
	if c.Redmine.BaseURL == "" {
		return fmt.Errorf("BaseUrlが設定されていません")
	}
	if c.Redmine.APIKey == "" {
		return fmt.Errorf("ApiKeyが設定されていません")
	}
	if c.Redmine.FilterURL == "" {
		return fmt.Errorf("FilterUrlが設定されていません")
	}
	return nil
}
