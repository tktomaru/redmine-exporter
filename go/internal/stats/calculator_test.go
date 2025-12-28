package stats

import (
	"testing"
	"time"

	"github.com/tktomaru/redmine-exporter/internal/redmine"
)

func TestCalculate(t *testing.T) {
	now := time.Now()
	weekStart := now.AddDate(0, 0, -7)
	weekEnd := now

	issues := []*redmine.Issue{
		{
			ID:      1,
			Subject: "Test Issue 1",
			Status:  redmine.IDName{Name: "進行中"},
			AssignedTo: &redmine.IDName{Name: "山田太郎"},
			Tracker: redmine.IDName{Name: "バグ"},
			Priority: redmine.IDName{Name: "高"},
			CreatedOn: &redmine.DateTime{Time: now.AddDate(0, 0, -1)},
			UpdatedOn: &redmine.DateTime{Time: now.AddDate(0, 0, -1)},
			DueDate:   &redmine.Date{Time: now.AddDate(0, 0, 5)}, // 期限間近
			Journals: []redmine.Journal{
				{Notes: "Comment 1", User: redmine.IDName{Name: "山田太郎"}},
				{Notes: "Comment 2", User: redmine.IDName{Name: "佐藤花子"}},
			},
		},
		{
			ID:      2,
			Subject: "Test Issue 2",
			Status:  redmine.IDName{Name: "完了"},
			AssignedTo: &redmine.IDName{Name: "佐藤花子"},
			Tracker: redmine.IDName{Name: "機能"},
			Priority: redmine.IDName{Name: "中"},
			CreatedOn: &redmine.DateTime{Time: now.AddDate(0, 0, -10)},
			UpdatedOn: &redmine.DateTime{Time: now.AddDate(0, 0, -2)},
		},
		{
			ID:      3,
			Subject: "Test Issue 3",
			Status:  redmine.IDName{Name: "進行中"},
			AssignedTo: &redmine.IDName{Name: "山田太郎"},
			Tracker: redmine.IDName{Name: "バグ"},
			Priority: redmine.IDName{Name: "高"},
			CreatedOn: &redmine.DateTime{Time: now.AddDate(0, 0, -15)},
			UpdatedOn: &redmine.DateTime{Time: now.AddDate(0, 0, -15)},
			DueDate:   &redmine.Date{Time: now.AddDate(0, 0, -1)}, // 期限切れ
		},
	}

	stats := Calculate(issues, weekStart, weekEnd)

	// 総チケット数
	if stats.TotalIssues != 3 {
		t.Errorf("TotalIssues = %d, want 3", stats.TotalIssues)
	}

	// ステータス別
	if stats.ByStatus["進行中"] != 2 {
		t.Errorf("ByStatus[進行中] = %d, want 2", stats.ByStatus["進行中"])
	}
	if stats.ByStatus["完了"] != 1 {
		t.Errorf("ByStatus[完了] = %d, want 1", stats.ByStatus["完了"])
	}

	// 担当者別
	if stats.ByAssignee["山田太郎"] != 2 {
		t.Errorf("ByAssignee[山田太郎] = %d, want 2", stats.ByAssignee["山田太郎"])
	}
	if stats.ByAssignee["佐藤花子"] != 1 {
		t.Errorf("ByAssignee[佐藤花子] = %d, want 1", stats.ByAssignee["佐藤花子"])
	}

	// トラッカー別
	if stats.ByTracker["バグ"] != 2 {
		t.Errorf("ByTracker[バグ] = %d, want 2", stats.ByTracker["バグ"])
	}
	if stats.ByTracker["機能"] != 1 {
		t.Errorf("ByTracker[機能] = %d, want 1", stats.ByTracker["機能"])
	}

	// 優先度別
	if stats.ByPriority["高"] != 2 {
		t.Errorf("ByPriority[高] = %d, want 2", stats.ByPriority["高"])
	}
	if stats.ByPriority["中"] != 1 {
		t.Errorf("ByPriority[中] = %d, want 1", stats.ByPriority["中"])
	}

	// 新規作成（過去7日以内）
	if stats.NewIssues != 1 {
		t.Errorf("NewIssues = %d, want 1", stats.NewIssues)
	}

	// 更新（過去7日以内）
	if stats.UpdatedIssues != 2 {
		t.Errorf("UpdatedIssues = %d, want 2", stats.UpdatedIssues)
	}

	// 完了
	if stats.ClosedIssues != 1 {
		t.Errorf("ClosedIssues = %d, want 1", stats.ClosedIssues)
	}

	// 期限切れ（完了していないもの）
	if len(stats.OverdueTasks) != 1 {
		t.Errorf("len(OverdueTasks) = %d, want 1", len(stats.OverdueTasks))
	}

	// 期限間近（7日以内、完了していないもの）
	if len(stats.DueSoonTasks) != 1 {
		t.Errorf("len(DueSoonTasks) = %d, want 1", len(stats.DueSoonTasks))
	}

	// コメント統計
	if stats.CommentStats.TotalComments != 2 {
		t.Errorf("CommentStats.TotalComments = %d, want 2", stats.CommentStats.TotalComments)
	}
	if stats.CommentStats.IssuesWithComments != 1 {
		t.Errorf("CommentStats.IssuesWithComments = %d, want 1", stats.CommentStats.IssuesWithComments)
	}
	if stats.CommentStats.ByUser["山田太郎"] != 1 {
		t.Errorf("CommentStats.ByUser[山田太郎] = %d, want 1", stats.CommentStats.ByUser["山田太郎"])
	}
	if stats.CommentStats.ByUser["佐藤花子"] != 1 {
		t.Errorf("CommentStats.ByUser[佐藤花子] = %d, want 1", stats.CommentStats.ByUser["佐藤花子"])
	}
}

