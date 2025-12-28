package filter

import (
	"fmt"
	"time"
)

// WeekCalculator は週の期間を計算する
type WeekCalculator struct {
	weekStart time.Weekday   // 週の起点（Monday or Sunday）
	location  *time.Location // タイムゾーン
}

// NewWeekCalculator は新しいWeekCalculatorを作成
func NewWeekCalculator(weekStart string, timezone string) (*WeekCalculator, error) {
	// タイムゾーンを読み込み
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("タイムゾーン読み込みエラー: %w", err)
	}

	// 週の起点を設定
	var ws time.Weekday
	switch weekStart {
	case "mon":
		ws = time.Monday
	case "sun":
		ws = time.Sunday
	default:
		return nil, fmt.Errorf("不正な週の起点: %s (mon or sun)", weekStart)
	}

	return &WeekCalculator{
		weekStart: ws,
		location:  loc,
	}, nil
}

// GetWeekRange は週番号から期間を計算
// spec: "last", "this", "YYYY-WW" (例: "2025-01")
func (wc *WeekCalculator) GetWeekRange(spec string) (start, end time.Time, err error) {
	now := time.Now().In(wc.location)

	switch spec {
	case "last":
		// 先週の開始日と終了日
		return wc.getLastWeek(now)
	case "this":
		// 今週の開始日と終了日
		return wc.getThisWeek(now)
	default:
		// YYYY-WW 形式
		return wc.parseWeekSpec(spec)
	}
}

// getThisWeek は今週の期間を返す
func (wc *WeekCalculator) getThisWeek(now time.Time) (start, end time.Time, err error) {
	// 今週の開始日（weekStartに合わせる）
	daysFromWeekStart := (int(now.Weekday()) - int(wc.weekStart) + 7) % 7
	start = now.AddDate(0, 0, -daysFromWeekStart)
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, wc.location)

	// 今週の終了日（開始日 + 6日の23:59:59）
	end = start.AddDate(0, 0, 6)
	end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, wc.location)

	return start, end, nil
}

// getLastWeek は先週の期間を返す
func (wc *WeekCalculator) getLastWeek(now time.Time) (start, end time.Time, err error) {
	// 今週の開始日を取得
	thisStart, _, _ := wc.getThisWeek(now)

	// 先週の開始日（今週の開始日 - 7日）
	start = thisStart.AddDate(0, 0, -7)

	// 先週の終了日（開始日 + 6日の23:59:59）
	end = start.AddDate(0, 0, 6)
	end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, wc.location)

	return start, end, nil
}

// parseWeekSpec は週番号（YYYY-WW形式）から期間を計算
func (wc *WeekCalculator) parseWeekSpec(spec string) (start, end time.Time, err error) {
	// "2025-01" 形式をパース
	var year, week int
	_, err = fmt.Sscanf(spec, "%d-%d", &year, &week)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("週番号の形式エラー: %s (YYYY-WW形式で指定)", spec)
	}

	// 年の1月1日を基準に週番号から日付を計算
	jan1 := time.Date(year, 1, 1, 0, 0, 0, 0, wc.location)

	// 1月1日が何曜日か
	jan1Weekday := int(jan1.Weekday())

	// weekStartまでの日数
	daysToWeekStart := (int(wc.weekStart) - jan1Weekday + 7) % 7

	// 第1週の開始日
	firstWeekStart := jan1.AddDate(0, 0, daysToWeekStart)
	if daysToWeekStart > 3 {
		// 1月1日が週の後半なら、翌週を第1週とする（ISO 8601準拠）
		firstWeekStart = firstWeekStart.AddDate(0, 0, -7)
	}

	// 指定週の開始日
	start = firstWeekStart.AddDate(0, 0, (week-1)*7)

	// 終了日（開始日 + 6日の23:59:59）
	end = start.AddDate(0, 0, 6)
	end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, wc.location)

	return start, end, nil
}
