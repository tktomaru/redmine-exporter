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
	Journals    []Journal  `json:"journals"`

	// 処理用フィールド（APIレスポンスには含まれない）
	CleanedSubject string            `json:"-"`
	Summary        string            `json:"-"`
	ExtractedTags  map[string]string `json:"-"` // タグ名 -> 抽出内容
	Children       []*Issue          `json:"-"`
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

// Journal はチケットのコメント（更新履歴）
type Journal struct {
	ID        int     `json:"id"`
	User      IDName  `json:"user"`
	Notes     string  `json:"notes"`
	CreatedOn string  `json:"created_on"`
	Details   []JournalDetail `json:"details"`
}

// JournalDetail はジャーナルの変更詳細
type JournalDetail struct {
	Property string `json:"property"`
	Name     string `json:"name"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
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
