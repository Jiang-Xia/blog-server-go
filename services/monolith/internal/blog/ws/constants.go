// Package ws 博客实时 WebSocket Hub：连接管理、topic 路由、Redis pub/sub 跨模块推送。
package ws

import "time"

const (
	// ChannelRealtimePush Redis pub/sub 频道，跨模块推送入口。
	ChannelRealtimePush = "realtime:push"

	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
	sendBufSize    = 256
)
