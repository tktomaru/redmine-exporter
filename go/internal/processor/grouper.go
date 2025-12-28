package processor

import (
	"github.com/tktomaru/redmine-exporter/internal/redmine"
)

// GroupedIssues はグルーピングされたチケット
type GroupedIssues struct {
	Groups map[string][]*redmine.Issue // グループ名 -> チケットリスト
	Keys   []string                     // グループ名のリスト（順序保持）
}

// Grouper はチケットをグルーピングするインターフェース
type Grouper interface {
	Group(issues []*redmine.Issue) *GroupedIssues
}

// NewGrouper は指定されたグルーピング方法に応じたGrouperを作成
func NewGrouper(groupBy string) Grouper {
	switch groupBy {
	case "assignee":
		return &AssigneeGrouper{}
	case "status":
		return &StatusGrouper{}
	case "tracker":
		return &TrackerGrouper{}
	case "project":
		return &ProjectGrouper{}
	case "priority":
		return &PriorityGrouper{}
	default:
		return nil // グルーピングなし
	}
}

// AssigneeGrouper は担当者別にグルーピング
type AssigneeGrouper struct{}

func (g *AssigneeGrouper) Group(issues []*redmine.Issue) *GroupedIssues {
	result := &GroupedIssues{
		Groups: make(map[string][]*redmine.Issue),
		Keys:   []string{},
	}

	keyOrder := make(map[string]bool)

	for _, issue := range issues {
		key := GetAssignee(issue)

		// 初めて出現するキーの場合、順序を記録
		if !keyOrder[key] {
			result.Keys = append(result.Keys, key)
			keyOrder[key] = true
		}

		result.Groups[key] = append(result.Groups[key], issue)
	}

	return result
}

// StatusGrouper はステータス別にグルーピング
type StatusGrouper struct{}

func (g *StatusGrouper) Group(issues []*redmine.Issue) *GroupedIssues {
	result := &GroupedIssues{
		Groups: make(map[string][]*redmine.Issue),
		Keys:   []string{},
	}

	keyOrder := make(map[string]bool)

	for _, issue := range issues {
		key := issue.Status.Name
		if key == "" {
			key = "ステータス未設定"
		}

		if !keyOrder[key] {
			result.Keys = append(result.Keys, key)
			keyOrder[key] = true
		}

		result.Groups[key] = append(result.Groups[key], issue)
	}

	return result
}

// TrackerGrouper はトラッカー別にグルーピング
type TrackerGrouper struct{}

func (g *TrackerGrouper) Group(issues []*redmine.Issue) *GroupedIssues {
	result := &GroupedIssues{
		Groups: make(map[string][]*redmine.Issue),
		Keys:   []string{},
	}

	keyOrder := make(map[string]bool)

	for _, issue := range issues {
		key := issue.Tracker.Name
		if key == "" {
			key = "トラッカー未設定"
		}

		if !keyOrder[key] {
			result.Keys = append(result.Keys, key)
			keyOrder[key] = true
		}

		result.Groups[key] = append(result.Groups[key], issue)
	}

	return result
}

// ProjectGrouper はプロジェクト別にグルーピング
type ProjectGrouper struct{}

func (g *ProjectGrouper) Group(issues []*redmine.Issue) *GroupedIssues {
	result := &GroupedIssues{
		Groups: make(map[string][]*redmine.Issue),
		Keys:   []string{},
	}

	keyOrder := make(map[string]bool)

	for _, issue := range issues {
		key := issue.Project.Name
		if key == "" {
			key = "プロジェクト未設定"
		}

		if !keyOrder[key] {
			result.Keys = append(result.Keys, key)
			keyOrder[key] = true
		}

		result.Groups[key] = append(result.Groups[key], issue)
	}

	return result
}

// PriorityGrouper は優先度別にグルーピング
type PriorityGrouper struct{}

func (g *PriorityGrouper) Group(issues []*redmine.Issue) *GroupedIssues {
	result := &GroupedIssues{
		Groups: make(map[string][]*redmine.Issue),
		Keys:   []string{},
	}

	keyOrder := make(map[string]bool)

	for _, issue := range issues {
		key := issue.Priority.Name
		if key == "" {
			key = "優先度未設定"
		}

		if !keyOrder[key] {
			result.Keys = append(result.Keys, key)
			keyOrder[key] = true
		}

		result.Groups[key] = append(result.Groups[key], issue)
	}

	return result
}

// FlattenGroupedIssues はグルーピングされたチケットをフラットなリストに戻す
// グループの順序とグループ内のチケットの順序を保持
func FlattenGroupedIssues(grouped *GroupedIssues) []*redmine.Issue {
	var result []*redmine.Issue

	for _, key := range grouped.Keys {
		result = append(result, grouped.Groups[key]...)
	}

	return result
}
