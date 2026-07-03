package rag

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// ChatTurn 对话历史轮次。
type ChatTurn struct {
	Role    string
	Content string
}

// UiMessageStreamIDs SSE 流 ID。
type UiMessageStreamIDs struct {
	MessageID string
	TextID    string
}

// CreateUiMessageStreamIDs 生成 messageId 与 textId。
func CreateUiMessageStreamIDs() UiMessageStreamIDs {
	return UiMessageStreamIDs{
		MessageID: uuid.NewString(),
		TextID:    "text-" + uuid.NewString(),
	}
}

// FormatUiMessageSSE 格式化为 AI SDK UI Message SSE data 行。
func FormatUiMessageSSE(payload interface{}) string {
	b, _ := json.Marshal(payload)
	return "data: " + string(b) + "\n\n"
}

// FormatUiMessageSSEDone 返回 [DONE] 帧。
func FormatUiMessageSSEDone() string {
	return "data: [DONE]\n\n"
}

type ragQueryBody struct {
	Question string        `json:"question"`
	Messages []uiMessage   `json:"messages"`
}

type uiMessage struct {
	Role    string    `json:"role"`
	Content string    `json:"content"`
	Parts   []uiPart  `json:"parts"`
}

type uiPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func extractUiMessageText(msg uiMessage) string {
	if strings.TrimSpace(msg.Content) != "" {
		return strings.TrimSpace(msg.Content)
	}
	var parts []string
	for _, p := range msg.Parts {
		if p.Type == "text" && strings.TrimSpace(p.Text) != "" {
			parts = append(parts, p.Text)
		}
	}
	return strings.TrimSpace(strings.Join(parts, ""))
}

// ResolveQuestion 从 body 解析用户问题。
func ResolveQuestion(body ragQueryBody) (string, error) {
	if q := strings.TrimSpace(body.Question); q != "" {
		if len([]rune(q)) > MaxQuestionChars {
			q = string([]rune(q)[:MaxQuestionChars])
		}
		return q, nil
	}
	if len(body.Messages) == 0 {
		return "", fmt.Errorf("请输入问题")
	}
	for i := len(body.Messages) - 1; i >= 0; i-- {
		msg := body.Messages[i]
		if msg.Role != "user" {
			continue
		}
		text := extractUiMessageText(msg)
		if text != "" {
			if len([]rune(text)) > MaxQuestionChars {
				text = string([]rune(text)[:MaxQuestionChars])
			}
			return text, nil
		}
	}
	return "", fmt.Errorf("请输入问题")
}

// ExtractChatHistory 提取多轮历史（不含当前最后一条 user）。
func ExtractChatHistory(body ragQueryBody) []ChatTurn {
	if strings.TrimSpace(body.Question) != "" {
		return nil
	}
	if len(body.Messages) == 0 {
		return nil
	}
	var turns []ChatTurn
	for _, raw := range body.Messages {
		if raw.Role != "user" && raw.Role != "assistant" {
			continue
		}
		text := extractUiMessageText(raw)
		if text == "" {
			continue
		}
		if len([]rune(text)) > MaxQuestionChars {
			text = string([]rune(text)[:MaxQuestionChars])
		}
		turns = append(turns, ChatTurn{Role: raw.Role, Content: text})
	}
	end := len(turns)
	for i := len(turns) - 1; i >= 0; i-- {
		if turns[i].Role == "user" {
			end = i
			break
		}
	}
	prior := turns[:end]
	maxMessages := MaxHistoryTurns * 2
	if len(prior) > maxMessages {
		prior = prior[len(prior)-maxMessages:]
	}
	total := 0
	for _, t := range prior {
		total += len([]rune(t.Content))
	}
	for len(prior) > 0 && total > MaxHistoryChars {
		total -= len([]rune(prior[0].Content))
		prior = prior[1:]
	}
	return prior
}

// ParseQueryBody 解析 query-stream 请求体。
func ParseQueryBody(raw []byte) (ragQueryBody, error) {
	var body ragQueryBody
	if err := json.Unmarshal(raw, &body); err != nil {
		return body, err
	}
	return body, nil
}
