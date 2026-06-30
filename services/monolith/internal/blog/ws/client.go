package ws

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

// Client 单 WebSocket 连接读写循环。
type Client struct {
	userID uint64
	conn   *websocket.Conn
	send   chan []byte
	hub    *Hub
}

// NewClient 构造 Client。
func NewClient(userID uint64, conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		userID: userID,
		conn:   conn,
		send:   make(chan []byte, sendBufSize),
		hub:    hub,
	}
}

// ReadPump 读循环：应用层 ping/subscribe + 控制帧 pong 续期。
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		var msg Message
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}
		switch msg.Type {
		case "ping":
			c.send <- []byte(`{"type":"pong"}`)
		case "subscribe":
			var d SubscribeData
			if json.Unmarshal(msg.Data, &d) == nil {
				c.hub.SubscribeTopic(c.userID, d.Topic)
			}
		case "unsubscribe":
			var d SubscribeData
			if json.Unmarshal(msg.Data, &d) == nil {
				c.hub.UnsubscribeTopic(c.userID, d.Topic)
			}
		}
	}
}

// WritePump 写循环：单一 writeLoop + 控制帧 ping。
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
