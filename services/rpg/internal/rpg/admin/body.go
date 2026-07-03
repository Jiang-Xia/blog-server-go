// Package admin 请求体解析辅助。
package admin

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func strField(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(fmt.Sprint(v))
}

func intField(m map[string]interface{}, key string, def int) int {
	if m == nil {
		return def
	}
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(n))
		if err != nil {
			return def
		}
		return i
	default:
		return def
	}
}

func floatField(m map[string]interface{}, key string, def float64) float64 {
	if m == nil {
		return def
	}
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(n), 64)
		if err != nil {
			return def
		}
		return f
	default:
		return def
	}
}

func boolToIntField(m map[string]interface{}, key string, def int) int {
	if m == nil {
		return def
	}
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	switch b := v.(type) {
	case bool:
		if b {
			return 1
		}
		return 0
	case float64:
		return int(b)
	case int:
		return b
	default:
		return def
	}
}

func effectJSONString(m map[string]interface{}) (*string, error) {
	if m == nil {
		return nil, nil
	}
	raw, ok := m["effectJson"]
	if !ok || raw == nil {
		return nil, nil
	}
	switch v := raw.(type) {
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return nil, nil
		}
		return &s, nil
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		s := string(b)
		return &s, nil
	}
}

func mergeEffectJSON(existing *string, patch map[string]interface{}) (map[string]interface{}, error) {
	base := map[string]interface{}{}
	if existing != nil && strings.TrimSpace(*existing) != "" {
		if err := json.Unmarshal([]byte(*existing), &base); err != nil {
			return nil, err
		}
	}
	if patch != nil {
		if raw, ok := patch["effectJson"]; ok && raw != nil {
			switch v := raw.(type) {
			case map[string]interface{}:
				for k, val := range v {
					base[k] = val
				}
			case string:
				if strings.TrimSpace(v) != "" {
					extra := map[string]interface{}{}
					if err := json.Unmarshal([]byte(v), &extra); err != nil {
						return nil, err
					}
					for k, val := range extra {
						base[k] = val
					}
				}
			}
		}
	}
	return base, nil
}

func effectToRepoString(effect map[string]interface{}) (*string, error) {
	if len(effect) == 0 {
		return nil, nil
	}
	b, err := json.Marshal(effect)
	if err != nil {
		return nil, err
	}
	s := string(b)
	return &s, nil
}

func parseTimeField(m map[string]interface{}, key string) (time.Time, error) {
	s := strField(m, key)
	if s == "" {
		return time.Time{}, fmt.Errorf("%s required", key)
	}
	layouts := []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05", "2006-01-02"}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time: %s", s)
}
