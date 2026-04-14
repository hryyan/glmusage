package timeutil

import "time"

// Last24Hours 返回最近 24 小时的时间范围，用于 API 查询。
// start: 昨天同一时刻，分钟归零
// end: 当前时刻，分钟设为 59
func Last24Hours() (start, end time.Time) {
	now := time.Now()
	start = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
	start = start.Add(-24 * time.Hour)
	end = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 59, 0, 0, now.Location())
	return
}

// FormatForAPI 将 time.Time 格式化为 API 要求的格式 "2006-01-02 15:04:05"
func FormatForAPI(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
