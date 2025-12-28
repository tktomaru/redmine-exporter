package config

import (
	"fmt"
	"strings"

	"gopkg.in/ini.v1"
)

// Config はアプリケーション設定を保持
type Config struct {
	Redmine       RedmineConfig
	TitleCleaning TitleCleaningConfig
	Output        OutputConfig
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

// OutputConfig は出力設定
type OutputConfig struct {
	Mode            string   // summary, full, tags
	TagNames        []string // 抽出するタグ名のリスト
	IncludeComments bool     // コメントからも抽出するか
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

	// [Output]セクション
	outputSection := cfg.Section("Output")
	config.Output.Mode = outputSection.Key("Mode").MustString("summary")

	// TagNames - カンマ区切りのリストを読み込む
	tagNamesStr := outputSection.Key("TagNames").MustString("要約")
	if tagNamesStr != "" {
		config.Output.TagNames = splitAndTrim(tagNamesStr, ",")
	} else {
		config.Output.TagNames = []string{"要約"}
	}

	config.Output.IncludeComments = outputSection.Key("IncludeComments").MustBool(false)

	// バリデーション
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// splitAndTrim はカンマ区切りの文字列を分割してトリムする
func splitAndTrim(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
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
