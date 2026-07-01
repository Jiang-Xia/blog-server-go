// Package handler ???????
package handler

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/Jiang-Xia/blog-server-go/services/user/ent"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/redis/rueidis"
)

// HealthHandler ???????????
type HealthHandler struct {
	ent   *ent.Client
	redis rueidis.Client
}

// NewHealthHandler ?? HealthHandler?
func NewHealthHandler(entClient *ent.Client, redisClient rueidis.Client) *HealthHandler {
	return &HealthHandler{ent: entClient, redis: redisClient}
}

// OK ?????????
func (h *HealthHandler) OK(ctx context.Context, c *app.RequestContext) {
	response.Success(ctx, c, "ok")
}
