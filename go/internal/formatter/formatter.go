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
}

// DetectFormatter は拡張子から適切なフォーマッターを返す
func DetectFormatter(filename string) (Formatter, error) {
	switch {
	case strings.HasSuffix(filename, ".md"):
		return &MarkdownFormatter{}, nil
	case strings.HasSuffix(filename, ".txt"):
		return &TextFormatter{}, nil
	case strings.HasSuffix(filename, ".xlsx"):
		return &ExcelFormatter{filename: filename}, nil
	default:
		return nil, fmt.Errorf("未対応の拡張子: %s (.md, .txt, .xlsx のみ対応)", filename)
	}
}

// formatDate は日付をフォーマット
func formatDate(d *redmine.Date) string {
	if d == nil {
		return "----/--/--"
	}
	return d.Format()
}
