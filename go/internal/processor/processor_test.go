package processor

import (
	"testing"

	"github.com/tktomaru/redmine-exporter/internal/redmine"
)

func TestCleanTitle(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		subject  string
		want     string
	}{
		{
			name:     "プレフィックス削除",
			patterns: []string{`^\[.*?\]\s*`},
			subject:  "[WIP] ログイン機能の実装",
			want:     "ログイン機能の実装",
		},
		{
			name:     "サフィックス削除",
			patterns: []string{`\s*\(.*?\)$`},
			subject:  "バグ修正 (完了予定)",
			want:     "バグ修正",
		},
		{
			name:     "複数パターン適用",
			patterns: []string{`^\[.*?\]\s*`, `\s*\(.*?\)$`},
			subject:  "[完了] セキュリティ対応 (重要)",
			want:     "セキュリティ対応",
		},
		{
			name:     "パターンなし",
			patterns: []string{},
			subject:  "通常のタイトル",
			want:     "通常のタイトル",
		},
		{
			name:     "マッチしないパターン",
			patterns: []string{`^\[.*?\]\s*`},
			subject:  "プレフィックスなしタイトル",
			want:     "プレフィックスなしタイトル",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proc, err := NewProcessor(tt.patterns, []string{"要約"}, "summary")
			if err != nil {
				t.Fatalf("NewProcessor()でエラー: %v", err)
			}

			got := proc.CleanTitle(tt.subject)
			if got != tt.want {
				t.Errorf("CleanTitle() = %q; want %q", got, tt.want)
			}
		})
	}
}

func TestExtractSummary(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        string
	}{
		{
			name: "要約タグあり",
			description: `[要約]ログイン機能を実装しました[/要約]

詳細:
- 認証処理の追加
- セッション管理`,
			want: "ログイン機能を実装しました",
		},
		{
			name: "要約タグなし（最初の非空行）",
			description: `
これが最初の行です
2行目はこちら`,
			want: "これが最初の行です",
		},
		{
			name: "要約タグあり（空白付き）",
			description: `[要約]  要約にスペースがあります  [/要約]

本文`,
			want: "要約にスペースがあります",
		},
		{
			name:        "空の説明",
			description: "",
			want:        "",
		},
		{
			name: "開始タグのみ",
			description: `[要約]タグが閉じていない

本文`,
			want: "[要約]タグが閉じていない",
		},
		{
			name: "終了タグのみ",
			description: `終了タグのみ[/要約]

本文`,
			want: "終了タグのみ[/要約]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proc, _ := NewProcessor([]string{}, []string{"要約"}, "summary")
			got := proc.ExtractSummary(tt.description)
			if got != tt.want {
				t.Errorf("ExtractSummary() = %q; want %q", got, tt.want)
			}
		})
	}
}

func TestFirstLine(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "単一行",
			input: "最初の行",
			want:  "最初の行",
		},
		{
			name:  "複数行",
			input: "最初の行\n2行目\n3行目",
			want:  "最初の行",
		},
		{
			name:  "先頭に空行",
			input: "\n\n最初の非空行",
			want:  "最初の非空行",
		},
		{
			name:  "空文字列",
			input: "",
			want:  "",
		},
		{
			name:  "空行のみ",
			input: "\n\n\n",
			want:  "",
		},
		{
			name:  "Windows改行",
			input: "最初の行\r\n2行目",
			want:  "最初の行",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proc, _ := NewProcessor([]string{}, []string{"要約"}, "summary")
			got := proc.firstLine(tt.input)
			if got != tt.want {
				t.Errorf("firstLine() = %q; want %q", got, tt.want)
			}
		})
	}
}

func TestGetAssignee(t *testing.T) {
	tests := []struct {
		name  string
		issue *redmine.Issue
		want  string
	}{
		{
			name: "担当者あり",
			issue: &redmine.Issue{
				AssignedTo: &redmine.IDName{
					ID:   1,
					Name: "佐藤",
				},
			},
			want: "佐藤",
		},
		{
			name: "担当者なし",
			issue: &redmine.Issue{
				AssignedTo: nil,
			},
			want: "担当者未定",
		},
		{
			name: "担当者名が空",
			issue: &redmine.Issue{
				AssignedTo: &redmine.IDName{
					ID:   1,
					Name: "",
				},
			},
			want: "担当者未定",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAssignee(tt.issue)
			if got != tt.want {
				t.Errorf("GetAssignee() = %q; want %q", got, tt.want)
			}
		})
	}
}

func TestProcess(t *testing.T) {
	// テストデータ準備
	parent1 := &redmine.Issue{
		ID:      1,
		Subject: "[親] 親タスクA",
	}

	child1 := &redmine.Issue{
		ID:      2,
		Subject: "[WIP] 子タスクB",
		Parent:  &redmine.IssueRef{ID: 1},
	}

	child2 := &redmine.Issue{
		ID:      3,
		Subject: "子タスクC (完了予定)",
		Parent:  &redmine.IssueRef{ID: 1},
	}

	parent2 := &redmine.Issue{
		ID:      4,
		Subject: "親タスクB",
	}

	issues := []*redmine.Issue{parent1, child1, child2, parent2}

	// プロセッサー作成
	patterns := []string{`^\[.*?\]\s*`, `\s*\(.*?\)$`}
	proc, err := NewProcessor(patterns, []string{"要約"}, "summary")
	if err != nil {
		t.Fatalf("NewProcessor()でエラー: %v", err)
	}

	// 処理実行
	roots := proc.Process(issues)

	// 検証
	if len(roots) != 2 {
		t.Fatalf("roots length = %d; want 2", len(roots))
	}

	// parent1の検証
	if roots[0].CleanedSubject != "親タスクA" {
		t.Errorf("roots[0].CleanedSubject = %q; want '親タスクA'", roots[0].CleanedSubject)
	}

	if len(roots[0].Children) != 2 {
		t.Fatalf("parent1の子タスク数 = %d; want 2", len(roots[0].Children))
	}

	if roots[0].Children[0].CleanedSubject != "子タスクB" {
		t.Errorf("child1.CleanedSubject = %q; want '子タスクB'", roots[0].Children[0].CleanedSubject)
	}

	if roots[0].Children[1].CleanedSubject != "子タスクC" {
		t.Errorf("child2.CleanedSubject = %q; want '子タスクC'", roots[0].Children[1].CleanedSubject)
	}

	// parent2の検証
	if roots[1].CleanedSubject != "親タスクB" {
		t.Errorf("roots[1].CleanedSubject = %q; want '親タスクB'", roots[1].CleanedSubject)
	}

	if len(roots[1].Children) != 0 {
		t.Errorf("parent2の子タスク数 = %d; want 0", len(roots[1].Children))
	}
}

func TestNewProcessorWithInvalidPattern(t *testing.T) {
	// 不正な正規表現パターン
	patterns := []string{`[未閉じ`, `正常なパターン`, `(未閉じ`}

	proc, err := NewProcessor(patterns, []string{"要約"}, "summary")
	if err != nil {
		t.Fatalf("NewProcessor()でエラー: %v", err)
	}

	// 不正なパターンはスキップされるはず
	// 正常なパターンのみが適用される
	if len(proc.cleaningPatterns) != 1 {
		t.Errorf("cleaningPatterns length = %d; want 1 (不正なパターンはスキップされる)", len(proc.cleaningPatterns))
	}
}
