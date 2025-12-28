package filter

import (
	"testing"
	"time"
)

func TestNewWeekCalculator(t *testing.T) {
	tests := []struct {
		name      string
		weekStart string
		timezone  string
		wantStart time.Weekday
		wantErr   bool
	}{
		{
			name:      "Sunday start with Asia/Tokyo",
			weekStart: "sun",
			timezone:  "Asia/Tokyo",
			wantStart: time.Sunday,
			wantErr:   false,
		},
		{
			name:      "Monday start with Asia/Tokyo",
			weekStart: "mon",
			timezone:  "Asia/Tokyo",
			wantStart: time.Monday,
			wantErr:   false,
		},
		{
			name:      "Tuesday start with Asia/Tokyo",
			weekStart: "tue",
			timezone:  "Asia/Tokyo",
			wantStart: time.Tuesday,
			wantErr:   false,
		},
		{
			name:      "Wednesday start with Asia/Tokyo",
			weekStart: "wed",
			timezone:  "Asia/Tokyo",
			wantStart: time.Wednesday,
			wantErr:   false,
		},
		{
			name:      "Thursday start with Asia/Tokyo",
			weekStart: "thu",
			timezone:  "Asia/Tokyo",
			wantStart: time.Thursday,
			wantErr:   false,
		},
		{
			name:      "Friday start with Asia/Tokyo",
			weekStart: "fri",
			timezone:  "Asia/Tokyo",
			wantStart: time.Friday,
			wantErr:   false,
		},
		{
			name:      "Saturday start with Asia/Tokyo",
			weekStart: "sat",
			timezone:  "Asia/Tokyo",
			wantStart: time.Saturday,
			wantErr:   false,
		},
		{
			name:      "Invalid week start",
			weekStart: "invalid",
			timezone:  "Asia/Tokyo",
			wantErr:   true,
		},
		{
			name:      "Invalid timezone",
			weekStart: "mon",
			timezone:  "Invalid/Zone",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wc, err := NewWeekCalculator(tt.weekStart, tt.timezone)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWeekCalculator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && wc.weekStart != tt.wantStart {
				t.Errorf("NewWeekCalculator() weekStart = %v, want %v", wc.weekStart, tt.wantStart)
			}
		})
	}
}

func TestGetWeekRange_Last(t *testing.T) {
	// 固定日時でテスト: 2025年1月15日（水曜日）
	location, _ := time.LoadLocation("Asia/Tokyo")

	tests := []struct {
		name      string
		weekStart string
		now       time.Time
		wantDays  int // 先週の月曜日が何日か
	}{
		{
			name:      "Monday start - last week from Wednesday",
			weekStart: "mon",
			now:       time.Date(2025, 1, 15, 10, 0, 0, 0, location), // 水曜日
			wantDays:  6,                                              // 1月6日（先週の月曜）
		},
		{
			name:      "Sunday start - last week from Wednesday",
			weekStart: "sun",
			now:       time.Date(2025, 1, 15, 10, 0, 0, 0, location), // 水曜日
			wantDays:  5,                                              // 1月5日（先週の日曜）
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wc, _ := NewWeekCalculator(tt.weekStart, "Asia/Tokyo")
			// 固定された now を使うため直接 getLastWeek を呼ぶ
			start, end, err := wc.getLastWeek(tt.now)
			if err != nil {
				t.Fatalf("getLastWeek() error = %v", err)
			}

			// 開始日のチェック
			if start.Day() != tt.wantDays {
				t.Errorf("getLastWeek() start day = %d, want %d", start.Day(), tt.wantDays)
			}

			// 期間が7日間であることを確認
			duration := end.Sub(start)
			expectedDuration := 7*24*time.Hour - time.Second // 23:59:59まで
			if duration != expectedDuration {
				t.Errorf("getLastWeek() duration = %v, want %v", duration, expectedDuration)
			}

			// 開始日が00:00:00であることを確認
			if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
				t.Errorf("getLastWeek() start time = %02d:%02d:%02d, want 00:00:00",
					start.Hour(), start.Minute(), start.Second())
			}

			// 終了日が23:59:59であることを確認
			if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
				t.Errorf("getLastWeek() end time = %02d:%02d:%02d, want 23:59:59",
					end.Hour(), end.Minute(), end.Second())
			}
		})
	}
}

