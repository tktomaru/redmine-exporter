package formatter

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tktomaru/redmine-exporter/internal/redmine"
)

func TestNewTemplateFormatter(t *testing.T) {
	// 一時テンプレートファイルを作成
	tmpDir := t.TempDir()
	tmplFile := filepath.Join(tmpDir, "test.tmpl")

	tmplContent := `# Test Template
{{ range .Issues }}
- {{ .Subject }}
{{ end }}`

	if err := os.WriteFile(tmplFile, []byte(tmplContent), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	// TemplateFormatterを作成
	fmtr, err := NewTemplateFormatter(tmplFile)
	if err != nil {
		t.Fatalf("NewTemplateFormatter() error = %v", err)
	}

	if fmtr == nil {
		t.Fatal("NewTemplateFormatter() returned nil")
	}

	if fmtr.tmplPath != tmplFile {
		t.Errorf("tmplPath = %s, want %s", fmtr.tmplPath, tmplFile)
	}
}

func TestNewTemplateFormatter_FileNotExist(t *testing.T) {
	_, err := NewTemplateFormatter("/nonexistent/template.tmpl")
	if err == nil {
		t.Error("NewTemplateFormatter() should return error for nonexistent file")
	}
}

func TestTemplateFormatter_Format(t *testing.T) {
	// 一時テンプレートファイルを作成
	tmpDir := t.TempDir()
	tmplFile := filepath.Join(tmpDir, "test.tmpl")

	tmplContent := `# Test Report
{{ range .Issues }}
## {{ .CleanedSubject }}
- ID: #{{ .ID }}
- Status: {{ status . }}
- Assignee: {{ assignee . }}
{{ end }}`

	if err := os.WriteFile(tmplFile, []byte(tmplContent), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	// TemplateFormatterを作成
	fmtr, err := NewTemplateFormatter(tmplFile)
	if err != nil {
		t.Fatalf("NewTemplateFormatter() error = %v", err)
	}

	// テストデータ
	issues := []*redmine.Issue{
		{
			ID:             1,
			Subject:        "Test Issue 1",
			CleanedSubject: "Test Issue 1",
			Status:         redmine.IDName{Name: "進行中"},
			AssignedTo:     &redmine.IDName{Name: "山田太郎"},
		},
		{
			ID:             2,
			Subject:        "Test Issue 2",
			CleanedSubject: "Test Issue 2",
			Status:         redmine.IDName{Name: "完了"},
			AssignedTo:     nil, // 未割り当て
		},
	}

	// 出力
	var buf bytes.Buffer
	if err := fmtr.Format(issues, &buf); err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()

	// 検証
	if !strings.Contains(output, "Test Issue 1") {
		t.Error("Output should contain 'Test Issue 1'")
	}
	if !strings.Contains(output, "Test Issue 2") {
		t.Error("Output should contain 'Test Issue 2'")
	}
	if !strings.Contains(output, "進行中") {
		t.Error("Output should contain '進行中'")
	}
	if !strings.Contains(output, "山田太郎") {
		t.Error("Output should contain '山田太郎'")
	}
}

func TestTemplateFormatter_SetMode(t *testing.T) {
	tmpDir := t.TempDir()
	tmplFile := filepath.Join(tmpDir, "test.tmpl")

	tmplContent := `Mode: {{ .Mode }}`
	if err := os.WriteFile(tmplFile, []byte(tmplContent), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	fmtr, err := NewTemplateFormatter(tmplFile)
	if err != nil {
		t.Fatalf("NewTemplateFormatter() error = %v", err)
	}

	fmtr.SetMode("full", []string{"tag1", "tag2"})

	if fmtr.mode != "full" {
		t.Errorf("mode = %s, want full", fmtr.mode)
	}

	if len(fmtr.tagNames) != 2 {
		t.Errorf("len(tagNames) = %d, want 2", len(fmtr.tagNames))
	}
}

func TestTemplateFuncs_FormatDate(t *testing.T) {
	tmpDir := t.TempDir()
	tmplFile := filepath.Join(tmpDir, "test.tmpl")

	tmplContent := `{{ range .Issues }}{{ formatDate .StartDate }}{{ end }}`
	if err := os.WriteFile(tmplFile, []byte(tmplContent), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	fmtr, err := NewTemplateFormatter(tmplFile)
	if err != nil {
		t.Fatalf("NewTemplateFormatter() error = %v", err)
	}

	date := &redmine.Date{Time: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)}
	issues := []*redmine.Issue{
		{StartDate: date},
	}

	var buf bytes.Buffer
	if err := fmtr.Format(issues, &buf); err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	if !strings.Contains(buf.String(), "2025/01/15") {
		t.Errorf("Output = %s, should contain 2025/01/15", buf.String())
	}
}

func TestTemplateFuncs_FormatDateTime(t *testing.T) {
	tmpDir := t.TempDir()
	tmplFile := filepath.Join(tmpDir, "test.tmpl")

	tmplContent := `{{ range .Issues }}{{ formatDateTime .UpdatedOn }}{{ end }}`
	if err := os.WriteFile(tmplFile, []byte(tmplContent), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	fmtr, err := NewTemplateFormatter(tmplFile)
	if err != nil {
		t.Fatalf("NewTemplateFormatter() error = %v", err)
	}

	dt := &redmine.DateTime{Time: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)}
	issues := []*redmine.Issue{
		{UpdatedOn: dt},
	}

	var buf bytes.Buffer
	if err := fmtr.Format(issues, &buf); err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	if !strings.Contains(buf.String(), "2025/01/15") {
		t.Errorf("Output = %s, should contain date", buf.String())
	}
}

func TestTemplateFuncs_HasChildren(t *testing.T) {
	tmpDir := t.TempDir()
	tmplFile := filepath.Join(tmpDir, "test.tmpl")

	tmplContent := `{{ range .Issues }}{{ if hasChildren . }}YES{{ else }}NO{{ end }}{{ end }}`
	if err := os.WriteFile(tmplFile, []byte(tmplContent), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	fmtr, err := NewTemplateFormatter(tmplFile)
	if err != nil {
		t.Fatalf("NewTemplateFormatter() error = %v", err)
	}

	issues := []*redmine.Issue{
		{
			ID:       1,
			Children: []*redmine.Issue{{ID: 2}},
		},
		{
			ID:       3,
			Children: nil,
		},
	}

	var buf bytes.Buffer
	if err := fmtr.Format(issues, &buf); err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "YES") {
		t.Error("Output should contain 'YES' for issue with children")
	}
	if !strings.Contains(output, "NO") {
		t.Error("Output should contain 'NO' for issue without children")
	}
}

func TestTemplateFuncs_CommentCount(t *testing.T) {
	tmpDir := t.TempDir()
	tmplFile := filepath.Join(tmpDir, "test.tmpl")

	tmplContent := `{{ range .Issues }}{{ commentCount . }}{{ end }}`
	if err := os.WriteFile(tmplFile, []byte(tmplContent), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	fmtr, err := NewTemplateFormatter(tmplFile)
	if err != nil {
		t.Fatalf("NewTemplateFormatter() error = %v", err)
	}

	issues := []*redmine.Issue{
		{
			Journals: []redmine.Journal{
				{Notes: "Comment 1"},
				{Notes: "Comment 2"},
				{Notes: ""}, // 空のコメント
			},
		},
	}

	var buf bytes.Buffer
	if err := fmtr.Format(issues, &buf); err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	if !strings.Contains(buf.String(), "2") {
		t.Errorf("Output = %s, should contain '2'", buf.String())
	}
}

func TestTemplateFuncs_LatestComment(t *testing.T) {
	tmpDir := t.TempDir()
	tmplFile := filepath.Join(tmpDir, "test.tmpl")

	tmplContent := `{{ range .Issues }}{{ latestComment . }}{{ end }}`
	if err := os.WriteFile(tmplFile, []byte(tmplContent), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	fmtr, err := NewTemplateFormatter(tmplFile)
	if err != nil {
		t.Fatalf("NewTemplateFormatter() error = %v", err)
	}

	issues := []*redmine.Issue{
		{
			Journals: []redmine.Journal{
				{Notes: "Old comment"},
				{Notes: "Latest comment"},
			},
		},
	}

	var buf bytes.Buffer
	if err := fmtr.Format(issues, &buf); err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	if !strings.Contains(buf.String(), "Latest comment") {
		t.Errorf("Output = %s, should contain 'Latest comment'", buf.String())
	}
}
