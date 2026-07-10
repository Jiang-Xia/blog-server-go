package tools

// CallRecord 单次 Tool 调用记录，注入 LLM prompt。
type CallRecord struct {
	Name   string
	Args   map[string]interface{}
	Result interface{}
}
