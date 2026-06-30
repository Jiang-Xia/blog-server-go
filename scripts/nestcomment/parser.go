// Package nestcomment 从 blog-server TypeORM 实体提取表/字段中文说明，供 Ent schema 生成使用。
package nestcomment

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var (
	reEntityString  = regexp.MustCompile(`@Entity\s*\(\s*['"]([^'"]+)['"]`)
	reEntityObject  = regexp.MustCompile(`@Entity\s*\(\s*\{[^}]*name:\s*['"]([^'"]+)['"]`)
	reClassName     = regexp.MustCompile(`export\s+class\s+(\w+)`)
	reExtendsBase   = regexp.MustCompile(`export\s+class\s+\w+\s+extends\s+BaseModel`)
	reProperty      = regexp.MustCompile(`^(\w+)\??:\s`)
	reDesc          = regexp.MustCompile(`description:\s*['"]([^'"]+)['"]`)
	reColumnComment = regexp.MustCompile(`comment:\s*['"]([^'"]+)['"]`)
	reLineComment   = regexp.MustCompile(`^\s*//+\s*(.+)$`)
	reBlockComment  = regexp.MustCompile(`/\*\*\s*(.+?)\s*\*/`)
)

var baseModelFields = map[string]string{
	"id":         "主键 ID",
	"createTime": "创建时间",
	"updateTime": "更新时间",
}

var defaultFields = map[string]string{
	"id":         "主键 ID",
	"createTime": "创建时间",
	"updateTime": "更新时间",
	"isDelete":   "软删除标记",
	"version":    "乐观锁版本号",
}

var relationDecorators = []string{
	"@OneToMany", "@ManyToOne", "@OneToOne", "@ManyToMany", "@JoinTable", "@JoinColumn",
}

var columnDecorators = []string{
	"@Column", "@PrimaryGeneratedColumn", "@CreateDateColumn", "@UpdateDateColumn", "@VersionColumn",
}

// Index 表名 -> 字段名 -> 中文说明。
type Index map[string]map[string]string

// Load 扫描 Nest 工程下实体文件并构建注释索引。
func Load(nestRoot string) (Index, error) {
	src := filepath.Join(nestRoot, "src")
	idx := Index{}
	err := filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !isEntityFile(path) {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		table, fields := parseEntity(string(b))
		if table == "" || len(fields) == 0 {
			return nil
		}
		if idx[table] == nil {
			idx[table] = map[string]string{}
		}
		for k, v := range fields {
			if v == "" {
				continue
			}
			idx[table][k] = v
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	mergeBaseModel(idx, nestRoot)
	return idx, nil
}

func isEntityFile(path string) bool {
	base := filepath.Base(path)
	if !strings.HasSuffix(base, ".ts") {
		return false
	}
	return strings.Contains(base, "entity") || strings.HasSuffix(base, ".entity.ts")
}

func mergeBaseModel(idx Index, nestRoot string) {
	path := filepath.Join(nestRoot, "src", "modules", "features", "common", "common.entiry.ts")
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	_, fields := parseEntity(string(b))
	for k, v := range fields {
		if v != "" {
			baseModelFields[k] = v
		}
	}
	for table := range idx {
		for k, v := range baseModelFields {
			if _, ok := idx[table][k]; !ok && v != "" {
				idx[table][k] = v
			}
		}
	}
}

func parseEntity(content string) (string, map[string]string) {
	table := extractTableName(content)
	if table == "" {
		return "", nil
	}
	extendsBase := reExtendsBase.MatchString(content)

	lines := strings.Split(content, "\n")
	fields := map[string]string{}
	var pendingComment string
	var decoratorBuf []string
	inClass := false
	depth := 0

	flushField := func(propLine string) {
		m := reProperty.FindStringSubmatch(strings.TrimSpace(propLine))
		if m == nil {
			return
		}
		name := m[1]
		if isRelationField(decoratorBuf) {
			decoratorBuf = nil
			pendingComment = ""
			return
		}
		if !hasColumnDecorator(decoratorBuf) {
			decoratorBuf = nil
			pendingComment = ""
			return
		}
		comment := pickComment(decoratorBuf, pendingComment, name)
		if comment != "" {
			fields[name] = comment
		}
		decoratorBuf = nil
		pendingComment = ""
	}

	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.Contains(trim, "export class") {
			inClass = true
			depth = strings.Count(trim, "{") - strings.Count(trim, "}")
			if depth <= 0 && strings.HasSuffix(trim, "{") {
				depth = 1
			}
			continue
		}
		if !inClass {
			continue
		}
		depth += strings.Count(line, "{") - strings.Count(line, "}")
		if depth <= 0 {
			break
		}

		if m := reLineComment.FindStringSubmatch(line); m != nil {
			text := strings.TrimSpace(m[1])
			if !strings.HasPrefix(text, "@") && !strings.HasPrefix(text, "deprecated") {
				pendingComment = text
			}
			continue
		}
		if m := reBlockComment.FindStringSubmatch(trim); m != nil {
			pendingComment = strings.TrimSpace(m[1])
			continue
		}
		if strings.HasPrefix(trim, "@") {
			decoratorBuf = append(decoratorBuf, trim)
			continue
		}
		if reProperty.MatchString(trim) {
			flushField(trim)
		}
	}

	if extendsBase {
		for k, v := range baseModelFields {
			if _, ok := fields[k]; !ok && v != "" {
				fields[k] = v
			}
		}
	}
	return table, fields
}

func extractTableName(content string) string {
	if m := reEntityString.FindStringSubmatch(content); m != nil {
		return m[1]
	}
	if m := reEntityObject.FindStringSubmatch(content); m != nil {
		return m[1]
	}
	if m := reClassName.FindStringSubmatch(content); m != nil {
		return pascalToSnake(m[1])
	}
	return ""
}

func pascalToSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func isRelationField(decorators []string) bool {
	for _, d := range decorators {
		for _, prefix := range []string{"@OneToMany", "@ManyToOne", "@OneToOne", "@ManyToMany"} {
			if strings.HasPrefix(d, prefix) {
				return true
			}
		}
	}
	return false
}

func hasColumnDecorator(decorators []string) bool {
	for _, d := range decorators {
		for _, prefix := range columnDecorators {
			if strings.HasPrefix(d, prefix) {
				return true
			}
		}
	}
	return false
}

func pickComment(decorators []string, pending, field string) string {
	block := strings.Join(decorators, " ")
	if m := reColumnComment.FindStringSubmatch(block); m != nil {
		return cleanComment(m[1])
	}
	if m := reDesc.FindStringSubmatch(block); m != nil {
		return cleanComment(m[1])
	}
	if pending != "" {
		return cleanComment(pending)
	}
	if c, ok := defaultFields[field]; ok {
		return c
	}
	return ""
}

func cleanComment(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "@description:")
	s = strings.TrimSpace(s)
	return s
}

// Lookup 按逻辑表名与列名取注释（含默认 fallback）。
func (idx Index) Lookup(table, column string) string {
	if cols, ok := idx[table]; ok {
		if c, ok := cols[column]; ok && c != "" {
			return c
		}
	}
	if c, ok := defaultFields[column]; ok {
		return c
	}
	return ""
}