func TestGetWeekRange_This(t *testing.T) {
	location, _ := time.LoadLocation("Asia/Tokyo")

	tests := []struct {
		name      string
		weekStart string
		now       time.Time
		wantDays  int // 今週の開始日が何日か
	}{
		{
			name:      "Monday start - this week from Wednesday",
			weekStart: "mon",
			now:       time.Date(2025, 1, 15, 10, 0, 0, 0, location), // 水曜日
			wantDays:  13,                                             // 1月13日（今週の月曜）
		},
		{
			name:      "Sunday start - this week from Wednesday",
			weekStart: "sun",
			now:       time.Date(2025, 1, 15, 10, 0, 0, 0, location), // 水曜日
			wantDays:  12,                                             // 1月12日（今週の日曜）
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wc, _ := NewWeekCalculator(tt.weekStart, "Asia/Tokyo")
			start, end, err := wc.getThisWeek(tt.now)
			if err != nil {
				t.Fatalf("getThisWeek() error = %v", err)
			}

			if start.Day() != tt.wantDays {
				t.Errorf("getThisWeek() start day = %d, want %d", start.Day(), tt.wantDays)
			}

			// 期間が7日間であることを確認
			duration := end.Sub(start)
			expectedDuration := 7*24*time.Hour - time.Second
			if duration != expectedDuration {
				t.Errorf("getThisWeek() duration = %v, want %v", duration, expectedDuration)
			}
		})
	}
}

func TestParseWeekSpec(t *testing.T) {
	tests := []struct {
		name      string
		weekStart string
		spec      string
		wantMonth time.Month
		wantDay   int
		wantErr   bool
	}{
		{
			name:      "Valid week spec 2025-01",
			weekStart: "mon",
			spec:      "2025-01",
			wantMonth: time.December,
			wantDay:   30, // 2025年の第1週（月曜起点）は2024年12月30日から（ISO 8601）
			wantErr:   false,
		},
		{
			name:      "Valid week spec 2025-02",
			weekStart: "mon",
			spec:      "2025-02",
			wantMonth: time.January,
			wantDay:   6, // 第2週
			wantErr:   false,
		},
		{
			name:      "Invalid format",
			weekStart: "mon",
			spec:      "2025/01",
			wantErr:   true,
		},
		{
			name:      "Invalid format - no dash",
			weekStart: "mon",
			spec:      "202501",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wc, _ := NewWeekCalculator(tt.weekStart, "Asia/Tokyo")
			start, end, err := wc.parseWeekSpec(tt.spec)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseWeekSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if start.Month() != tt.wantMonth || start.Day() != tt.wantDay {
				t.Errorf("parseWeekSpec() start = %v-%02d, want %v-%02d",
					start.Month(), start.Day(), tt.wantMonth, tt.wantDay)
			}

			// 期間が7日間であることを確認
			duration := end.Sub(start)
			expectedDuration := 7*24*time.Hour - time.Second
			if duration != expectedDuration {
				t.Errorf("parseWeekSpec() duration = %v, want %v", duration, expectedDuration)
			}
		})
	}
}

func TestGetWeekRange_Integration(t *testing.T) {
	wc, err := NewWeekCalculator("mon", "Asia/Tokyo")
	if err != nil {
		t.Fatalf("NewWeekCalculator() failed: %v", err)
	}

	tests := []struct {
		name    string
		spec    string
		wantErr bool
	}{
		{
			name:    "last week",
			spec:    "last",
			wantErr: false,
		},
		{
			name:    "this week",
			spec:    "this",
			wantErr: false,
		},
		{
			name:    "specific week 2025-01",
			spec:    "2025-01",
			wantErr: false,
		},
		{
			name:    "invalid spec",
			spec:    "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := wc.GetWeekRange(tt.spec)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetWeekRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// 基本的な検証
			if start.IsZero() || end.IsZero() {
				t.Error("GetWeekRange() returned zero time")
			}

			if !start.Before(end) {
				t.Errorf("GetWeekRange() start (%v) should be before end (%v)", start, end)
			}

			// 開始時刻が00:00:00であることを確認
			if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
				t.Errorf("GetWeekRange() start time should be 00:00:00, got %02d:%02d:%02d",
					start.Hour(), start.Minute(), start.Second())
			}

			// 終了時刻が23:59:59であることを確認
			if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
				t.Errorf("GetWeekRange() end time should be 23:59:59, got %02d:%02d:%02d",
					end.Hour(), end.Minute(), end.Second())
			}
		})
	}
}
