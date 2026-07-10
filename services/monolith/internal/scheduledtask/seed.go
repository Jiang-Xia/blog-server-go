// Package scheduledtask 定时任务定义、调度与运维 API（Plan 12）。
package scheduledtask

// SeedTask 内置种子任务定义（对齐 Nest SEED_TASKS）。
type SeedTask struct {
	Name        string
	Description string
	Cron        string
	CronHuman   string
	SortOrder   int
}

// SeedTasks 8 个内置业务 cron（6 段含秒）。
var SeedTasks = []SeedTask{
	{Name: "daily_interaction_notify", Description: "每日互动通知", Cron: "0 0 10 * * *", CronHuman: "每天 10:00", SortOrder: 1},
	{Name: "scheduled_publish", Description: "文章定时发布", Cron: "0 * * * * *", CronHuman: "每分钟", SortOrder: 2},
	{Name: "monthly_report", Description: "每月博客月报", Cron: "0 0 10 1 * *", CronHuman: "每月1号 10:00", SortOrder: 3},
	{Name: "link_health_check", Description: "友链健康检测", Cron: "0 0 3 * * 1", CronHuman: "每周一 03:00", SortOrder: 4},
	{Name: "expired_data_cleanup", Description: "过期数据清理", Cron: "0 30 2 * * *", CronHuman: "每天 02:30", SortOrder: 5},
	{Name: "database_backup", Description: "数据库备份", Cron: "0 0 4 * * *", CronHuman: "每天 04:00", SortOrder: 6},
	{Name: "draft_cleanup", Description: "草稿90天自动清理", Cron: "0 0 5 * * *", CronHuman: "每天 05:00", SortOrder: 7},
	{Name: "sensitive_word_alert", Description: "敏感词命中告警", Cron: "0 0 9 * * *", CronHuman: "每天 09:00", SortOrder: 8},
}

const maxTasks = 100
