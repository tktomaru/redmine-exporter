package formatter

import (
	"fmt"
	"io"

	"github.com/tktomaru/redmine-exporter/internal/processor"
	"github.com/tktomaru/redmine-exporter/internal/redmine"
)

// MarkdownFormatter はMarkdown形式で出力
type MarkdownFormatter struct {
	mode     string
	tagNames []string
}

// Format はMarkdown形式で出力
func (f *MarkdownFormatter) Format(roots []*redmine.Issue, w io.Writer) error {
	for _, parent := range roots {
		if len(parent.Children) > 0 {
			// 親タスク（見出し）
			fmt.Fprintf(w, "# %s\n\n", parent.CleanedSubject)

			// 子タスク（箇条書き）
			for _, child := range parent.Children {
				f.printIssueDetails(w, child, "子")
			}

			fmt.Fprintln(w) // 親タスク間の空行
		} else {
			// スタンドアロンチケット（子を持たない）も見出しとして出力
			fmt.Fprintf(w, "# %s\n\n", parent.CleanedSubject)
			f.printIssueDetails(w, parent, "単独")
		}
	}

	return nil
}

// SetMode はモードとタグ名を設定
func (f *MarkdownFormatter) SetMode(mode string, tagNames []string) {
	f.mode = mode
	f.tagNames = tagNames
}

// printIssueDetails はモードに応じてチケットの詳細を出力
func (f *MarkdownFormatter) printIssueDetails(w io.Writer, issue *redmine.Issue, issueType string) {
	assignee := processor.GetAssignee(issue)
	startDate := formatDate(issue.StartDate)
	dueDate := formatDate(issue.DueDate)

	switch f.mode {
	case "full":
		// フルモード：すべての情報を表示
		if issueType == "子" {
			// 子タスクはリスト形式
			fmt.Fprintf(w, "- **%s** [%s] %s-%s 担当: %s\n",
				issue.CleanedSubject, issue.Status.Name, startDate, dueDate, assignee)
		} else {
			// スタンドアロンタスクは箇条書き形式
			fmt.Fprintf(w, "**ステータス**: %s | **期間**: %s-%s | **担当**: %s\n\n",
				issue.Status.Name, startDate, dueDate, assignee)
		}

		fmt.Fprintf(w, "  - **ID**: %d\n", issue.ID)
		fmt.Fprintf(w, "  - **プロジェクト**: %s\n", issue.Project.Name)
		fmt.Fprintf(w, "  - **トラッカー**: %s\n", issue.Tracker.Name)
		fmt.Fprintf(w, "  - **優先度**: %s\n", issue.Priority.Name)

		if issue.Description != "" {
			fmt.Fprintf(w, "  - **説明**: %s\n", issue.Description)
		}
		if len(issue.Journals) > 0 {
			fmt.Fprintf(w, "  - **コメント数**: %d\n", len(issue.Journals))
		}
		fmt.Fprintln(w)

	case "tags":
		// タグモード：指定されたタグの内容を表示
		if issueType == "子" {
			fmt.Fprintf(w, "- **%s** [%s] %s-%s 担当: %s\n",
				issue.CleanedSubject, issue.Status.Name, startDate, dueDate, assignee)
		} else {
			fmt.Fprintf(w, "**ステータス**: %s | **期間**: %s-%s | **担当**: %s\n\n",
				issue.Status.Name, startDate, dueDate, assignee)
		}

		for _, tagName := range f.tagNames {
			if contents, ok := issue.ExtractedTags[tagName]; ok && len(contents) > 0 {
				if len(contents) == 1 {
					// 1つだけの場合は単一行で表示
					fmt.Fprintf(w, "  - **[%s]**: %s\n", tagName, contents[0])
				} else {
					// 複数ある場合はリスト形式で表示
					fmt.Fprintf(w, "  - **[%s]**:\n", tagName)
					for i, content := range contents {
						fmt.Fprintf(w, "    %d. %s\n", i+1, content)
					}
				}
			}
		}
		fmt.Fprintln(w)

	default:
		// summaryモード：要約のみ表示（デフォルト）
		if issueType == "子" {
			fmt.Fprintf(w, "- **%s** [%s] %s-%s 担当: %s\n",
				issue.CleanedSubject, issue.Status.Name, startDate, dueDate, assignee)

			if issue.Summary != "" {
				fmt.Fprintf(w, "  > %s\n", issue.Summary)
			}
		} else {
			// スタンドアロンチケット
			fmt.Fprintf(w, "**ステータス**: %s | **期間**: %s-%s | **担当**: %s\n\n",
				issue.Status.Name, startDate, dueDate, assignee)

			if issue.Summary != "" {
				fmt.Fprintf(w, "> %s\n\n", issue.Summary)
			}
		}
	}
}
