package formatter

import (
	"fmt"
	"io"
	"os"
	"text/template"
	"time"

	"github.com/tktomaru/redmine-exporter/internal/processor"
	"github.com/tktomaru/redmine-exporter/internal/redmine"
	"github.com/tktomaru/redmine-exporter/internal/stats"
)

// TemplateFormatter はGo templateを使用した出力
type TemplateFormatter struct {
	tmplPath  string
	mode      string
	tagNames  []string
	tmpl      *template.Template
	stats     *stats.WeeklyStats // 統計情報
	weekStart time.Time
	weekEnd   time.Time
}

// TemplateData はテンプレートに渡すデータ
type TemplateData struct {
	Now       time.Time
	Issues    []*redmine.Issue
	Mode      string
	TagNames  []string
	Stats     *stats.WeeklyStats // 統計情報（オプション）
	WeekStart time.Time          // 週の開始日（統計計算用）
	WeekEnd   time.Time          // 週の終了日（統計計算用）
}

// NewTemplateFormatter は新しいTemplateFormatterを作成
func NewTemplateFormatter(tmplPath string) (*TemplateFormatter, error) {
	// テンプレートファイルを読み込む
	tmpl, err := template.New("").Funcs(templateFuncs()).ParseFiles(tmplPath)
	if err != nil {
		return nil, fmt.Errorf("テンプレート読み込みエラー: %w", err)
	}

	return &TemplateFormatter{
		tmplPath: tmplPath,
		tmpl:     tmpl,
	}, nil
}

// Format はテンプレートを使用して出力
func (f *TemplateFormatter) Format(roots []*redmine.Issue, w io.Writer) error {
	data := TemplateData{
		Now:       time.Now(),
		Issues:    roots,
		Mode:      f.mode,
		TagNames:  f.tagNames,
		Stats:     f.stats,
		WeekStart: f.weekStart,
		WeekEnd:   f.weekEnd,
	}

	// テンプレート名はファイル名のベース名
	tmplName := f.tmplPath
	if stat, err := os.Stat(f.tmplPath); err == nil {
		tmplName = stat.Name()
	}

	// テンプレートを実行
	if err := f.tmpl.ExecuteTemplate(w, tmplName, data); err != nil {
		return fmt.Errorf("テンプレート実行エラー: %w", err)
	}

	return nil
}

// SetMode はモードとタグ名を設定
func (f *TemplateFormatter) SetMode(mode string, tagNames []string) {
	f.mode = mode
	f.tagNames = tagNames
}

// SetStats は統計情報を設定
func (f *TemplateFormatter) SetStats(stats *stats.WeeklyStats, weekStart, weekEnd time.Time) {
	f.stats = stats
	f.weekStart = weekStart
	f.weekEnd = weekEnd
}

// templateFuncs はテンプレートで使用できる関数を定義
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		// 日付フォーマット
		"formatDate": func(d *redmine.Date) string {
			if d == nil {
				return "----/--/--"
			}
			return d.Format()
		},

		// 日時フォーマット
		"formatDateTime": func(dt *redmine.DateTime) string {
			if dt == nil {
				return "----/--/-- --:--:--"
			}
			return dt.Time.Format("2006/01/02 15:04:05")
		},

		// 担当者名
		"assignee": func(issue *redmine.Issue) string {
			return processor.GetAssignee(issue)
		},

		// ステータス名
		"status": func(issue *redmine.Issue) string {
			if issue.Status.Name != "" {
				return issue.Status.Name
			}
			return "未設定"
		},

		// トラッカー名
		"tracker": func(issue *redmine.Issue) string {
			if issue.Tracker.Name != "" {
				return issue.Tracker.Name
			}
			return "未設定"
		},

		// プロジェクト名
		"project": func(issue *redmine.Issue) string {
			if issue.Project.Name != "" {
				return issue.Project.Name
			}
			return "未設定"
		},

		// 優先度名
		"priority": func(issue *redmine.Issue) string {
			if issue.Priority.Name != "" {
				return issue.Priority.Name
			}
			return "未設定"
		},

		// チケットURL
		"ticketURL": func(issue *redmine.Issue, baseURL string) string {
			if baseURL == "" {
				return ""
			}
			return fmt.Sprintf("%s/issues/%d", baseURL, issue.ID)
		},

		// 子チケットの有無
		"hasChildren": func(issue *redmine.Issue) bool {
			return len(issue.Children) > 0
		},

		// コメント数
		"commentCount": func(issue *redmine.Issue) int {
			count := 0
			for _, j := range issue.Journals {
				if j.Notes != "" {
					count++
				}
			}
			return count
		},

		// 最新コメント
		"latestComment": func(issue *redmine.Issue) string {
			for i := len(issue.Journals) - 1; i >= 0; i-- {
				if issue.Journals[i].Notes != "" {
					return issue.Journals[i].Notes
				}
			}
			return ""
		},

		// 最新コメントのユーザー
		"latestCommentUser": func(issue *redmine.Issue) string {
			for i := len(issue.Journals) - 1; i >= 0; i-- {
				if issue.Journals[i].Notes != "" {
					return issue.Journals[i].User.Name
				}
			}
			return ""
		},
	}
}
