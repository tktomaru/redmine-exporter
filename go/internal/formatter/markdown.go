package formatter

import (
	"fmt"
	"io"

	"github.com/tktomaru/redmine-exporter/internal/processor"
	"github.com/tktomaru/redmine-exporter/internal/redmine"
)

// MarkdownFormatter はMarkdown形式で出力
type MarkdownFormatter struct{}

// Format はMarkdown形式で出力
func (f *MarkdownFormatter) Format(roots []*redmine.Issue, w io.Writer) error {
	for _, parent := range roots {
		// 子チケットがある場合は親子形式で出力
		if len(parent.Children) > 0 {
			// 親タスク（見出し）
			fmt.Fprintf(w, "# %s\n\n", parent.CleanedSubject)

			// 子タスク（箇条書き）
			for _, child := range parent.Children {
				assignee := processor.GetAssignee(child)
				startDate := formatDate(child.StartDate)
				dueDate := formatDate(child.DueDate)

				fmt.Fprintf(w, "- **%s** [%s] %s-%s 担当: %s\n",
					child.CleanedSubject,
					child.Status.Name,
					startDate,
					dueDate,
					assignee,
				)

				// 要約がある場合（引用ブロック）
				if child.Summary != "" {
					fmt.Fprintf(w, "  > %s\n", child.Summary)
				}
			}

			fmt.Fprintln(w) // 親タスク間の空行
		} else {
			// スタンドアロンチケット（子を持たない）は単独で出力
			assignee := processor.GetAssignee(parent)
			startDate := formatDate(parent.StartDate)
			dueDate := formatDate(parent.DueDate)

			fmt.Fprintf(w, "- **%s** [%s] %s-%s 担当: %s\n",
				parent.CleanedSubject,
				parent.Status.Name,
				startDate,
				dueDate,
				assignee,
			)

			// 要約がある場合（引用ブロック）
			if parent.Summary != "" {
				fmt.Fprintf(w, "  > %s\n", parent.Summary)
			}

			fmt.Fprintln(w) // チケット間の空行
		}
	}

	return nil
}
