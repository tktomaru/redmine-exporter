package formatter

import (
	"fmt"
	"io"

	"github.com/tktomaru/redmine-exporter/internal/processor"
	"github.com/tktomaru/redmine-exporter/internal/redmine"
)

// TextFormatter はVBA版と同じテキスト形式で出力
type TextFormatter struct {
	mode     string
	tagNames []string
}

// SetMode はモードとタグ名を設定
func (f *TextFormatter) SetMode(mode string, tagNames []string) {
	f.mode = mode
	f.tagNames = tagNames
}

// Format はテキスト形式で出力（VBA版の出力形式を再現）
func (f *TextFormatter) Format(roots []*redmine.Issue, w io.Writer) error {
	for _, parent := range roots {
		if len(parent.Children) > 0 {
			// 親タスク
			fmt.Fprintf(w, "■%s　", parent.CleanedSubject)
			f.printIssueDetails(w, parent, "親")

			// 子タスク
			for _, child := range parent.Children {
				fmt.Fprintf(w, "・%s　", child.CleanedSubject)
				f.printIssueDetails(w, child, "子")
			}

			fmt.Fprintln(w)
		} else {
			// スタンドアロンチケット
			fmt.Fprintf(w, "■%s　", parent.CleanedSubject)
			f.printIssueDetails(w, parent, "単独")
			fmt.Fprintln(w)
		}
	}

	return nil
}

// printIssueDetails はモードに応じてチケットの詳細を出力
func (f *TextFormatter) printIssueDetails(w io.Writer, issue *redmine.Issue, issueType string) {
	assignee := processor.GetAssignee(issue)
	startDate := formatDate(issue.StartDate)
	dueDate := formatDate(issue.DueDate)

	switch f.mode {
	case "full":
		// フルモード：すべての情報を表示
		fmt.Fprintf(w, "　ID: %d\n", issue.ID)
		fmt.Fprintf(w, "　プロジェクト: %s\n", issue.Project.Name)
		fmt.Fprintf(w, "　トラッカー: %s\n", issue.Tracker.Name)
		fmt.Fprintf(w, "　ステータス: %s\n", issue.Status.Name)
		fmt.Fprintf(w, "　優先度: %s\n", issue.Priority.Name)
		fmt.Fprintf(w, "　開始日: %s\n", startDate)
		fmt.Fprintf(w, "　終了日: %s\n", dueDate)
		fmt.Fprintf(w, "　担当者: %s\n", assignee)
		if issue.Description != "" {
			fmt.Fprintf(w, "　説明: %s\n", issue.Description)
		}
		if len(issue.Journals) > 0 {
			fmt.Fprintf(w, "　コメント数: %d\n", len(issue.Journals))
		}

	case "tags":
		// タグモード：指定されたタグの内容を表示
		fmt.Fprintf(w, "　【%s】 %s-%s 担当: %s\n", issue.Status.Name, startDate, dueDate, assignee)
		for _, tagName := range f.tagNames {
			if contents, ok := issue.ExtractedTags[tagName]; ok && len(contents) > 0 {
				if len(contents) == 1 {
					// 1つだけの場合は単一行で表示
					fmt.Fprintf(w, "　[%s] %s\n", tagName, contents[0])
				} else {
					// 複数ある場合はリスト形式で表示
					for i, content := range contents {
						fmt.Fprintf(w, "　[%s%d] %s\n", tagName, i+1, content)
					}
				}
			}
		}

	default:
		// summaryモード：要約のみ表示（デフォルト）
		if issueType == "子" {
			fmt.Fprintf(w, "　【%s】 %s-%s 担当: %s\n", issue.Status.Name, startDate, dueDate, assignee)
		} else {
			fmt.Fprintf(w, "　【%s】 %s-%s 担当: %s\n", issue.Status.Name, startDate, dueDate, assignee)
		}
		if issue.Summary != "" {
			fmt.Fprintf(w, "　⇒%s\n", issue.Summary)
		}
	}
}
