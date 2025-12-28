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
			proc, err := NewProcessor(tt.patterns, []TagConfig{{Name: "要約", Limit: 0}}, "summary", false, false, "newest")
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
			name: "要約タグなし",
			description: `
これが最初の行です
2行目はこちら`,
			want: "",
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
			want: "",
		},
		{
			name: "終了タグのみ",
			description: `終了タグのみ[/要約]

本文`,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proc, _ := NewProcessor([]string{}, []TagConfig{{Name: "要約", Limit: 0}}, "summary", false, false, "newest")
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
			proc, _ := NewProcessor([]string{}, []TagConfig{{Name: "要約", Limit: 0}}, "summary", false, false, "newest")
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
	proc, err := NewProcessor(patterns, []TagConfig{{Name: "要約", Limit: 0}}, "summary", false, false, "newest")
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

	proc, err := NewProcessor(patterns, []TagConfig{{Name: "要約", Limit: 0}}, "summary", false, false, "newest")
	if err != nil {
		t.Fatalf("NewProcessor()でエラー: %v", err)
	}

	// 不正なパターンはスキップされるはず
	// 正常なパターンのみが適用される
	if len(proc.cleaningPatterns) != 1 {
		t.Errorf("cleaningPatterns length = %d; want 1 (不正なパターンはスキップされる)", len(proc.cleaningPatterns))
	}
}

func TestExtractTags_FromJournals(t *testing.T) {
	tests := []struct {
		name             string
		tagConfigs       []TagConfig
		description      string
		journals         []redmine.Journal
		includeComments  bool
		want             map[string][]string
	}{
		{
			name:        "ジャーナルからタグ抽出（includeComments=true）",
			tagConfigs:  []TagConfig{{Name: "進捗", Limit: 0}, {Name: "課題", Limit: 0}},
			description: "説明文",
			journals: []redmine.Journal{
				{
					ID:    1,
					Notes: "[進捗]バグが発生しました[/進捗]",
				},
				{
					ID:    2,
					Notes: "[課題]修正が必要です[/課題]",
				},
			},
			includeComments: true,
			want: map[string][]string{
				"進捗": {"バグが発生しました"},
				"課題": {"修正が必要です"},
			},
		},
		{
			name:        "ジャーナルからタグ抽出（includeComments=false）",
			tagConfigs:  []TagConfig{{Name: "進捗", Limit: 0}, {Name: "課題", Limit: 0}},
			description: "説明文",
			journals: []redmine.Journal{
				{
					ID:    1,
					Notes: "[進捗]バグが発生しました[/進捗]",
				},
			},
			includeComments: false,
			want:            map[string][]string{}, // includeComments=falseなので抽出されない
		},
		{
			name:        "説明文とジャーナル両方にタグ（両方抽出）",
			tagConfigs:  []TagConfig{{Name: "進捗", Limit: 0}},
			description: "[進捗]説明文の進捗[/進捗]",
			journals: []redmine.Journal{
				{
					ID:    1,
					Notes: "[進捗]コメントの進捗[/進捗]",
				},
			},
			includeComments: true,
			want: map[string][]string{
				"進捗": {"説明文の進捗", "コメントの進捗"}, // 両方抽出される
			},
		},
		{
			name:        "空のジャーナル",
			tagConfigs:  []TagConfig{{Name: "進捗", Limit: 0}},
			description: "説明文",
			journals: []redmine.Journal{
				{
					ID:    1,
					Notes: "",
				},
			},
			includeComments: true,
			want:            map[string][]string{},
		},
		{
			name:        "複数のジャーナル（すべて抽出）",
			tagConfigs:  []TagConfig{{Name: "進捗", Limit: 0}},
			description: "説明文",
			journals: []redmine.Journal{
				{
					ID:    1,
					Notes: "[進捗]最初のコメント[/進捗]",
				},
				{
					ID:    2,
					Notes: "[進捗]二番目のコメント[/進捗]",
				},
			},
			includeComments: true,
			want: map[string][]string{
				"進捗": {"二番目のコメント", "最初のコメント"}, // 最新から順に抽出
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proc, err := NewProcessor([]string{}, tt.tagConfigs, "tags", false, tt.includeComments, "newest")
			if err != nil {
				t.Fatalf("NewProcessor()でエラー: %v", err)
			}

			got := proc.ExtractTags(tt.description, tt.journals)

			if len(got) != len(tt.want) {
				t.Errorf("ExtractTags() 結果の数 = %d; want %d", len(got), len(tt.want))
			}

			for tagName, wantContents := range tt.want {
				gotContents, ok := got[tagName]
				if !ok {
					t.Errorf("ExtractTags() タグ %q が見つかりません", tagName)
					continue
				}
				if len(gotContents) != len(wantContents) {
					t.Errorf("ExtractTags() タグ %q の値の数 = %d; want %d", tagName, len(gotContents), len(wantContents))
					continue
				}
				for i := range wantContents {
					if gotContents[i] != wantContents[i] {
						t.Errorf("ExtractTags() タグ %q の値[%d] = %q; want %q", tagName, i, gotContents[i], wantContents[i])
					}
				}
			}
		})
	}
}