func TestCalculate_WithChildren(t *testing.T) {
	now := time.Now()
	weekStart := now.AddDate(0, 0, -7)
	weekEnd := now

	// 親子構造のチケット
	issues := []*redmine.Issue{
		{
			ID:      1,
			Subject: "Parent Issue",
			Status:  redmine.IDName{Name: "進行中"},
			Children: []*redmine.Issue{
				{
					ID:      2,
					Subject: "Child Issue 1",
					Status:  redmine.IDName{Name: "完了"},
				},
				{
					ID:      3,
					Subject: "Child Issue 2",
					Status:  redmine.IDName{Name: "進行中"},
				},
			},
		},
	}

	stats := Calculate(issues, weekStart, weekEnd)

	// 親+子の合計
	if stats.TotalIssues != 3 {
		t.Errorf("TotalIssues = %d, want 3", stats.TotalIssues)
	}

	// ステータス別（親子含む）
	if stats.ByStatus["進行中"] != 2 {
		t.Errorf("ByStatus[進行中] = %d, want 2", stats.ByStatus["進行中"])
	}
	if stats.ByStatus["完了"] != 1 {
		t.Errorf("ByStatus[完了] = %d, want 1", stats.ByStatus["完了"])
	}
}

func TestCalculate_EmptyIssues(t *testing.T) {
	now := time.Now()
	weekStart := now.AddDate(0, 0, -7)
	weekEnd := now

	stats := Calculate([]*redmine.Issue{}, weekStart, weekEnd)

	if stats.TotalIssues != 0 {
		t.Errorf("TotalIssues = %d, want 0", stats.TotalIssues)
	}

	if len(stats.ByStatus) != 0 {
		t.Errorf("len(ByStatus) = %d, want 0", len(stats.ByStatus))
	}
}

func TestIsClosedStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{"完了", "完了", true},
		{"終了", "終了", true},
		{"クローズ", "クローズ", true},
		{"Closed", "Closed", true},
		{"Resolved", "Resolved", true},
		{"Done", "Done", true},
		{"進行中", "進行中", false},
		{"新規", "新規", false},
		{"待機", "待機", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isClosedStatus(tt.status)
			if got != tt.want {
				t.Errorf("isClosedStatus(%s) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestFlattenIssues(t *testing.T) {
	issues := []*redmine.Issue{
		{
			ID: 1,
			Children: []*redmine.Issue{
				{
					ID: 2,
					Children: []*redmine.Issue{
						{ID: 3},
					},
				},
				{ID: 4},
			},
		},
		{ID: 5},
	}

	flattened := flattenIssues(issues)

	// 1, 2, 3, 4, 5 の5件
	if len(flattened) != 5 {
		t.Errorf("len(flattened) = %d, want 5", len(flattened))
	}

	// ID順にチェック
	expectedIDs := []int{1, 2, 3, 4, 5}
	for i, issue := range flattened {
		if issue.ID != expectedIDs[i] {
			t.Errorf("flattened[%d].ID = %d, want %d", i, issue.ID, expectedIDs[i])
		}
	}
}
