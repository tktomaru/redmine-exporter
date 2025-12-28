package filter

import (
	"testing"
	"time"

	"github.com/tktomaru/redmine-exporter/internal/redmine"
)

func TestNewCommentFilter(t *testing.T) {
	tests := []struct {
		name    string
		mode    string
		wantErr bool
	}{
		{
			name:    "Valid mode: all",
			mode:    "all",
			wantErr: false,
		},
		{
			name:    "Valid mode: last",
			mode:    "last",
			wantErr: false,
		},
		{
			name:    "Valid mode: n:3",
			mode:    "n:3",
			wantErr: false,
		},
		{
			name:    "Valid mode: n:10",
			mode:    "n:10",
			wantErr: false,
		},
		{
			name:    "Empty mode",
			mode:    "",
			wantErr: false,
		},
		{
			name:    "Invalid mode",
			mode:    "invalid",
			wantErr: true,
		},
		{
			name:    "Invalid n format - no number",
			mode:    "n:",
			wantErr: true,
		},
		{
			name:    "Invalid n format - not a number",
			mode:    "n:abc",
			wantErr: true,
		},
		{
			name:    "Invalid n format - zero",
			mode:    "n:0",
			wantErr: true,
		},
		{
			name:    "Invalid n format - negative",
			mode:    "n:-1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCommentFilter(tt.mode, nil, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCommentFilter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommentFilter_FilterByUser(t *testing.T) {
	journals := []redmine.Journal{
		{
			ID:    1,
			User:  redmine.IDName{ID: 1, Name: "佐藤"},
			Notes: "佐藤のコメント1",
		},
		{
			ID:    2,
			User:  redmine.IDName{ID: 2, Name: "鈴木"},
			Notes: "鈴木のコメント",
		},
		{
			ID:    3,
			User:  redmine.IDName{ID: 1, Name: "佐藤"},
			Notes: "佐藤のコメント2",
		},
	}

	cf, _ := NewCommentFilter("all", nil, "佐藤")
	result := cf.Filter(journals)

	if len(result) != 2 {
		t.Errorf("Filter() returned %d journals, want 2", len(result))
	}

	for _, j := range result {
		if j.User.Name != "佐藤" {
			t.Errorf("Filter() returned journal with user %s, want 佐藤", j.User.Name)
		}
	}
}

func TestCommentFilter_FilterByDate(t *testing.T) {
	sinceDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	journals := []redmine.Journal{
		{
			ID:        1,
			Notes:     "古いコメント",
			CreatedOn: "2025-01-10T10:00:00Z",
		},
		{
			ID:        2,
			Notes:     "新しいコメント1",
			CreatedOn: "2025-01-16T10:00:00Z",
		},
		{
			ID:        3,
			Notes:     "新しいコメント2",
			CreatedOn: "2025-01-20T10:00:00Z",
		},
	}

	cf, _ := NewCommentFilter("all", &sinceDate, "")
	result := cf.Filter(journals)

	if len(result) != 2 {
		t.Errorf("Filter() returned %d journals, want 2", len(result))
	}

	// 結果のジャーナルが全て sinceDate 以降であることを確認
	for _, j := range result {
		if j.ParsedCreatedOn == nil {
			t.Error("ParsedCreatedOn should not be nil")
			continue
		}
		if j.ParsedCreatedOn.Before(sinceDate) {
			t.Errorf("Filter() returned journal created on %v, which is before %v", j.ParsedCreatedOn.Time, sinceDate)
		}
	}
}

func TestCommentFilter_FilterByMode_Last(t *testing.T) {
	journals := []redmine.Journal{
		{
			ID:    1,
			Notes: "コメント1",
		},
		{
			ID:    2,
			Notes: "",
		},
		{
			ID:    3,
			Notes: "コメント3（最新）",
		},
	}

	cf, _ := NewCommentFilter("last", nil, "")
	result := cf.Filter(journals)

	if len(result) != 1 {
		t.Fatalf("Filter() returned %d journals, want 1", len(result))
	}

	if result[0].Notes != "コメント3（最新）" {
		t.Errorf("Filter() returned journal with notes %q, want 'コメント3（最新）'", result[0].Notes)
	}
}

func TestCommentFilter_FilterByMode_N(t *testing.T) {
	journals := []redmine.Journal{
		{ID: 1, Notes: "コメント1"},
		{ID: 2, Notes: "コメント2"},
		{ID: 3, Notes: ""},
		{ID: 4, Notes: "コメント4"},
		{ID: 5, Notes: "コメント5"},
		{ID: 6, Notes: "コメント6"},
	}

	tests := []struct {
		name string
		mode string
		want int
	}{
		{
			name: "n:2 - 最新2件",
			mode: "n:2",
			want: 2,
		},
		{
			name: "n:3 - 最新3件",
			mode: "n:3",
			want: 3,
		},
		{
			name: "n:10 - 件数より多く指定（全件）",
			mode: "n:10",
			want: 5, // Notes が空でないものは5件
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf, _ := NewCommentFilter(tt.mode, nil, "")
			result := cf.Filter(journals)

			if len(result) != tt.want {
				t.Errorf("Filter() returned %d journals, want %d", len(result), tt.want)
			}

			// 全てのジャーナルがNotesを持っていることを確認
			for _, j := range result {
				if j.Notes == "" {
					t.Error("Filter() returned journal with empty Notes")
				}
			}
		})
	}
}

func TestCommentFilter_FilterByMode_All(t *testing.T) {
	journals := []redmine.Journal{
		{ID: 1, Notes: "コメント1"},
		{ID: 2, Notes: "コメント2"},
		{ID: 3, Notes: "コメント3"},
	}

	cf, _ := NewCommentFilter("all", nil, "")
	result := cf.Filter(journals)

	if len(result) != 3 {
		t.Errorf("Filter() returned %d journals, want 3", len(result))
	}
}

func TestCommentFilter_CombinedFilters(t *testing.T) {
	sinceDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	journals := []redmine.Journal{
		{
			ID:        1,
			User:      redmine.IDName{ID: 1, Name: "佐藤"},
			Notes:     "佐藤の古いコメント",
			CreatedOn: "2025-01-10T10:00:00Z",
		},
		{
			ID:        2,
			User:      redmine.IDName{ID: 1, Name: "佐藤"},
			Notes:     "佐藤の新しいコメント1",
			CreatedOn: "2025-01-16T10:00:00Z",
		},
		{
			ID:        3,
			User:      redmine.IDName{ID: 2, Name: "鈴木"},
			Notes:     "鈴木の新しいコメント",
			CreatedOn: "2025-01-17T10:00:00Z",
		},
		{
			ID:        4,
			User:      redmine.IDName{ID: 1, Name: "佐藤"},
			Notes:     "佐藤の新しいコメント2",
			CreatedOn: "2025-01-18T10:00:00Z",
		},
	}

	// 佐藤のコメントで、sinceDate以降のものをフィルタ
	cf, _ := NewCommentFilter("all", &sinceDate, "佐藤")
	result := cf.Filter(journals)

	if len(result) != 2 {
		t.Errorf("Filter() returned %d journals, want 2", len(result))
	}

	for _, j := range result {
		if j.User.Name != "佐藤" {
			t.Errorf("Filter() returned journal with user %s, want 佐藤", j.User.Name)
		}
		if j.ParsedCreatedOn == nil || j.ParsedCreatedOn.Before(sinceDate) {
			t.Error("Filter() returned journal before sinceDate")
		}
	}
}

func TestCommentFilter_NilFilter(t *testing.T) {
	journals := []redmine.Journal{
		{ID: 1, Notes: "コメント1"},
		{ID: 2, Notes: "コメント2"},
	}

	var cf *CommentFilter = nil
	result := cf.Filter(journals)

	if len(result) != 2 {
		t.Errorf("Filter() with nil filter returned %d journals, want 2", len(result))
	}
}

func TestOnlyWithNotes(t *testing.T) {
	journals := []redmine.Journal{
		{ID: 1, Notes: "コメント1"},
		{ID: 2, Notes: ""},
		{ID: 3, Notes: "コメント3"},
		{ID: 4, Notes: ""},
		{ID: 5, Notes: "コメント5"},
	}

	result := OnlyWithNotes(journals)

	if len(result) != 3 {
		t.Errorf("OnlyWithNotes() returned %d journals, want 3", len(result))
	}

	for _, j := range result {
		if j.Notes == "" {
			t.Error("OnlyWithNotes() returned journal with empty Notes")
		}
	}
}
