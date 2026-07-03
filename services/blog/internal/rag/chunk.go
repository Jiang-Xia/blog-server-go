package rag

import (
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
)

// ChunkInput 分块结果，供 Indexer 写入 knowledge_chunk。
type ChunkInput struct {
	ChunkIndex  int
	Content     string
	HeadingPath string
	ContentType string
}

// ChunkService Markdown 结构感知分块。
type ChunkService struct {
	maxChars    int
	minChars    int
	overlapChars int
}

// NewChunkService 构造分块服务；size/overlap 来自配置，0 时用默认常量。
func NewChunkService(cfg *config.Config) *ChunkService {
	max := cfg.Rag.Chunk.Size
	if max <= 0 {
		max = ChunkMaxChars
	}
	overlap := cfg.Rag.Chunk.Overlap
	if overlap <= 0 {
		overlap = ChunkOverlapChars
	}
	return &ChunkService{maxChars: max, minChars: ChunkMinChars, overlapChars: overlap}
}

// SplitMarkdown 将 Markdown 正文切分为 RAG 块；description 非空时插入摘要块。
func (s *ChunkService) SplitMarkdown(markdown, title, description string) []ChunkInput {
	normalized := strings.ReplaceAll(markdown, "\r\n", "\n")
	normalized = strings.TrimSpace(normalized)
	if normalized == "" {
		content := "# " + title + "\n\n"
		if strings.TrimSpace(description) != "" {
			content += strings.TrimSpace(description)
		} else {
			content += "（无正文）"
		}
		return []ChunkInput{{
			ChunkIndex: 0, Content: content, HeadingPath: title, ContentType: "prose",
		}}
	}

	segments := s.parseSegments(normalized)
	var raw []parsedSegment
	for _, seg := range segments {
		if seg.contentType != "prose" {
			raw = append(raw, seg)
			continue
		}
		if len(seg.content) <= s.maxChars {
			raw = append(raw, seg)
			continue
		}
		raw = append(raw, s.splitLongProse(seg.content, seg.headingPath)...)
	}
	merged := s.mergeSmallPieces(raw)

	var chunks []ChunkInput
	hasDesc := strings.TrimSpace(description) != ""
	if hasDesc {
		chunks = append(chunks, ChunkInput{
			ChunkIndex: 0,
			Content:    "# " + title + "\n\n" + strings.TrimSpace(description),
			HeadingPath: title,
			ContentType: "prose",
		})
	}
	for i, piece := range merged {
		idx := i
		if hasDesc {
			idx = i + 1
		}
		content := strings.TrimSpace(piece.content)
		if content == "" {
			continue
		}
		hp := piece.headingPath
		if hp == "" {
			hp = title
		}
		ct := piece.contentType
		if ct == "" {
			ct = "prose"
		}
		chunks = append(chunks, ChunkInput{
			ChunkIndex: idx, Content: content, HeadingPath: hp, ContentType: ct,
		})
	}
	return chunks
}

type parsedSegment struct {
	content     string
	contentType string
	headingPath string
}

func (s *ChunkService) parseSegments(text string) []parsedSegment {
	var segments []parsedSegment
	headingStack := make([]string, 0, 4)
	var buffer []string
	bufferType := "prose"

	flush := func() {
		content := strings.TrimSpace(strings.Join(buffer, "\n"))
		if content != "" {
			hp := strings.Join(nonEmpty(headingStack), " > ")
			segments = append(segments, parsedSegment{content: content, contentType: bufferType, headingPath: hp})
		}
		buffer = nil
		bufferType = "prose"
	}

	lines := strings.Split(text, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if level, headingText, ok := parseHeading(line); ok {
			flush()
			if level <= len(headingStack) {
				headingStack = headingStack[:level-1]
			}
			for len(headingStack) < level {
				headingStack = append(headingStack, "")
			}
			headingStack[level-1] = headingText
			buffer = append(buffer, line)
			continue
		}
		if strings.HasPrefix(line, "```") {
			if bufferType == "code" && len(buffer) > 0 {
				buffer = append(buffer, line)
				flush()
				continue
			}
			if bufferType == "prose" && len(buffer) > 0 {
				flush()
			}
			bufferType = "code"
			buffer = append(buffer, line)
			for i+1 < len(lines) {
				i++
				buffer = append(buffer, lines[i])
				if strings.HasPrefix(lines[i], "```") {
					break
				}
			}
			flush()
			continue
		}
		if isTableLine(line) && (bufferType == "table" || (bufferType == "prose" && len(buffer) == 0)) {
			if bufferType == "prose" && len(buffer) > 0 {
				flush()
			}
			bufferType = "table"
			buffer = append(buffer, line)
			for i+1 < len(lines) && isTableLine(lines[i+1]) {
				i++
				buffer = append(buffer, lines[i])
			}
			flush()
			continue
		}
		if bufferType != "prose" && len(buffer) > 0 {
			flush()
		}
		bufferType = "prose"
		buffer = append(buffer, line)
	}
	flush()
	return segments
}

