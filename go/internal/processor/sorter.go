package processor

import (
	"sort"

	"github.com/tktomaru/redmine-exporter/internal/redmine"
)

// Sorter はチケットをソートするインターフェース
type Sorter interface {
	Sort(issues []*redmine.Issue)
}

// NewSorter は指定されたソート方法に応じたSorterを作成
func NewSorter(sortBy string) Sorter {
	switch sortBy {
	case "updated_on":
		return &UpdatedOnSorter{Desc: true} // デフォルトは降順（最新が先）
	case "updated_on_asc":
		return &UpdatedOnSorter{Desc: false}
	case "created_on":
		return &CreatedOnSorter{Desc: true}
	case "created_on_asc":
		return &CreatedOnSorter{Desc: false}
	case "due_date":
		return &DueDateSorter{Desc: false} // 期日は昇順（近い順）がデフォルト
	case "due_date_desc":
		return &DueDateSorter{Desc: true}
	case "start_date":
		return &StartDateSorter{Desc: false}
	case "start_date_desc":
		return &StartDateSorter{Desc: true}
	case "priority":
		return &PrioritySorter{Desc: true} // 優先度は降順（高い順）
	case "priority_asc":
		return &PrioritySorter{Desc: false}
	case "id":
		return &IDSorter{Desc: false} // ID昇順
	case "id_desc":
		return &IDSorter{Desc: true}
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
