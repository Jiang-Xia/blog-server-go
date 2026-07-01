// Package util 博客域通用工具。
package util

import "html"

// EscapeHTML 转义用户输入 HTML，对齐 Nest escapeHtml。
func EscapeHTML(s string) string {
	return html.EscapeString(s)
}
