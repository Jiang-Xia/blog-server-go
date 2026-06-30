package handler

import (
	"net/http"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/ws"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/auth"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/adaptor"
	"github.com/gorilla/websocket"
)

// WSHandler WebSocket 升级入口 GET /realtime。
type WSHandler struct {
	hub *ws.Hub
	jwt *auth.JWTService
}

// NewWSHandler 构造 WSHandler。
func NewWSHandler(hub *ws.Hub, jwt *auth.JWTService) *WSHandler {
	return &WSHandler{hub: hub, jwt: jwt}
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin:     func(_ *http.Request) bool { return true },
}

// Register 注册 /realtime WebSocket 路由。
func (h *WSHandler) Register(r *server.Hertz) {
	r.GET("/realtime", adaptor.HertzHandler(http.HandlerFunc(h.serveWS)))
}

func (h *WSHandler) serveWS(w http.ResponseWriter, r *http.Request) {
	if h.jwt == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		authz := strings.TrimSpace(r.Header.Get("Authorization"))
		token = strings.TrimPrefix(authz, "Bearer ")
		token = strings.TrimSpace(token)
	}
	if token == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	claims, err := h.jwt.Verify(token)
	if err != nil || claims == nil || claims.ID <= 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := ws.NewClient(uint64(claims.ID), conn, h.hub)
	h.hub.Register(client)
	go client.WritePump()
	go client.ReadPump()
}
