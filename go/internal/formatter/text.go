package formatter

import (
	"fmt"
	"io"

	"github.com/tktomaru/redmine-exporter/internal/processor"
	"github.com/tktomaru/redmine-exporter/internal/redmine"
)

// TextFormatter はVBA版と同じテキスト形式で出力
type TextFormatter struct{}

// Format はテキスト形式で出力（VBA版の出力形式を再現）
func (f *TextFormatter) Format(roots []*redmine.Issue, w io.Writer) error {
	for _, parent := range roots {
		// 子チケットがある場合は親子形式で出力
		if len(parent.Children) > 0 {
			// 親タスク
			fmt.Fprintf(w, "■%s\n", parent.CleanedSubject)

			// 子タスク
			for _, child := range parent.Children {
				assignee := processor.GetAssignee(child)
				startDate := formatDate(child.StartDate)
				dueDate := formatDate(child.DueDate)

				fmt.Fprintf(w, "・%s 【%s】 %s-%s 担当: %s\n",
					child.CleanedSubject,
					child.Status.Name,
					startDate,
					dueDate,
					assignee,
				)

				// 要約がある場合
				if child.Summary != "" {
					fmt.Fprintf(w, "　⇒%s\n", child.Summary)
				}
			}

			fmt.Fprintln(w) // 親タスク間の空行
		} else {
			// スタンドアロンチケット（子を持たない）も親タスクとして出力
			fmt.Fprintf(w, "■%s\n", parent.CleanedSubject)

			// ステータスや担当者などの詳細情報を表示
			assignee := processor.GetAssignee(parent)
			startDate := formatDate(parent.StartDate)
			dueDate := formatDate(parent.DueDate)

			fmt.Fprintf(w, "　【%s】 %s-%s 担当: %s\n",
				parent.Status.Name,
				startDate,
				dueDate,
				assignee,
			)

			// 要約がある場合
			if parent.Summary != "" {
				fmt.Fprintf(w, "　⇒%s\n", parent.Summary)
			}

			fmt.Fprintln(w) // チケット間の空行
		}
	}

	return nil
}
