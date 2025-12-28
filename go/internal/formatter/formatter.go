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
// templatePathが指定されている場合、そちらを優先
func DetectFormatter(filename string, mode string, tagNames []string, templatePath string) (Formatter, error) {
	var formatter Formatter

	// テンプレートが指定されている場合はTemplateFormatterを使用
	if templatePath != "" {
		tmplFormatter, err := NewTemplateFormatter(templatePath)
		if err != nil {
			return nil, err
		}
		formatter = tmplFormatter
	} else if strings.HasSuffix(filename, ".tmpl") {
		// 出力ファイル自体が.tmplの場合もTemplateFormatterを使用
		tmplFormatter, err := NewTemplateFormatter(filename)
		if err != nil {
			return nil, err
		}
		formatter = tmplFormatter
	} else {
		// 既存のフォーマッター
		switch {
		case strings.HasSuffix(filename, ".md"):
			formatter = &MarkdownFormatter{}
		case strings.HasSuffix(filename, ".txt"):
			formatter = &TextFormatter{}
		case strings.HasSuffix(filename, ".xlsx"):
			formatter = &ExcelFormatter{filename: filename}
		default:
			return nil, fmt.Errorf("未対応の拡張子: %s (.md, .txt, .xlsx, .tmpl のみ対応)", filename)
		}
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
