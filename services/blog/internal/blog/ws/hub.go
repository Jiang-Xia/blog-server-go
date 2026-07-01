package ws

import (
	"context"
	"encoding/json"
	"sync"
)

// Hub 管理 WebSocket 连接、topic 路由与消息分发。
type Hub struct {
	mu sync.RWMutex

	clients map[uint64]map[*Client]struct{}
	topics  map[string]map[uint64]struct{}

	register   chan *Client
	unregister chan *Client
	broadcast  chan broadcastMsg
}

type broadcastMsg struct {
	userIDFilter uint64
	topicFilter  string
	payload      []byte
}

// NewHub 构造 Hub。
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uint64]map[*Client]struct{}),
		topics:     make(map[string]map[uint64]struct{}),
		register:   make(chan *Client, 64),
		unregister: make(chan *Client, 64),
		broadcast:  make(chan broadcastMsg, 256),
	}
}

// Run 处理注册/注销/广播；ctx 取消时优雅关闭全部连接。
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case c := <-h.register:
			h.addClient(c)
		case c := <-h.unregister:
			h.removeClient(c)
		case msg := <-h.broadcast:
			h.dispatch(msg)
		case <-ctx.Done():
			h.closeAll()
			return
		}
	}
}

// Register 注册新连接（由 handler 调用）。
func (h *Hub) Register(c *Client) { h.register <- c }

// Unregister 注销连接。
func (h *Hub) Unregister(c *Client) { h.unregister <- c }

// PublishToUser 向指定用户推送。
func (h *Hub) PublishToUser(userID uint64, payload []byte) {
	h.broadcast <- broadcastMsg{userIDFilter: userID, payload: payload}
}

// OnlineUIDs 返回当前在线用户 uid 列表（快照）。
func (h *Hub) OnlineUIDs() []uint64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]uint64, 0, len(h.clients))
	for uid := range h.clients {
		out = append(out, uid)
	}
	return out
}

// PublishToTopic 向 topic 订阅者推送。
func (h *Hub) PublishToTopic(topic string, payload []byte) {
	h.broadcast <- broadcastMsg{topicFilter: topic, payload: payload}
}

// DispatchRaw 解析 Redis pub/sub 消息后分发。
func (h *Hub) DispatchRaw(raw []byte) {
	var msg RedisPushMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		// 兼容直接推送 WS 消息 JSON（无 envelope）
		h.broadcast <- broadcastMsg{payload: raw}
		return
	}
	h.broadcast <- broadcastMsg{
		userIDFilter: msg.UserID,
		topicFilter:  msg.Topic,
		payload:      msg.Body,
	}
}

// SubscribeTopic 用户订阅 topic。
func (h *Hub) SubscribeTopic(userID uint64, topic string) {
	if topic == "" {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.topics[topic] == nil {
		h.topics[topic] = make(map[uint64]struct{})
	}
	h.topics[topic][userID] = struct{}{}
}

// UnsubscribeTopic 用户取消订阅 topic。
func (h *Hub) UnsubscribeTopic(userID uint64, topic string) {
	if topic == "" {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if users, ok := h.topics[topic]; ok {
		delete(users, userID)
		if len(users) == 0 {
			delete(h.topics, topic)
		}
	}
}

func (h *Hub) addClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[c.userID] == nil {
		h.clients[c.userID] = make(map[*Client]struct{})
	}
	h.clients[c.userID][c] = struct{}{}
}

func (h *Hub) removeClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.clients[c.userID]; ok {
		delete(conns, c)
		if len(conns) == 0 {
			delete(h.clients, c.userID)
			for topic, users := range h.topics {
				delete(users, c.userID)
				if len(users) == 0 {
					delete(h.topics, topic)
				}
			}
		}
	}
}

func (h *Hub) dispatch(msg broadcastMsg) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var targets map[uint64]struct{}
	if msg.topicFilter != "" {
		targets = h.topics[msg.topicFilter]
	}

	for uid, conns := range h.clients {
		if msg.userIDFilter != 0 && uid != msg.userIDFilter {
			continue
		}
		if msg.topicFilter != "" {
			if _, ok := targets[uid]; !ok {
				continue
			}
		}
		for c := range conns {
			select {
			case c.send <- msg.payload:
			default:
				go func(cl *Client) { h.unregister <- cl }(c)
			}
		}
	}
}

func (h *Hub) closeAll() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, conns := range h.clients {
		for c := range conns {
			close(c.send)
		}
	}
}
