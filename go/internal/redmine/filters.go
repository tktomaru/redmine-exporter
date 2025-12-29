package redmine

import (
	"net/url"
	"time"
)

// DateFilter は日時範囲でのフィルタリング条件
type DateFilter struct {
	Field string    // "updated_on", "created_on", "start_date", "due_date"
	Start time.Time // 開始日時
	End   time.Time // 終了日時
}

// ToQueryString はRedmine APIのクエリパラメータ文字列を生成
// 例: updated_on>=2025-01-01&updated_on<=2025-01-07
func (df *DateFilter) ToQueryString() string {
	fb := NewFilterBuilder()
	fb.AddDateRange(df.Field, df.Start, df.End)
	return fb.Build()
}

// FilterBuilder はRedmine APIのクエリパラメータを構築する
type FilterBuilder struct {
	params url.Values
}

// NewFilterBuilder は新しいFilterBuilderを作成
func NewFilterBuilder() *FilterBuilder {
	return &FilterBuilder{
		params: url.Values{},
	}
}

// AddDateRange は日時範囲フィルタを追加
// Redmine REST API: field=><YYYY-MM-DD|YYYY-MM-DD（範囲）
// 例: created_on=%3E%3C2012-03-01|2012-03-07 :contentReference[oaicite:3]{index=3}
func (fb *FilterBuilder) AddDateRange(field string, start, end time.Time) {
	startStr := start.Format("2006-01-02")

	// end がゼロなら「以降」だけ（>=）
	if end.IsZero() {
		fb.params.Set(field, ">="+startStr)
		return
	}

	endStr := end.Format("2006-01-02")

	// もし start > end を許すなら入れ替え
	if start.After(end) {
		startStr, endStr = endStr, startStr
	}

	// 範囲指定は 1つの値にまとめる（同じキーを2回にしない）
	fb.params.Set(field, "><"+startStr+"|"+endStr)
}

// Build はクエリパラメータ文字列を返す
func (fb *FilterBuilder) Build() string {
	return fb.params.Encode()
}
