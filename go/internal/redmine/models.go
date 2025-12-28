package redmine

import (
	"strings"
	"time"
)

// APIResponse はRedmine API /issues.jsonのレスポンス
type APIResponse struct {
	Issues     []*Issue `json:"issues"`
	TotalCount int      `json:"total_count"`
	Offset     int      `json:"offset"`
	Limit      int      `json:"limit"`
}

// Issue はRedmineのチケット
type Issue struct {
	ID          int        `json:"id"`
	Project     IDName     `json:"project"`
	Tracker     IDName     `json:"tracker"`
	Status      IDName     `json:"status"`
	Priority    IDName     `json:"priority"`
	Subject     string     `json:"subject"`
	Description string     `json:"description"`
	StartDate   *Date      `json:"start_date"`
	DueDate     *Date      `json:"due_date"`
	AssignedTo  *IDName    `json:"assigned_to"`
	Parent      *IssueRef  `json:"parent"`

	// 処理用フィールド（APIレスポンスには含まれない）
	CleanedSubject string   `json:"-"`
	Summary        string   `json:"-"`
	Children       []*Issue `json:"-"`
}

// IDName はID+名前を持つRedmineオブジェクト
type IDName struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// IssueRef は親チケットへの参照
type IssueRef struct {
	ID int `json:"id"`
}

// Date はRedmineの日付フォーマット（YYYY-MM-DD）
type Date struct {
	time.Time
}

// UnmarshalJSON はカスタム日付パース
func (d *Date) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "null" || s == `""` {
		return nil
	}
	// 引用符を削除
	s = strings.Trim(s, `"`)
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

// Format は日付を指定フォーマットで返す（VBA版互換）
func (d *Date) Format() string {
	if d == nil || d.Time.IsZero() {
		return "----/--/--"
	}
	return d.Time.Format("2006/01/02")
}
