// Package handler RPG 后台管理 HTTP 端点，路径对齐 Nest RpgAdminController。
package handler

import (
	"context"
	"strconv"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/auth"
	"github.com/cloudwego/hertz/pkg/app"
)

// RPGAdminService 后台 RPG 管理（由 rpg/admin 实现）。
type RPGAdminService interface {
	ListAchievements(ctx context.Context, query map[string]string) (interface{}, error)
	CreateAchievement(ctx context.Context, body map[string]interface{}) (interface{}, error)
	UpdateAchievement(ctx context.Context, id string, body map[string]interface{}) (interface{}, error)
	DeleteAchievement(ctx context.Context, id string) (interface{}, error)
	ListQuests(ctx context.Context, query map[string]string) (interface{}, error)
	CreateQuest(ctx context.Context, body map[string]interface{}) (interface{}, error)
	UpdateQuest(ctx context.Context, id string, body map[string]interface{}) (interface{}, error)
	DeleteQuest(ctx context.Context, id string) (interface{}, error)
	ListLotteryPool(ctx context.Context, query map[string]string) (interface{}, error)
	CreateLotteryPool(ctx context.Context, body map[string]interface{}) (interface{}, error)
	UpdateLotteryPool(ctx context.Context, id string, body map[string]interface{}) (interface{}, error)
	DeleteLotteryPool(ctx context.Context, id string) (interface{}, error)
	ListLotteryRecords(ctx context.Context, query map[string]string) (interface{}, error)
	ListUserRpgData(ctx context.Context, query map[string]string) (interface{}, error)
	RechargeCurrency(ctx context.Context, uid string, body map[string]interface{}, operatorUID int) (interface{}, error)
	DeductCurrency(ctx context.Context, uid string, body map[string]interface{}, operatorUID int) (interface{}, error)
	UnbanUser(ctx context.Context, uid string, operatorUID int) (interface{}, error)
	GetUserRpgDetail(ctx context.Context, uid string) (interface{}, error)
	GetStats(ctx context.Context) (interface{}, error)
	ListItems(ctx context.Context, query map[string]string) (interface{}, error)
	CreateItem(ctx context.Context, body map[string]interface{}) (interface{}, error)
	UpdateItem(ctx context.Context, id string, body map[string]interface{}) (interface{}, error)
	DeleteItem(ctx context.Context, id string) (interface{}, error)
	UploadItemAsset(ctx context.Context, icon, assetType string, file []byte, filename string) (interface{}, error)
	DeleteItemAsset(ctx context.Context, icon, assetType string) (interface{}, error)
	ListActivities(ctx context.Context, query map[string]string) (interface{}, error)
	CreateActivity(ctx context.Context, body map[string]interface{}) (interface{}, error)
	UpdateActivity(ctx context.Context, id string, body map[string]interface{}) (interface{}, error)
	DeleteActivity(ctx context.Context, id string) (interface{}, error)
	ListGuilds(ctx context.Context, query map[string]string) (interface{}, error)
	DeleteGuild(ctx context.Context, id string) (interface{}, error)
	ListGuildMembers(ctx context.Context, id string) (interface{}, error)
	RemoveGuildMember(ctx context.Context, guildID, uid string) (interface{}, error)
	ListTips(ctx context.Context, query map[string]string) (interface{}, error)
	ListSocialLogs(ctx context.Context, query map[string]string) (interface{}, error)
}

// RPGAdminHandler /admin/rpg/* 路由（需 JWT）。
type RPGAdminHandler struct {
	svc RPGAdminService
	jwt *auth.JWTService
}

// NewRPGAdminHandler 构造 RPGAdminHandler。
func NewRPGAdminHandler(svc RPGAdminService, jwt *auth.JWTService) *RPGAdminHandler {
	return &RPGAdminHandler{svc: svc, jwt: jwt}
}

func (h *RPGAdminHandler) uid(ctx context.Context, c *app.RequestContext) int {
	return articleUID(ctx, c, h.jwt)
}

func (h *RPGAdminHandler) requireSvc(ctx context.Context, c *app.RequestContext) bool {
	if h.svc == nil {
		response.Error(ctx, c, errcode.WithMessage(errcode.InternalError, "RPG 管理模块加载中"))
		return false
	}
	return true
}

func queryMap(c *app.RequestContext) map[string]string {
	out := map[string]string{}
	c.QueryArgs().VisitAll(func(k, v []byte) {
		out[string(k)] = string(v)
	})
	return out
}

