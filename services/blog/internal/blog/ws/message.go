package ws

import "encoding/json"

// Message 业务消息格式（替代 Socket.IO event 名）。
type Message struct {
	Type string          `json:"type"`
	Seq  uint64          `json:"seq,omitempty"`
	Data json.RawMessage `json:"data,omitempty"`
}

// SubscribeData topic 订阅/取消订阅 payload。
type SubscribeData struct {
	Topic string `json:"topic"`
}

// RedisPushMsg 跨模块经 Redis pub/sub 推送的结构。
type RedisPushMsg struct {
	UserID uint64          `json:"userId"`
	Topic  string          `json:"topic,omitempty"`
	Body   json.RawMessage `json:"body"`
}

// MsgSiteNotification 站内通知 WS 事件名，与 Nest RealtimeWsEvents.SITE_NOTIFICATION 对齐。
const MsgSiteNotification = "siteNotification"
