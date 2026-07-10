package rag

import "strings"

// BuildRagEmbedText 拼装 embedding API 输入。
func BuildRagEmbedText(title, content, description, category, headingPath, sourceLabel string, tags []string) string {
	lines := []string{"标题：" + title}
	if sourceLabel != "" {
		lines = append(lines, "来源："+sourceLabel)
	}
	if category != "" {
		lines = append(lines, "分类："+category)
	}
	if len(tags) > 0 {
		lines = append(lines, "标签："+strings.Join(tags, "、"))
	}
	if headingPath != "" {
		lines = append(lines, "章节："+headingPath)
	}
	if strings.TrimSpace(description) != "" {
		lines = append(lines, "摘要："+strings.TrimSpace(description))
	}
	lines = append(lines, "", content)
	return strings.Join(lines, "\n")
}

// BuildRagSearchText 写入 knowledge_chunk.search_text。
func BuildRagSearchText(title, content, category, headingPath string, tags []string) string {
	parts := []string{title, category}
	if len(tags) > 0 {
		parts = append(parts, strings.Join(tags, " "))
	}
	if headingPath != "" {
		parts = append(parts, headingPath)
	}
	parts = append(parts, content)
	var out []string
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, "\n")
}

// BuildEmbedTextsFromChunks 批量将分块转为 embedding 文本。
func BuildEmbedTextsFromChunks(title string, pieces []ChunkInput, description, category, sourceLabel string, tags []string) []string {
	out := make([]string, len(pieces))
	for i, piece := range pieces {
		out[i] = BuildRagEmbedText(title, piece.Content, description, category, piece.HeadingPath, sourceLabel, tags)
	}
	return out
}