func (h *RPGAdminHandler) ListAchievements(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	data, err := h.svc.ListAchievements(ctx, queryMap(c))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) CreateAchievement(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.CreateAchievement(ctx, body)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) UpdateAchievement(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.UpdateAchievement(ctx, c.Param("id"), body)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) DeleteAchievement(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	data, err := h.svc.DeleteAchievement(ctx, c.Param("id"))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) ListQuests(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	data, err := h.svc.ListQuests(ctx, queryMap(c))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) CreateQuest(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.CreateQuest(ctx, body)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) UpdateQuest(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.UpdateQuest(ctx, c.Param("id"), body)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) DeleteQuest(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	data, err := h.svc.DeleteQuest(ctx, c.Param("id"))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) ListLotteryPool(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	data, err := h.svc.ListLotteryPool(ctx, queryMap(c))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) CreateLotteryPool(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.CreateLotteryPool(ctx, body)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) UpdateLotteryPool(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.UpdateLotteryPool(ctx, c.Param("id"), body)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) DeleteLotteryPool(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	data, err := h.svc.DeleteLotteryPool(ctx, c.Param("id"))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) ListLotteryRecords(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	data, err := h.svc.ListLotteryRecords(ctx, queryMap(c))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) ListUsers(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	data, err := h.svc.ListUserRpgData(ctx, queryMap(c))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) RechargeCurrency(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.RechargeCurrency(ctx, c.Param("uid"), body, h.uid(ctx, c))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) DeductCurrency(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.DeductCurrency(ctx, c.Param("uid"), body, h.uid(ctx, c))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) UnbanUser(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	data, err := h.svc.UnbanUser(ctx, c.Param("uid"), h.uid(ctx, c))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) GetUserDetail(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	data, err := h.svc.GetUserRpgDetail(ctx, c.Param("uid"))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) Stats(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	data, err := h.svc.GetStats(ctx)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) ListItems(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	data, err := h.svc.ListItems(ctx, queryMap(c))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) CreateItem(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.CreateItem(ctx, body)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) UploadItemAsset(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	fileHeader, err := c.FormFile("file")
	if err != nil || fileHeader == nil {
		response.Error(ctx, c, errcode.WithMessage(errcode.InvalidParam, "未收到上传文件"))
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	defer f.Close()
	buf := make([]byte, fileHeader.Size)
	_, _ = f.Read(buf)
	icon := string(c.FormValue("icon"))
	assetType := string(c.FormValue("assetType"))
	data, err := h.svc.UploadItemAsset(ctx, icon, assetType, buf, fileHeader.Filename)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) DeleteItemAsset(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	data, err := h.svc.DeleteItemAsset(ctx, string(c.Query("icon")), string(c.Query("assetType")))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) UpdateItem(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.UpdateItem(ctx, c.Param("id"), body)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) DeleteItem(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	data, err := h.svc.DeleteItem(ctx, c.Param("id"))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) ListActivities(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	data, err := h.svc.ListActivities(ctx, queryMap(c))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) CreateActivity(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.CreateActivity(ctx, body)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) UpdateActivity(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.UpdateActivity(ctx, c.Param("id"), body)
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) DeleteActivity(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	data, err := h.svc.DeleteActivity(ctx, c.Param("id"))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) ListGuilds(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	data, err := h.svc.ListGuilds(ctx, queryMap(c))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) DeleteGuild(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	data, err := h.svc.DeleteGuild(ctx, c.Param("id"))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) ListGuildMembers(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	data, err := h.svc.ListGuildMembers(ctx, c.Param("id"))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) RemoveGuildMember(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	data, err := h.svc.RemoveGuildMember(ctx, c.Param("id"), c.Param("uid"))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) ListTips(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	data, err := h.svc.ListTips(ctx, queryMap(c))
	handleAdminResult(ctx, c, data, err)
}

func (h *RPGAdminHandler) ListSocialLogs(ctx context.Context, c *app.RequestContext) {
	if !h.requireSvc(ctx, c) {
		return
	}
	data, err := h.svc.ListSocialLogs(ctx, queryMap(c))
	handleAdminResult(ctx, c, data, err)
}

// parseUIDParam 解析路径 uid 参数。
func parseUIDParam(c *app.RequestContext) int {
	uid, _ := strconv.Atoi(c.Param("uid"))
	return uid
}
