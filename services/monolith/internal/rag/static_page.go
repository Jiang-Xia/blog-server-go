package rag

import (
	"embed"
	"fmt"
)

//go:embed content/pages/*.md
var staticPageMarkdown embed.FS

// StaticPageDef RAG 静态页索引元数据（须与 Nuxt 对应页面人工同步）。
type StaticPageDef struct {
	Slug         string
	Title        string
	URL          string
	Category     string
	Tags         []string
	Description  string
	MarkdownFile string
	SourceLabel  string
}

// RAGStaticPages 纳入 RAG 的静态页列表。
var RAGStaticPages = []StaticPageDef{
	{
		Slug: "features/index", Title: "站点特性概览", URL: "/features",
		Category: "站点说明", Tags: []string{"特性", "功能", "RPG", "工具箱"},
		Description: "博客核心功能模块与 RPG 冒险体系概览。",
		MarkdownFile: "features-index.md", SourceLabel: "站点特性页",
	},
	{
		Slug: "features/rpg-guide", Title: "博客 RPG 冒险攻略", URL: "/features/rpg-guide",
		Category: "RPG 攻略", Tags: []string{"RPG", "签到", "任务", "抽奖", "排行榜", "钻石"},
		Description: "从签到升级到赛季排行的完整 RPG 玩法攻略。",
		MarkdownFile: "rpg-guide.md", SourceLabel: "RPG 攻略页",
	},
	{
		Slug: "tools/guide", Title: "工具箱说明", URL: "/tool",
		Category: "工具箱", Tags: []string{"工具", "加密", "PDF", "AI摘要", "WebRTC"},
		Description: "站内 14+ 在线工具的用途与入口路径说明。",
		MarkdownFile: "tools-guide.md", SourceLabel: "工具箱说明",
	},
}

// LoadStaticPageMarkdown 读取嵌入的 Markdown 正文。
func LoadStaticPageMarkdown(def StaticPageDef) (string, error) {
	path := "content/pages/" + def.MarkdownFile
	b, err := staticPageMarkdown.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read static page %s: %w", def.MarkdownFile, err)
	}
	return string(b), nil
}
