// Package util RPG 子模块通用工具。
package util

import "time"

// SpecialQuestDate 特殊/一次性任务的固定 questDate 键（对齐 Nest 2000-01-01）。
var SpecialQuestDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)

// TodayQuestDate 返回本地时区当日零点。
func TodayQuestDate() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// WeekQuestDate 返回当周周一零点（周日视为上一周第 7 天）。
func WeekQuestDate() time.Time {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	daysSinceMonday := weekday - 1
	monday := now.AddDate(0, 0, -daysSinceMonday)
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
}

// QuestDateKey 按任务类型返回 progress 表的 questDate 键。
func QuestDateKey(questType, questSubtype string) time.Time {
	if questType == "special" || questSubtype == "special" {
		return SpecialQuestDate
	}
	if questType == "weekly" || questSubtype == "weekly" {
		return WeekQuestDate()
	}
	return TodayQuestDate()
}

// SecondsUntilMidnight 距离本地次日零点的秒数（含 60 秒缓冲）。
func SecondsUntilMidnight() int {
	now := time.Now()
	end := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	sec := int(end.Sub(now).Seconds()) + 60
	if sec < 60 {
		return 60
	}
	return sec
}
