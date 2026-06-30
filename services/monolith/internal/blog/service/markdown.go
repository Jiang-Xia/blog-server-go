// Package service 提供 Markdown 安全渲染（goldmark），对照 Nest marked 安全配置防 XSS。
package service

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var md goldmark.Markdown

func init() {
	md = goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(), // 与 Nest md-editor 客户端渲染一致；生产 contentHtml 由客户端传入
		),
	)
}

// RenderMarkdown 将 Markdown 转为 HTML；contentHtml 为空时服务端补渲染。
func RenderMarkdown(source string) (string, error) {
	if source == "" {
		return "", nil
	}
	var buf bytes.Buffer
	if err := md.Convert([]byte(source), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
