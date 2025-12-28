package stats

import (
	"time"

	"github.com/tktomaru/redmine-exporter/internal/processor"
	"github.com/tktomaru/redmine-exporter/internal/redmine"
)

// WeeklyStats は週報向けの統計情報
type WeeklyStats struct {
	TotalIssues   int                  // 総チケット数
	ByStatus      map[string]int       // ステータス別件数
	ByAssignee    map[string]int       // 担当者別件数
	ByTracker     map[string]int       // トラッカー別件数
	ByPriority    map[string]int       // 優先度別件数
	NewIssues     int                  // 新規作成チケット数（期間内）
	UpdatedIssues int                  // 更新チケット数（期間内）
	ClosedIssues  int                  // 完了チケット数
	OverdueTasks  []*redmine.Issue     // 期限切れタスク
	DueSoonTasks  []*redmine.Issue     // 期限間近タスク（7日以内）
	CommentStats  CommentStats         // コメント統計
}

// CommentStats はコメントの統計情報
type CommentStats struct {
	TotalComments    int            // 総コメント数
	IssuesWithComments int          // コメントのあるチケット数
	ByUser           map[string]int // ユーザー別コメント数
}

// Calculate は週報統計を計算
// weekStart, weekEndは集計期間（期限切れ・期限間近の判定に使用）
func Calculate(issues []*redmine.Issue, weekStart, weekEnd time.Time) *WeeklyStats {
	stats := &WeeklyStats{
		ByStatus:   make(map[string]int),
		ByAssignee: make(map[string]int),
		ByTracker:  make(map[string]int),
		ByPriority: make(map[string]int),
		CommentStats: CommentStats{
			ByUser: make(map[string]int),
		},
		OverdueTasks: make([]*redmine.Issue, 0),
		DueSoonTasks: make([]*redmine.Issue, 0),
	}

	now := time.Now()
	dueSoonThreshold := now.AddDate(0, 0, 7) // 7日後

	// 全チケットを集計
	allIssues := flattenIssues(issues)
	stats.TotalIssues = len(allIssues)

	for _, issue := range allIssues {
		// ステータス別
		statusName := issue.Status.Name
		if statusName == "" {
			statusName = "未設定"
		}
		stats.ByStatus[statusName]++

		// 担当者別
		assignee := processor.GetAssignee(issue)
		stats.ByAssignee[assignee]++

		// トラッカー別
		trackerName := issue.Tracker.Name
		if trackerName == "" {
			trackerName = "未設定"
		}
		stats.ByTracker[trackerName]++

		// 優先度別
		priorityName := issue.Priority.Name
		if priorityName == "" {
			priorityName = "未設定"
		}
		stats.ByPriority[priorityName]++

		// 新規作成・更新・完了の判定
		if issue.CreatedOn != nil && !issue.CreatedOn.IsZero() {
			if issue.CreatedOn.After(weekStart) && issue.CreatedOn.Before(weekEnd) {
				stats.NewIssues++
			}
		}

		if issue.UpdatedOn != nil && !issue.UpdatedOn.IsZero() {
			if issue.UpdatedOn.After(weekStart) && issue.UpdatedOn.Before(weekEnd) {
				stats.UpdatedIssues++
			}
		}

		// 完了判定（ステータスに「完了」「終了」「クローズ」などが含まれる）
		if isClosedStatus(statusName) {
			stats.ClosedIssues++
		}

		// 期限切れ・期限間近の判定
		if issue.DueDate != nil && !issue.DueDate.IsZero() {
			dueDate := issue.DueDate.Time

			// 期限切れ（期限が現在より前）
			if dueDate.Before(now) && !isClosedStatus(statusName) {
				stats.OverdueTasks = append(stats.OverdueTasks, issue)
			} else if dueDate.After(now) && dueDate.Before(dueSoonThreshold) && !isClosedStatus(statusName) {
				// 期限間近（7日以内）
				stats.DueSoonTasks = append(stats.DueSoonTasks, issue)
			}
		}

		// コメント統計
		commentCount := 0
		for _, journal := range issue.Journals {
			if journal.Notes != "" {
				commentCount++
				stats.CommentStats.TotalComments++

				// ユーザー別カウント
				userName := journal.User.Name
				if userName == "" {
					userName = "不明"
				}
				stats.CommentStats.ByUser[userName]++
			}
		}

		if commentCount > 0 {
			stats.CommentStats.IssuesWithComments++
		}
	}

	return stats
}

// flattenIssues はツリー構造のチケットをフラットなリストに変換
func flattenIssues(issues []*redmine.Issue) []*redmine.Issue {
	result := make([]*redmine.Issue, 0)

	for _, issue := range issues {
		result = append(result, issue)

		// 子チケットも再帰的に追加
		if len(issue.Children) > 0 {
			result = append(result, flattenIssues(issue.Children)...)
		}
	}

	return result
}

// isClosedStatus はステータスが完了系かどうかを判定
func isClosedStatus(status string) bool {
	closedKeywords := []string{"完了", "終了", "クローズ", "Closed", "Resolved", "Done"}

	for _, keyword := range closedKeywords {
		if contains(status, keyword) {
			return true
		}
	}

	return false
}

// contains は文字列に部分文字列が含まれるかチェック
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
