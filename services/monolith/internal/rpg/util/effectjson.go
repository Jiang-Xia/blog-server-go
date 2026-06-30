package util

import "encoding/json"

// ParseEffectJSON 解析 ent 中 *string 类型的 effectJson 为 map。
func ParseEffectJSON(raw *string) map[string]interface{} {
	if raw == nil || *raw == "" {
		return map[string]interface{}{}
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(*raw), &m); err != nil {
		return map[string]interface{}{}
	}
	return m
}

// MustEffectJSON 序列化 effectJson，失败返回 nil。
func MustEffectJSON(v interface{}) *string {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	s := string(b)
	return &s
}
