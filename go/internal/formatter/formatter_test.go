package formatter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/tktomaru/redmine-exporter/internal/redmine"
)

func createTestData() []*redmine.Issue {
	// テスト用のチケットデータを作成
	startDate := &redmine.Date{Time: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)}
	dueDate := &redmine.Date{Time: time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)}

	parent := &redmine.Issue{
		ID:             1,
		Subject:        "親タスクA",
		CleanedSubject: "親タスクA",
	}

	child := &redmine.Issue{
		ID:             2,
		Subject:        "タスクB",
		CleanedSubject: "タスクB",
		Status:         redmine.IDName{ID: 1, Name: "進行中"},
		StartDate:      startDate,
		DueDate:        dueDate,
		AssignedTo:     &redmine.IDName{ID: 1, Name: "佐藤"},
		Summary:        "ひとこと整形で変更しました",
		Parent:         &redmine.IssueRef{ID: 1},
	}

	parent.Children = []*redmine.Issue{child}

	return []*redmine.Issue{parent}
}

func TestTextFormatter(t *testing.T) {
	formatter := &TextFormatter{}
	formatter.SetMode("summary", []string{"要約"})
	roots := createTestData()

	var buf bytes.Buffer
	err := formatter.Format(roots, &buf)
	if err != nil {
		t.Fatalf("Format()でエラー: %v", err)
	}

	output := buf.String()

	// 出力内容の検証
	if !strings.Contains(output, "■親タスクA") {
		t.Error("親タスクが正しく出力されていない")
	}

	if !strings.Contains(output, "・タスクB") {
		t.Error("子タスクが正しく出力されていない")
	}

	if !strings.Contains(output, "【進行中】") {
		t.Error("ステータスが正しく出力されていない")
	}

	if !strings.Contains(output, "2026/01/02-2025/12/31") {
		t.Error("日付が正しく出力されていない")
	}

	if !strings.Contains(output, "担当: 佐藤") {
		t.Error("担当者が正しく出力されていない")
	}

	if !strings.Contains(output, "⇒ひとこと整形で変更しました") {
		t.Error("要約が正しく出力されていない")
	}
}

func TestMarkdownFormatter(t *testing.T) {
	formatter := &MarkdownFormatter{}
	formatter.SetMode("summary", []string{"要約"})
	roots := createTestData()

	var buf bytes.Buffer
	err := formatter.Format(roots, &buf)
	if err != nil {
		t.Fatalf("Format()でエラー: %v", err)
	}

	output := buf.String()

	// 出力内容の検証
	if !strings.Contains(output, "# 親タスクA") {
		t.Error("親タスクが見出しとして出力されていない")
	}

	if !strings.Contains(output, "- **タスクB**") {
		t.Error("子タスクが箇条書きとして出力されていない")
	}

	if !strings.Contains(output, "[進行中]") {
		t.Error("ステータスが正しく出力されていない")
	}

	if !strings.Contains(output, "  > ひとこと整形で変更しました") {
		t.Error("要約が引用ブロックとして出力されていない")
	}
}

func TestDetectFormatter(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantType string
		wantErr  bool
	}{
		{
			name:     "Markdown形式",
			filename: "output.md",
			wantType: "*formatter.MarkdownFormatter",
			wantErr:  false,
		},
		{
			name:     "テキスト形式",
			filename: "output.txt",
			wantType: "*formatter.TextFormatter",
			wantErr:  false,
		},
		{
			name:     "Excel形式",
			filename: "output.xlsx",
			wantType: "*formatter.ExcelFormatter",
			wantErr:  false,
		},
		{
			name:     "未対応の拡張子",
			filename: "output.csv",
			wantType: "",
			wantErr:  true,
		},
		{
			name:     "拡張子なし",
			filename: "output",
			wantType: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter, err := DetectFormatter(tt.filename, "summary", []string{"要約"}, "")

			if (err != nil) != tt.wantErr {
				t.Errorf("DetectFormatter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				switch tt.wantType {
				case "*formatter.MarkdownFormatter":
					if _, ok := formatter.(*MarkdownFormatter); !ok {
						t.Errorf("DetectFormatter() type = %T; want %s", formatter, tt.wantType)
					}
				case "*formatter.TextFormatter":
					if _, ok := formatter.(*TextFormatter); !ok {
						t.Errorf("DetectFormatter() type = %T; want %s", formatter, tt.wantType)
					}
				case "*formatter.ExcelFormatter":
					if _, ok := formatter.(*ExcelFormatter); !ok {
						t.Errorf("DetectFormatter() type = %T; want %s", formatter, tt.wantType)
					}
				}
			}
		})
	}
}

func TestFormatDate(t *testing.T) {
	tests := []struct {
		name string
		date *redmine.Date
		want string
	}{
		{
			name: "通常の日付",
			date: &redmine.Date{Time: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)},
			want: "2026/01/02",
		},
		{
			name: "nil",
			date: nil,
			want: "----/--/--",
		},
		{
			name: "ゼロ値",
			date: &redmine.Date{},
			want: "----/--/--",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDate(tt.date)
			if got != tt.want {
				t.Errorf("formatDate() = %q; want %q", got, tt.want)
			}
		})
	}
}

func TestTextFormatterEmptySummary(t *testing.T) {
	formatter := &TextFormatter{}
	formatter.SetMode("summary", []string{"要約"})

	parent := &redmine.Issue{
		ID:             1,
		CleanedSubject: "親タスク",
	}

	child := &redmine.Issue{
		ID:             2,
		CleanedSubject: "子タスク",
		Status:         redmine.IDName{Name: "新規"},
		Summary:        "", // 要約なし
	}

	parent.Children = []*redmine.Issue{child}
	roots := []*redmine.Issue{parent}

	var buf bytes.Buffer
	err := formatter.Format(roots, &buf)
	if err != nil {
		t.Fatalf("Format()でエラー: %v", err)
	}

	output := buf.String()

	// 要約が空の場合、⇒が出力されないことを確認
	if strings.Contains(output, "⇒") {
		t.Error("要約が空の場合は⇒が出力されないはず")
	}
}
