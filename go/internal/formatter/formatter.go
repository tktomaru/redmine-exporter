package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/tktomaru/redmine-exporter/internal/redmine"
)

// Formatter は出力形式のインターフェース
type Formatter interface {
	Format(roots []*redmine.Issue, w io.Writer) error
	SetMode(mode string, tagNames []string)
}

// DetectFormatter は拡張子から適切なフォーマッターを返す
func DetectFormatter(filename string, mode string, tagNames []string) (Formatter, error) {
	var formatter Formatter

	switch {
	case strings.HasSuffix(filename, ".md"):
		formatter = &MarkdownFormatter{}
	case strings.HasSuffix(filename, ".txt"):
		formatter = &TextFormatter{}
	case strings.HasSuffix(filename, ".xlsx"):
		formatter = &ExcelFormatter{filename: filename}
	default:
		return nil, fmt.Errorf("未対応の拡張子: %s (.md, .txt, .xlsx のみ対応)", filename)
	}

	formatter.SetMode(mode, tagNames)
	return formatter, nil
}

// formatDate は日付をフォーマット
func formatDate(d *redmine.Date) string {
	if d == nil {
		return "----/--/--"
	}
	return d.Format()
}
