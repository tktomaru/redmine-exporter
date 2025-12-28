package filter

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tktomaru/redmine-exporter/internal/redmine"
)

// CommentFilter はコメントのフィルタリング条件
type CommentFilter struct {
	Mode      string     // "last", "all", "n:3" (最新N件)
	SinceDate *time.Time // この日時以降のコメントのみ
	ByUser    string     // 特定ユーザーのコメントのみ
}

// NewCommentFilter は新しいCommentFilterを作成
func NewCommentFilter(mode string, sinceDate *time.Time, byUser string) (*CommentFilter, error) {
	// モードの検証
	if mode != "" && mode != "all" && mode != "last" {
		// "n:3" のような形式もチェック
		if !strings.HasPrefix(mode, "n:") {
			return nil, fmt.Errorf("不正なコメントモード: %s (all, last, n:N)", mode)
		}
		// "n:3" から数値を抽出して検証
		parts := strings.Split(mode, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("不正なコメントモード形式: %s (n:N)", mode)
		}
		n, err := strconv.Atoi(parts[1])
		if err != nil || n <= 0 {
			return nil, fmt.Errorf("不正なコメント数: %s (正の整数を指定)", parts[1])
		}
	}

	return &CommentFilter{
		Mode:      mode,
		SinceDate: sinceDate,
		ByUser:    byUser,
	}, nil
}

// Filter はジャーナルをフィルタリング
func (cf *CommentFilter) Filter(journals []redmine.Journal) []redmine.Journal {
	if cf == nil {
		return journals
	}

	// 各ジャーナルのCreatedOnをパース
	for i := range journals {
		_ = journals[i].ParseCreatedOn()
	}

	// 1. ユーザーフィルタ
	filtered := cf.filterByUser(journals)

	// 2. 日時フィルタ
	filtered = cf.filterByDate(filtered)

	// 3. モード別フィルタ（最新N件など）
	filtered = cf.filterByMode(filtered)

	return filtered
}

// filterByUser はユーザーでフィルタリング
func (cf *CommentFilter) filterByUser(journals []redmine.Journal) []redmine.Journal {
	if cf.ByUser == "" {
		return journals
	}

	var result []redmine.Journal
	for _, j := range journals {
		if j.User.Name == cf.ByUser {
			result = append(result, j)
		}
	}
	return result
}

// filterByDate は日時でフィルタリング
func (cf *CommentFilter) filterByDate(journals []redmine.Journal) []redmine.Journal {
	if cf.SinceDate == nil {
		return journals
	}

	var result []redmine.Journal
	for _, j := range journals {
		if j.ParsedCreatedOn != nil && !j.ParsedCreatedOn.Before(*cf.SinceDate) {
			result = append(result, j)
		}
	}
	return result
}

// filterByMode はモード別にフィルタリング
func (cf *CommentFilter) filterByMode(journals []redmine.Journal) []redmine.Journal {
	if cf.Mode == "" || cf.Mode == "all" {
		return journals
	}

	// "last" - 最新の1件のみ
	if cf.Mode == "last" {
		if len(journals) == 0 {
			return journals
		}
		// 最新のコメント（Notes が空でないもの）を探す
		for i := len(journals) - 1; i >= 0; i-- {
			if journals[i].Notes != "" {
				return []redmine.Journal{journals[i]}
			}
		}
		return []redmine.Journal{}
	}

	// "n:3" - 最新のN件
	if strings.HasPrefix(cf.Mode, "n:") {
		parts := strings.Split(cf.Mode, ":")
		n, _ := strconv.Atoi(parts[1])

		// Notes が空でないジャーナルのみ抽出
		var withNotes []redmine.Journal
		for _, j := range journals {
			if j.Notes != "" {
				withNotes = append(withNotes, j)
			}
		}

		// 最新N件を取得
		if len(withNotes) <= n {
			return withNotes
		}
		return withNotes[len(withNotes)-n:]
	}

	return journals
}

// OnlyWithNotes はNotesが空でないジャーナルのみを返す
func OnlyWithNotes(journals []redmine.Journal) []redmine.Journal {
	var result []redmine.Journal
	for _, j := range journals {
		if j.Notes != "" {
			result = append(result, j)
		}
	}
	return result
}
