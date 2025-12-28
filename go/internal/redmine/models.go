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
	UpdatedOn   *DateTime  `json:"updated_on"` // 更新日時（週報機能用）
	CreatedOn   *DateTime  `json:"created_on"` // 作成日時（週報機能用）

	// 処理用フィールド（APIレスポンスには含まれない）
	CleanedSubject string              `json:"-"`
	Summary        string              `json:"-"`
	ExtractedTags  map[string][]string `json:"-"` // タグ名 -> 抽出内容の配列（複数値対応）
	Children       []*Issue            `json:"-"`
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

	// 処理用フィールド（APIレスポンスには含まれない）
	ParsedCreatedOn *DateTime `json:"-"` // パース済みの作成日時（コメントフィルタ用）
}

// ParseCreatedOn はCreatedOn文字列をパースしてParsedCreatedOnに設定
func (j *Journal) ParseCreatedOn() error {
	if j.CreatedOn == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, j.CreatedOn)
	if err != nil {
		return err
	}
	j.ParsedCreatedOn = &DateTime{Time: t}
	return nil
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

// DateTime はRedmineのタイムスタンプフォーマット（ISO 8601）
type DateTime struct {
	time.Time
}

// UnmarshalJSON は ISO 8601形式のタイムスタンプをパース
func (dt *DateTime) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "null" || s == `""` {
		return nil
	}
	// 引用符を削除
	s = strings.Trim(s, `"`)
	// ISO 8601形式でパース (例: "2025-01-15T10:30:45Z")
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	dt.Time = t
	return nil
}

// Format は日時を指定フォーマットで返す
func (dt *DateTime) Format() string {
	if dt == nil || dt.Time.IsZero() {
		return ""
	}
	return dt.Time.Format("2006/01/02 15:04:05")
}
