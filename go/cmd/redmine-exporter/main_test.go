package main

import (
	"testing"
)

func TestParseTags(t *testing.T) {
	tests := []struct {
		name        string
		tagsStr     string
		commentsMax int
		wantLimits  map[string]int
		wantErr     bool
	}{
		{
			name:        "commentsが上限（すべてのタグ）",
			tagsStr:     "要約,進捗,課題",
			commentsMax: 3,
			wantLimits:  map[string]int{"要約": 3, "進捗": 3, "課題": 3},
		},
		{
			name:        "個別指定とcommentsの小さい方",
			tagsStr:     "要約:5,進捗:2,課題",
			commentsMax: 3,
			wantLimits:  map[string]int{"要約": 3, "進捗": 2, "課題": 3},
		},
		{
			name:        "commentsなし、個別指定のみ",
			tagsStr:     "要約:5,進捗:2,課題",
			commentsMax: 0,
			wantLimits:  map[string]int{"要約": 5, "進捗": 2, "課題": 0},
		},
		{
			name:        "commentsなし、個別指定なし",
			tagsStr:     "要約,進捗,課題",
			commentsMax: 0,
			wantLimits:  map[string]int{"要約": 0, "進捗": 0, "課題": 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configs, _, err := parseTags(tt.tagsStr, tt.commentsMax)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				for _, cfg := range configs {
					want, ok := tt.wantLimits[cfg.Name]
					if !ok {
						t.Errorf("予期しないタグ: %s", cfg.Name)
						continue
					}
					if cfg.Limit != want {
						t.Errorf("タグ %s のLimit = %d; want %d", cfg.Name, cfg.Limit, want)
					}
				}
			}
		})
	}
}

func TestParseCommentsLimit(t *testing.T) {
	tests := []struct {
		name         string
		commentsMode string
		want         int
		wantErr      bool
	}{
		{name: "n:3", commentsMode: "n:3", want: 3, wantErr: false},
		{name: "n:5", commentsMode: "n:5", want: 5, wantErr: false},
		{name: "last", commentsMode: "last", want: 0, wantErr: false},
		{name: "all", commentsMode: "all", want: 0, wantErr: false},
		{name: "empty", commentsMode: "", want: 0, wantErr: false},
		{name: "invalid", commentsMode: "invalid", want: 0, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCommentsLimit(tt.commentsMode)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCommentsLimit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseCommentsLimit() = %v, want %v", got, tt.want)
			}
		})
	}
}
