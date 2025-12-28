package processor

import (
	"testing"
	"time"

	"github.com/tktomaru/redmine-exporter/internal/redmine"
)

func TestNewSorter(t *testing.T) {
	tests := []struct {
		name   string
		sortBy string
		want   Sorter
	}{
		{
			name:   "updated_on",
			sortBy: "updated_on",
			want:   &UpdatedOnSorter{Desc: true},
		},
		{
			name:   "updated_on_asc",
			sortBy: "updated_on_asc",
			want:   &UpdatedOnSorter{Desc: false},
		},
		{
			name:   "due_date",
			sortBy: "due_date",
			want:   &DueDateSorter{Desc: false},
		},
		{
			name:   "priority",
			sortBy: "priority",
			want:   &PrioritySorter{Desc: true},
		},
		{
			name:   "id",
			sortBy: "id",
			want:   &IDSorter{Desc: false},
		},
		{
			name:   "updated_on:asc (colon format)",
			sortBy: "updated_on:asc",
			want:   &UpdatedOnSorter{Desc: false},
		},
		{
			name:   "updated_on:desc (colon format)",
			sortBy: "updated_on:desc",
			want:   &UpdatedOnSorter{Desc: true},
		},
		{
			name:   "due_date:asc (colon format)",
			sortBy: "due_date:asc",
			want:   &DueDateSorter{Desc: false},
		},
		{
			name:   "due_date:desc (colon format)",
			sortBy: "due_date:desc",
			want:   &DueDateSorter{Desc: true},
		},
		{
			name:   "priority:asc (colon format)",
			sortBy: "priority:asc",
			want:   &PrioritySorter{Desc: false},
		},
		{
			name:   "id:desc (colon format)",
			sortBy: "id:desc",
			want:   &IDSorter{Desc: true},
		},
		{
			name:   "invalid",
			sortBy: "invalid",
			want:   nil,
		},
		{
			name:   "empty",
			sortBy: "",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewSorter(tt.sortBy)
			if (got == nil) != (tt.want == nil) {
				t.Errorf("NewSorter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdatedOnSorter_Sort(t *testing.T) {
	now := time.Now()

	issues := []*redmine.Issue{
		{
			ID:        1,
			Subject:   "古いチケット",
			UpdatedOn: &redmine.DateTime{Time: now.Add(-48 * time.Hour)},
		},
		{
			ID:        2,
			Subject:   "最新チケット",
			UpdatedOn: &redmine.DateTime{Time: now},
		},
		{
			ID:        3,
			Subject:   "中間チケット",
			UpdatedOn: &redmine.DateTime{Time: now.Add(-24 * time.Hour)},
		},
		{
			ID:        4,
			Subject:   "UpdatedOnなし",
			UpdatedOn: nil,
		},
	}

	t.Run("Descending", func(t *testing.T) {
		issuesCopy := make([]*redmine.Issue, len(issues))
		copy(issuesCopy, issues)

		sorter := &UpdatedOnSorter{Desc: true}
		sorter.Sort(issuesCopy)

		// 降順: 最新 -> 中間 -> 古い -> nil
		if issuesCopy[0].ID != 2 {
			t.Errorf("Sort()[0].ID = %d, want 2", issuesCopy[0].ID)
		}
		if issuesCopy[1].ID != 3 {
			t.Errorf("Sort()[1].ID = %d, want 3", issuesCopy[1].ID)
		}
		if issuesCopy[2].ID != 1 {
			t.Errorf("Sort()[2].ID = %d, want 1", issuesCopy[2].ID)
		}
		if issuesCopy[3].ID != 4 {
			t.Errorf("Sort()[3].ID = %d, want 4", issuesCopy[3].ID)
		}
	})

	t.Run("Ascending", func(t *testing.T) {
		issuesCopy := make([]*redmine.Issue, len(issues))
		copy(issuesCopy, issues)

		sorter := &UpdatedOnSorter{Desc: false}
		sorter.Sort(issuesCopy)

		// 昇順: 古い -> 中間 -> 最新 -> nil
		if issuesCopy[0].ID != 1 {
			t.Errorf("Sort()[0].ID = %d, want 1", issuesCopy[0].ID)
		}
		if issuesCopy[1].ID != 3 {
			t.Errorf("Sort()[1].ID = %d, want 3", issuesCopy[1].ID)
		}
		if issuesCopy[2].ID != 2 {
			t.Errorf("Sort()[2].ID = %d, want 2", issuesCopy[2].ID)
		}
	})
}

func TestDueDateSorter_Sort(t *testing.T) {
	date1 := &redmine.Date{Time: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)}
	date2 := &redmine.Date{Time: time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)}
	date3 := &redmine.Date{Time: time.Date(2025, 1, 30, 0, 0, 0, 0, time.UTC)}

	issues := []*redmine.Issue{
		{ID: 1, Subject: "期日なし", DueDate: nil},
		{ID: 2, Subject: "期日最後", DueDate: date3},
		{ID: 3, Subject: "期日最初", DueDate: date1},
		{ID: 4, Subject: "期日中間", DueDate: date2},
	}

	sorter := &DueDateSorter{Desc: false}
	sorter.Sort(issues)

	// 昇順: 近い順 -> nilは最後
	if issues[0].ID != 3 {
		t.Errorf("Sort()[0].ID = %d, want 3", issues[0].ID)
	}
	if issues[1].ID != 4 {
		t.Errorf("Sort()[1].ID = %d, want 4", issues[1].ID)
	}
	if issues[2].ID != 2 {
		t.Errorf("Sort()[2].ID = %d, want 2", issues[2].ID)
	}
	if issues[3].ID != 1 {
		t.Errorf("Sort()[3].ID = %d, want 1 (nil should be last)", issues[3].ID)
	}
}

func TestPrioritySorter_Sort(t *testing.T) {
	issues := []*redmine.Issue{
		{ID: 1, Priority: redmine.IDName{ID: 2, Name: "通常"}},
		{ID: 2, Priority: redmine.IDName{ID: 5, Name: "緊急"}},
		{ID: 3, Priority: redmine.IDName{ID: 3, Name: "高め"}},
		{ID: 4, Priority: redmine.IDName{ID: 1, Name: "低"}},
	}

	t.Run("Descending (高い順)", func(t *testing.T) {
		issuesCopy := make([]*redmine.Issue, len(issues))
		copy(issuesCopy, issues)

		sorter := &PrioritySorter{Desc: true}
		sorter.Sort(issuesCopy)

		// 降順: 5 -> 3 -> 2 -> 1
		if issuesCopy[0].ID != 2 {
			t.Errorf("Sort()[0].ID = %d, want 2", issuesCopy[0].ID)
		}
		if issuesCopy[1].ID != 3 {
			t.Errorf("Sort()[1].ID = %d, want 3", issuesCopy[1].ID)
		}
		if issuesCopy[2].ID != 1 {
			t.Errorf("Sort()[2].ID = %d, want 1", issuesCopy[2].ID)
		}
		if issuesCopy[3].ID != 4 {
			t.Errorf("Sort()[3].ID = %d, want 4", issuesCopy[3].ID)
		}
	})
}

func TestIDSorter_Sort(t *testing.T) {
	issues := []*redmine.Issue{
		{ID: 42, Subject: "チケット42"},
		{ID: 10, Subject: "チケット10"},
		{ID: 99, Subject: "チケット99"},
		{ID: 5, Subject: "チケット5"},
	}

	sorter := &IDSorter{Desc: false}
	sorter.Sort(issues)

	// 昇順: 5 -> 10 -> 42 -> 99
	expectedIDs := []int{5, 10, 42, 99}
	for i, expected := range expectedIDs {
		if issues[i].ID != expected {
			t.Errorf("Sort()[%d].ID = %d, want %d", i, issues[i].ID, expected)
		}
	}
}
