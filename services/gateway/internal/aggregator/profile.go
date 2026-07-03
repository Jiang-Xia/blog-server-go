// Package aggregator gateway BFF 聚合接口。
package aggregator

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	rpgv1 "github.com/Jiang-Xia/blog-server-go/proto/gen/go/rpg/v1"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/Jiang-Xia/blog-server-go/services/gateway/internal/grpcclient"
	"github.com/cloudwego/hertz/pkg/app"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProfileHandler GET /user/public/:uid BFF：经 rpg gRPC 取公开主页。
type ProfileHandler struct {
	clients *grpcclient.Clients
}

// NewProfileHandler 构造 user/public BFF handler。
func NewProfileHandler(clients *grpcclient.Clients) *ProfileHandler {
	return &ProfileHandler{clients: clients}
}

// PublicProfile 返回公开用户主页（user + RPG 聚合）。
func (h *ProfileHandler) PublicProfile(ctx context.Context, c *app.RequestContext) {
	uid, ok := publicProfileUID(string(c.Path()))
	if !ok {
		response.Error(ctx, c, errcode.WithMessage(errcode.InvalidParam, "无效的用户 ID"))
		return
	}
	if h.clients == nil || h.clients.RPG == nil {
		response.Error(ctx, c, errcode.WithMessage(errcode.InternalError, "rpg gRPC 未配置"))
		return
	}
	resp, err := h.clients.RPG.GetPublicProfile(ctx, &rpgv1.GetPublicProfileRequest{UserId: uint64(uid)})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			response.Error(ctx, c, errcode.WithMessage(errcode.NotFound, "用户不存在"))
			return
		}
		response.Error(ctx, c, errcode.WithMessage(errcode.InternalError, err.Error()))
		return
	}
	var data interface{}
	if err := json.Unmarshal(resp.GetProfileJson(), &data); err != nil {
		response.Error(ctx, c, errcode.WithMessage(errcode.InternalError, "解析主页数据失败"))
		return
	}
	response.Success(ctx, c, data)
}

func publicProfileUID(path string) (int, bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 {
		return 0, false
	}
	n := len(parts)
	if parts[n-3] != "user" || parts[n-2] != "public" {
		return 0, false
	}
	uid, err := strconv.Atoi(parts[n-1])
	return uid, err == nil && uid > 0
}
