package processor

import (
	"sort"
	"strings"

	"github.com/tktomaru/redmine-exporter/internal/redmine"
)

// Sorter はチケットをソートするインターフェース
type Sorter interface {
	Sort(issues []*redmine.Issue)
}

// NewSorter は指定されたソート方法に応じたSorterを作成
// 形式: "field" または "field:order" または "field_order"
// 例: "updated_on", "updated_on:asc", "updated_on_desc"
func NewSorter(sortBy string) Sorter {
	// コロン区切りの形式をパース (例: "updated_on:asc")
	field := sortBy
	order := "" // デフォルトはフィールドごとに異なる

	if strings.Contains(sortBy, ":") {
		parts := strings.SplitN(sortBy, ":", 2)
		field = strings.TrimSpace(parts[0])
		order = strings.TrimSpace(parts[1])
	}

	// order が指定されていない場合、_asc や _desc サフィックスをチェック
	if order == "" {
		if strings.HasSuffix(field, "_asc") {
			field = strings.TrimSuffix(field, "_asc")
			order = "asc"
		} else if strings.HasSuffix(field, "_desc") {
			field = strings.TrimSuffix(field, "_desc")
			order = "desc"
		}
	}

	// フィールドごとのデフォルト順序を決定
	var defaultDesc bool
	switch field {
	case "updated_on", "created_on":
		defaultDesc = true // 日時は降順（最新が先）がデフォルト
	case "due_date", "start_date":
		defaultDesc = false // 日付は昇順（近い順）がデフォルト
	case "priority":
		defaultDesc = true // 優先度は降順（高い順）がデフォルト
	case "id":
		defaultDesc = false // IDは昇順がデフォルト
	default:
		return nil // 未対応のフィールド
	}

	// order に基づいて降順フラグを設定
	desc := defaultDesc
	if order == "asc" {
		desc = false
	} else if order == "desc" {
		desc = true
	}

	// Sorterを作成
	switch field {
	case "updated_on":
		return &UpdatedOnSorter{Desc: desc}
	case "created_on":
		return &CreatedOnSorter{Desc: desc}
	case "due_date":
		return &DueDateSorter{Desc: desc}
	case "start_date":
		return &StartDateSorter{Desc: desc}
	case "priority":
		return &PrioritySorter{Desc: desc}
	case "id":
		return &IDSorter{Desc: desc}
	default:
		return nil // ソートなし
	}
}

// UpdatedOnSorter は更新日時でソート
type UpdatedOnSorter struct {
	Desc bool // 降順フラグ
}

func (s *UpdatedOnSorter) Sort(issues []*redmine.Issue) {
	sort.Slice(issues, func(i, j int) bool {
		// UpdatedOnがnilの場合は最後尾に
		if issues[i].UpdatedOn == nil {
			return false
		}
		if issues[j].UpdatedOn == nil {
			return true
		}

		if s.Desc {
			return issues[i].UpdatedOn.After(issues[j].UpdatedOn.Time)
		}
		return issues[i].UpdatedOn.Before(issues[j].UpdatedOn.Time)
	})
}

// CreatedOnSorter は作成日時でソート
type CreatedOnSorter struct {
	Desc bool
}

func (s *CreatedOnSorter) Sort(issues []*redmine.Issue) {
	sort.Slice(issues, func(i, j int) bool {
		if issues[i].CreatedOn == nil {
			return false
		}
		if issues[j].CreatedOn == nil {
			return true
		}

		if s.Desc {
			return issues[i].CreatedOn.After(issues[j].CreatedOn.Time)
		}
		return issues[i].CreatedOn.Before(issues[j].CreatedOn.Time)
	})
}

// DueDateSorter は期日でソート
type DueDateSorter struct {
	Desc bool
}

func (s *DueDateSorter) Sort(issues []*redmine.Issue) {
	sort.Slice(issues, func(i, j int) bool {
		// DueDateがnilの場合は最後尾に
		if issues[i].DueDate == nil || issues[i].DueDate.IsZero() {
			return false
		}
		if issues[j].DueDate == nil || issues[j].DueDate.IsZero() {
			return true
		}

		if s.Desc {
			return issues[i].DueDate.After(issues[j].DueDate.Time)
		}
		return issues[i].DueDate.Before(issues[j].DueDate.Time)
	})
}

// StartDateSorter は開始日でソート
type StartDateSorter struct {
	Desc bool
}

func (s *StartDateSorter) Sort(issues []*redmine.Issue) {
	sort.Slice(issues, func(i, j int) bool {
		if issues[i].StartDate == nil || issues[i].StartDate.IsZero() {
			return false
		}
		if issues[j].StartDate == nil || issues[j].StartDate.IsZero() {
			return true
		}

		if s.Desc {
			return issues[i].StartDate.After(issues[j].StartDate.Time)
		}
		return issues[i].StartDate.Before(issues[j].StartDate.Time)
	})
}

// PrioritySorter は優先度でソート
type PrioritySorter struct {
	Desc bool
}

func (s *PrioritySorter) Sort(issues []*redmine.Issue) {
	sort.Slice(issues, func(i, j int) bool {
		// 優先度IDが大きいほど優先度が高いと仮定
		if s.Desc {
			return issues[i].Priority.ID > issues[j].Priority.ID
		}
		return issues[i].Priority.ID < issues[j].Priority.ID
	})
}

// IDSorter はチケットIDでソート
type IDSorter struct {
	Desc bool
}

func (s *IDSorter) Sort(issues []*redmine.Issue) {
	sort.Slice(issues, func(i, j int) bool {
		if s.Desc {
			return issues[i].ID > issues[j].ID
		}
		return issues[i].ID < issues[j].ID
	})
}