func parseHeading(line string) (level int, text string, ok bool) {
	if len(line) < 3 || line[0] != '#' {
		return 0, "", false
	}
	level = 0
	for level < len(line) && level < 4 && line[level] == '#' {
		level++
	}
	if level == 0 || level > 4 || level >= len(line) || line[level] != ' ' {
		return 0, "", false
	}
	return level, strings.TrimSpace(line[level+1:]), true
}

func isTableLine(line string) bool {
	t := strings.TrimSpace(line)
	return len(t) >= 2 && t[0] == '|' && t[len(t)-1] == '|'
}

func (s *ChunkService) splitLongProse(text, headingPath string) []parsedSegment {
	sections := splitByHeadings(text)
	var pieces []parsedSegment
	for _, section := range sections {
		if len(section) <= s.maxChars {
			pieces = append(pieces, parsedSegment{content: section, contentType: "prose", headingPath: headingPath})
			continue
		}
		pieces = append(pieces, s.splitLongText(section, headingPath)...)
	}
	return pieces
}

func splitByHeadings(text string) []string {
	lines := strings.Split(text, "\n")
	var sections []string
	var current []string
	for _, line := range lines {
		if len(line) >= 2 && line[0] == '#' && (line[1] == ' ' || line[1] == '#') && len(current) > 0 {
			sections = append(sections, strings.TrimSpace(strings.Join(current, "\n")))
			current = []string{line}
		} else {
			current = append(current, line)
		}
	}
	if len(current) > 0 {
		sections = append(sections, strings.TrimSpace(strings.Join(current, "\n")))
	}
	return filterNonEmpty(sections)
}

func (s *ChunkService) splitLongText(text, headingPath string) []parsedSegment {
	paragraphs := filterNonEmpty(strings.Split(text, "\n\n"))
	var chunks []parsedSegment
	buf := ""
	for _, para := range paragraphs {
		if len(para) > s.maxChars {
			if buf != "" {
				chunks = append(chunks, parsedSegment{content: buf, contentType: "prose", headingPath: headingPath})
				buf = ""
			}
			chunks = append(chunks, s.splitParagraph(para, headingPath)...)
			continue
		}
		candidate := para
		if buf != "" {
			candidate = buf + "\n\n" + para
		}
		if len(candidate) > s.maxChars {
			if buf != "" {
				chunks = append(chunks, parsedSegment{content: buf, contentType: "prose", headingPath: headingPath})
			}
			buf = para
		} else {
			buf = candidate
		}
	}
	if buf != "" {
		chunks = append(chunks, parsedSegment{content: buf, contentType: "prose", headingPath: headingPath})
	}
	return chunks
}

func (s *ChunkService) splitParagraph(text, headingPath string) []parsedSegment {
	var result []parsedSegment
	start := 0
	for start < len(text) {
		end := start + s.maxChars
		if end > len(text) {
			end = len(text)
		}
		if end < len(text) {
			slice := text[start:end]
			lastBreak := maxInt(strings.LastIndex(slice, "\n"), strings.LastIndex(slice, "。"),
				strings.LastIndex(slice, "."), strings.LastIndex(slice, "；"))
			if lastBreak > s.minChars/2 {
				end = start + lastBreak + 1
			}
		}
		piece := strings.TrimSpace(text[start:end])
		if piece != "" {
			result = append(result, parsedSegment{content: piece, contentType: "prose", headingPath: headingPath})
		}
		overlap := s.overlapChars
		if overlap > len(piece) {
			overlap = len(piece)
		}
		next := end - overlap
		if next <= start {
			next = start + 1
		}
		start = next
		if end >= len(text) {
			break
		}
	}
	return result
}

func (s *ChunkService) mergeSmallPieces(pieces []parsedSegment) []parsedSegment {
	var merged []parsedSegment
	var buffer *parsedSegment
	for _, piece := range pieces {
		if piece.contentType != "prose" {
			if buffer != nil {
				merged = append(merged, *buffer)
				buffer = nil
			}
			merged = append(merged, piece)
			continue
		}
		if buffer == nil {
			p := piece
			buffer = &p
			continue
		}
		candidate := buffer.content + "\n\n" + piece.content
		if len(candidate) <= s.maxChars {
			buffer.content = candidate
			if piece.headingPath != "" {
				buffer.headingPath = piece.headingPath
			}
		} else {
			merged = append(merged, *buffer)
			p := piece
			buffer = &p
		}
	}
	if buffer != nil {
		merged = append(merged, *buffer)
	}
	return merged
}

func nonEmpty(ss []string) []string {
	var out []string
	for _, s := range ss {
		if strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
	}
	return out
}

func filterNonEmpty(ss []string) []string {
	var out []string
	for _, s := range ss {
		if strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
	}
	return out
}

func maxInt(vals ...int) int {
	m := -1
	for _, v := range vals {
		if v > m {
			m = v
		}
	}
	return m
}
