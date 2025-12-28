package processor

import (
	"testing"

	"github.com/tktomaru/redmine-exporter/internal/redmine"
)

func TestNewGrouper(t *testing.T) {
	tests := []struct {
		name    string
		groupBy string
		want    Grouper
	}{
		{
			name:    "assignee",
			groupBy: "assignee",
			want:    &AssigneeGrouper{},
		},
		{
			name:    "status",
			groupBy: "status",
			want:    &StatusGrouper{},
		},
		{
			name:    "tracker",
			groupBy: "tracker",
			want:    &TrackerGrouper{},
		},
		{
			name:    "project",
			groupBy: "project",
			want:    &ProjectGrouper{},
		},
		{
			name:    "priority",
			groupBy: "priority",
			want:    &PriorityGrouper{},
		},
		{
			name:    "invalid",
			groupBy: "invalid",
			want:    nil,
		},
		{
			name:    "empty",
			groupBy: "",
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewGrouper(tt.groupBy)
			if (got == nil) != (tt.want == nil) {
				t.Errorf("NewGrouper() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssigneeGrouper_Group(t *testing.T) {
	issues := []*redmine.Issue{
		{
			ID:         1,
			Subject:    "佐藤さんのタスク1",
			AssignedTo: &redmine.IDName{ID: 1, Name: "佐藤"},
		},
		{
			ID:         2,
			Subject:    "鈴木さんのタスク",
			AssignedTo: &redmine.IDName{ID: 2, Name: "鈴木"},
		},
		{
			ID:         3,
			Subject:    "佐藤さんのタスク2",
			AssignedTo: &redmine.IDName{ID: 1, Name: "佐藤"},
		},
		{
			ID:         4,
			Subject:    "担当者未定のタスク",
			AssignedTo: nil,
		},
	}

	grouper := &AssigneeGrouper{}
	result := grouper.Group(issues)

	// グループ数の確認
	if len(result.Groups) != 3 {
		t.Errorf("Group count = %d, want 3", len(result.Groups))
	}

	// 佐藤グループの確認
	if len(result.Groups["佐藤"]) != 2 {
		t.Errorf("佐藤 group size = %d, want 2", len(result.Groups["佐藤"]))
	}

	// 鈴木グループの確認
	if len(result.Groups["鈴木"]) != 1 {
		t.Errorf("鈴木 group size = %d, want 1", len(result.Groups["鈴木"]))
	}

	// 担当者未定グループの確認
	if len(result.Groups["担当者未定"]) != 1 {
		t.Errorf("担当者未定 group size = %d, want 1", len(result.Groups["担当者未定"]))
	}

	// キーの順序確認（出現順）
	if len(result.Keys) != 3 {
		t.Errorf("Keys length = %d, want 3", len(result.Keys))
	}
	if result.Keys[0] != "佐藤" {
		t.Errorf("Keys[0] = %s, want 佐藤", result.Keys[0])
	}
	if result.Keys[1] != "鈴木" {
		t.Errorf("Keys[1] = %s, want 鈴木", result.Keys[1])
	}
	if result.Keys[2] != "担当者未定" {
		t.Errorf("Keys[2] = %s, want 担当者未定", result.Keys[2])
	}
}

func TestStatusGrouper_Group(t *testing.T) {
	issues := []*redmine.Issue{
		{ID: 1, Status: redmine.IDName{ID: 1, Name: "新規"}},
		{ID: 2, Status: redmine.IDName{ID: 2, Name: "進行中"}},
		{ID: 3, Status: redmine.IDName{ID: 1, Name: "新規"}},
		{ID: 4, Status: redmine.IDName{ID: 3, Name: "完了"}},
	}

	grouper := &StatusGrouper{}
	result := grouper.Group(issues)

	// グループ数の確認
	if len(result.Groups) != 3 {
		t.Errorf("Group count = %d, want 3", len(result.Groups))
	}

	// 新規グループの確認
	if len(result.Groups["新規"]) != 2 {
		t.Errorf("新規 group size = %d, want 2", len(result.Groups["新規"]))
	}

	// キーの順序確認
	if result.Keys[0] != "新規" {
		t.Errorf("Keys[0] = %s, want 新規", result.Keys[0])
	}
	if result.Keys[1] != "進行中" {
		t.Errorf("Keys[1] = %s, want 進行中", result.Keys[1])
	}
	if result.Keys[2] != "完了" {
		t.Errorf("Keys[2] = %s, want 完了", result.Keys[2])
	}
}

func TestTrackerGrouper_Group(t *testing.T) {
	issues := []*redmine.Issue{
		{ID: 1, Tracker: redmine.IDName{ID: 1, Name: "バグ"}},
		{ID: 2, Tracker: redmine.IDName{ID: 2, Name: "機能"}},
		{ID: 3, Tracker: redmine.IDName{ID: 1, Name: "バグ"}},
	}

	grouper := &TrackerGrouper{}
	result := grouper.Group(issues)

	if len(result.Groups) != 2 {
		t.Errorf("Group count = %d, want 2", len(result.Groups))
	}

	if len(result.Groups["バグ"]) != 2 {
		t.Errorf("バグ group size = %d, want 2", len(result.Groups["バグ"]))
	}

	if len(result.Groups["機能"]) != 1 {
		t.Errorf("機能 group size = %d, want 1", len(result.Groups["機能"]))
	}
}

func TestProjectGrouper_Group(t *testing.T) {
	issues := []*redmine.Issue{
		{ID: 1, Project: redmine.IDName{ID: 1, Name: "プロジェクトA"}},
		{ID: 2, Project: redmine.IDName{ID: 2, Name: "プロジェクトB"}},
		{ID: 3, Project: redmine.IDName{ID: 1, Name: "プロジェクトA"}},
	}

	grouper := &ProjectGrouper{}
	result := grouper.Group(issues)

	if len(result.Groups) != 2 {
		t.Errorf("Group count = %d, want 2", len(result.Groups))
	}

	if len(result.Groups["プロジェクトA"]) != 2 {
		t.Errorf("プロジェクトA group size = %d, want 2", len(result.Groups["プロジェクトA"]))
	}
}

func TestPriorityGrouper_Group(t *testing.T) {
	issues := []*redmine.Issue{
		{ID: 1, Priority: redmine.IDName{ID: 1, Name: "低"}},
		{ID: 2, Priority: redmine.IDName{ID: 5, Name: "緊急"}},
		{ID: 3, Priority: redmine.IDName{ID: 1, Name: "低"}},
		{ID: 4, Priority: redmine.IDName{ID: 3, Name: "高"}},
	}

	grouper := &PriorityGrouper{}
	result := grouper.Group(issues)

	if len(result.Groups) != 3 {
		t.Errorf("Group count = %d, want 3", len(result.Groups))
	}

	if len(result.Groups["低"]) != 2 {
		t.Errorf("低 group size = %d, want 2", len(result.Groups["低"]))
	}
}

func TestFlattenGroupedIssues(t *testing.T) {
	grouped := &GroupedIssues{
		Groups: map[string][]*redmine.Issue{
			"グループA": {
				{ID: 1, Subject: "A-1"},
				{ID: 2, Subject: "A-2"},
			},
			"グループB": {
				{ID: 3, Subject: "B-1"},
			},
			"グループC": {
				{ID: 4, Subject: "C-1"},
				{ID: 5, Subject: "C-2"},
				{ID: 6, Subject: "C-3"},
			},
		},
		Keys: []string{"グループA", "グループB", "グループC"},
	}

	result := FlattenGroupedIssues(grouped)

	// 全チケット数の確認
	if len(result) != 6 {
		t.Errorf("Flattened issues count = %d, want 6", len(result))
	}

	// 順序の確認（Keys順にフラット化される）
	expectedIDs := []int{1, 2, 3, 4, 5, 6}
	for i, expected := range expectedIDs {
		if result[i].ID != expected {
			t.Errorf("result[%d].ID = %d, want %d", i, result[i].ID, expected)
		}
	}
}