func TestExtractTags_WithLimit(t *testing.T) {
	tests := []struct {
		name        string
		tagConfigs  []TagConfig
		description string
		journals    []redmine.Journal
		want        map[string][]string
	}{
		{
			name: "件数制限あり（3件）",
			tagConfigs: []TagConfig{
				{Name: "進捗", Limit: 3},
			},
			description: "[進捗]説明文の進捗[/進捗]",
			journals: []redmine.Journal{
				{ID: 1, Notes: "[進捗]コメント1[/進捗]"},
				{ID: 2, Notes: "[進捗]コメント2[/進捗]"},
				{ID: 3, Notes: "[進捗]コメント3[/進捗]"},
				{ID: 4, Notes: "[進捗]コメント4[/進捗]"},
				{ID: 5, Notes: "[進捗]コメント5[/進捗]"},
			},
			want: map[string][]string{
				"進捗": {"説明文の進捗", "コメント5", "コメント4"}, // 説明文+コメント最新2件（合計3件）
			},
		},
		{
			name: "件数制限なし（0）",
			tagConfigs: []TagConfig{
				{Name: "進捗", Limit: 0},
			},
			description: "[進捗]説明文の進捗[/進捗]",
			journals: []redmine.Journal{
				{ID: 1, Notes: "[進捗]コメント1[/進捗]"},
				{ID: 2, Notes: "[進捗]コメント2[/進捗]"},
			},
			want: map[string][]string{
				"進捗": {"説明文の進捗", "コメント2", "コメント1"}, // すべて（コメントは新しい順）
			},
		},
		{
			name: "タグごとに異なる件数制限",
			tagConfigs: []TagConfig{
				{Name: "進捗", Limit: 2},
				{Name: "課題", Limit: 1},
			},
			description: "[進捗]説明文の進捗[/進捗]\n[課題]説明文の課題[/課題]",
			journals: []redmine.Journal{
				{ID: 1, Notes: "[進捗]進捗1[/進捗]\n[課題]課題1[/課題]"},
				{ID: 2, Notes: "[進捗]進捗2[/進捗]\n[課題]課題2[/課題]"},
			},
			want: map[string][]string{
				"進捗": {"説明文の進捗", "進捗2"}, // 説明文+コメント最新1件（合計2件）
				"課題": {"説明文の課題"},         // 説明文のみ（合計1件）
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proc, err := NewProcessor([]string{}, tt.tagConfigs, "tags", false, true, "newest")
			if err != nil {
				t.Fatalf("NewProcessor()でエラー: %v", err)
			}

			got := proc.ExtractTags(tt.description, tt.journals)

			if len(got) != len(tt.want) {
				t.Errorf("ExtractTags() 結果の数 = %d; want %d", len(got), len(tt.want))
			}

			for tagName, wantValues := range tt.want {
				gotValues, ok := got[tagName]
				if !ok {
					t.Errorf("ExtractTags() タグ %s が見つかりません", tagName)
					continue
				}

				if len(gotValues) != len(wantValues) {
					t.Errorf("ExtractTags() タグ %s の値の数 = %d; want %d", tagName, len(gotValues), len(wantValues))
					t.Errorf("  got:  %v", gotValues)
					t.Errorf("  want: %v", wantValues)
					continue
				}

				for i, wantValue := range wantValues {
					if gotValues[i] != wantValue {
						t.Errorf("ExtractTags() タグ %s の値[%d] = %q; want %q", tagName, i, gotValues[i], wantValue)
					}
				}
			}
		})
	}
}

func TestProcess_WithJournalTags(t *testing.T) {
	// テストデータ準備
	issue1 := &redmine.Issue{
		ID:          1,
		Subject:     "タスク1",
		Description: "説明文",
		Journals: []redmine.Journal{
			{
				ID:    1,
				Notes: "[進捗]バグが発生しました[/進捗]\n[課題]修正が必要です[/課題]",
			},
		},
	}

	issues := []*redmine.Issue{issue1}

	// プロセッサー作成（includeComments=true）
	proc, err := NewProcessor([]string{}, []TagConfig{{Name: "進捗", Limit: 0}, {Name: "課題", Limit: 0}, {Name: "要約", Limit: 0}}, "tags", false, true, "newest")
	if err != nil {
		t.Fatalf("NewProcessor()でエラー: %v", err)
	}

	// 処理実行
	roots := proc.Process(issues)

	// 検証
	if len(roots) != 1 {
		t.Fatalf("roots length = %d; want 1", len(roots))
	}

	// ExtractedTagsの検証
	if roots[0].ExtractedTags == nil {
		t.Fatal("ExtractedTags is nil")
	}

	if got, ok := roots[0].ExtractedTags["進捗"]; !ok {
		t.Error("タグ '進捗' が見つかりません")
	} else if len(got) != 1 || got[0] != "バグが発生しました" {
		t.Errorf("タグ '進捗' の内容 = %q; want ['バグが発生しました']", got)
	}

	if got, ok := roots[0].ExtractedTags["課題"]; !ok {
		t.Error("タグ '課題' が見つかりません")
	} else if len(got) != 1 || got[0] != "修正が必要です" {
		t.Errorf("タグ '課題' の内容 = %q; want ['修正が必要です']", got)
	}
}
