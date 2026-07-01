// Package handler 博客域路由注册依赖。
package handler

import (
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/auth"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/middleware"
	"github.com/Jiang-Xia/blog-server-go/pkg/usersvc"
)

// RegisterDeps 路由注册依赖，由 wire 装配后传入。
type RegisterDeps struct {
	Health       *HealthHandler
	Article      *ArticleHandler
	Category     *CategoryHandler
	Tag          *TagHandler
	Comment      *CommentHandler
	Reply        *ReplyHandler
	Like         *LikeHandler
	Collect      *CollectHandler
	Msgboard     *MsgboardHandler
	Link         *LinkHandler
	File         *FileHandler
	Resources    *ResourcesHandler
	Notification *NotificationHandler
	WS           *WSHandler
	DevPush      *DevPushHandler
	JWT          *auth.JWTService
	Users        usersvc.UserService
	Permission   middleware.PermissionDeps
}
