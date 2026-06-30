// Package timeutil 提供时间格式化等通用工具。
package timeutil

import "time"

const DefaultLayout = "2006-01-02 15:04:05"

// Format 将时间格式化为默认字符串。
func Format(t time.Time) string {
	return t.Format(DefaultLayout)
}

// Now 返回当前本地时间。
func Now() time.Time {
	return time.Now()
}
