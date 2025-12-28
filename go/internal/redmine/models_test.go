package redmine

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDateUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		want    time.Time
	}{
		{
			name:    "正常な日付",
			input:   `"2026-01-02"`,
			wantErr: false,
			want:    time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "null",
			input:   `null`,
			wantErr: false,
			want:    time.Time{},
		},
		{
			name:    "空文字列",
			input:   `""`,
			wantErr: false,
			want:    time.Time{},
		},
		{
			name:    "不正な形式",
			input:   `"2026/01/02"`,
			wantErr: true,
			want:    time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Date
			err := json.Unmarshal([]byte(tt.input), &d)

			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !d.Time.Equal(tt.want) {
				t.Errorf("UnmarshalJSON() = %v; want %v", d.Time, tt.want)
			}
		})
	}
}

func TestDateFormat(t *testing.T) {
	tests := []struct {
		name string
		date *Date
		want string
	}{
		{
			name: "通常の日付",
			date: &Date{Time: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)},
			want: "2026/01/02",
		},
		{
			name: "nil",
			date: nil,
			want: "----/--/--",
		},
		{
			name: "ゼロ値",
			date: &Date{},
			want: "----/--/--",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.date.Format()
			if got != tt.want {
				t.Errorf("Format() = %q; want %q", got, tt.want)
			}
		})
	}
}

func TestIssueUnmarshal(t *testing.T) {
	// RedmineのAPIレスポンスをシミュレート
	jsonData := `{
		"id": 123,
		"project": {"id": 1, "name": "テストプロジェクト"},
		"tracker": {"id": 1, "name": "バグ"},
		"status": {"id": 1, "name": "新規"},
		"priority": {"id": 2, "name": "通常"},
		"subject": "テストチケット",
		"description": "これはテストです",
		"start_date": "2026-01-01",
		"due_date": "2026-01-31",
		"assigned_to": {"id": 5, "name": "佐藤"},
		"parent": {"id": 100}
	}`

	var issue Issue
	err := json.Unmarshal([]byte(jsonData), &issue)
	if err != nil {
		t.Fatalf("Unmarshal()でエラー: %v", err)
	}

	// 検証
	if issue.ID != 123 {
		t.Errorf("ID = %d; want 123", issue.ID)
	}

	if issue.Subject != "テストチケット" {
		t.Errorf("Subject = %q; want 'テストチケット'", issue.Subject)
	}

	if issue.Status.Name != "新規" {
		t.Errorf("Status.Name = %q; want '新規'", issue.Status.Name)
	}

	if issue.AssignedTo == nil || issue.AssignedTo.Name != "佐藤" {
		t.Error("AssignedToが正しく解析されていない")
	}

	if issue.Parent == nil || issue.Parent.ID != 100 {
		t.Error("Parentが正しく解析されていない")
	}

	expectedStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if !issue.StartDate.Time.Equal(expectedStart) {
		t.Errorf("StartDate = %v; want %v", issue.StartDate.Time, expectedStart)
	}
}

func TestAPIResponseUnmarshal(t *testing.T) {
	jsonData := `{
		"issues": [
			{
				"id": 1,
				"subject": "チケット1",
				"status": {"id": 1, "name": "新規"}
			},
			{
				"id": 2,
				"subject": "チケット2",
				"status": {"id": 2, "name": "進行中"}
			}
		],
		"total_count": 150,
		"offset": 0,
		"limit": 100
	}`

	var resp APIResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("Unmarshal()でエラー: %v", err)
	}

	if resp.TotalCount != 150 {
		t.Errorf("TotalCount = %d; want 150", resp.TotalCount)
	}

	if resp.Offset != 0 {
		t.Errorf("Offset = %d; want 0", resp.Offset)
	}

	if resp.Limit != 100 {
		t.Errorf("Limit = %d; want 100", resp.Limit)
	}

	if len(resp.Issues) != 2 {
		t.Fatalf("Issues length = %d; want 2", len(resp.Issues))
	}

	if resp.Issues[0].ID != 1 {
		t.Errorf("Issues[0].ID = %d; want 1", resp.Issues[0].ID)
	}

	if resp.Issues[1].Subject != "チケット2" {
		t.Errorf("Issues[1].Subject = %q; want 'チケット2'", resp.Issues[1].Subject)
	}
}
