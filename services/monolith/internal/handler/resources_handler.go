// resources_handler 用户资源/文件 HTTP 端点。
package handler

import (
	"context"
	"strconv"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	blogsvc "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/service"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/auth"
	"github.com/cloudwego/hertz/pkg/app"
)

// FileHandler 大文件分片上传 HTTP 端点。
type FileHandler struct {
	svc *blogsvc.ResourcesService
	jwt *auth.JWTService
}

// NewFileHandler 构造 FileHandler。
func NewFileHandler(svc *blogsvc.ResourcesService, jwt *auth.JWTService) *FileHandler {
	return &FileHandler{svc: svc, jwt: jwt}
}

func (h *FileHandler) UploadBigFile(ctx context.Context, c *app.RequestContext) {
	hash := string(c.Query("hash"))
	index, _ := strconv.Atoi(string(c.Query("index")))
	file, err := c.FormFile("fileContents")
	if err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	f, err := file.Open()
	if err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	defer f.Close()
	data := make([]byte, file.Size)
	_, _ = f.Read(data)
	handleAdminResult(ctx, c, nil, h.svc.SaveBigFileChunk(hash, index, data))
}

func (h *FileHandler) MergeBigFile(ctx context.Context, c *app.RequestContext) {
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	chunks := intFieldDefault(body, "chunks", 0)
	data, err := h.svc.MergeBigFile(ctx, strField(body, "hash"), strField(body, "fileName"), chunks)
	handleAdminResult(ctx, c, data, err)
}

func (h *FileHandler) CheckBigFile(ctx context.Context, c *app.RequestContext) {
	hash := string(c.Query("hash"))
	data, err := h.svc.CheckBigFile(ctx, hash)
	handleAdminResult(ctx, c, data, err)
}

// ResourcesHandler 资源库与代理 HTTP 端点。
type ResourcesHandler struct {
	svc *blogsvc.ResourcesService
	jwt *auth.JWTService
}

// NewResourcesHandler 构造 ResourcesHandler。
func NewResourcesHandler(svc *blogsvc.ResourcesService, jwt *auth.JWTService) *ResourcesHandler {
	return &ResourcesHandler{svc: svc, jwt: jwt}
}

func (h *ResourcesHandler) DailyImg(ctx context.Context, c *app.RequestContext) {
	data, err := h.svc.DailyImg(ctx)
	handleAdminResult(ctx, c, data, err)
}

func (h *ResourcesHandler) Weather(ctx context.Context, c *app.RequestContext) {
	ip := string(c.Query("ip"))
	if ip == "" {
		ip = string(c.ClientIP())
	}
	data, err := h.svc.Weather(ctx, ip)
	handleAdminResult(ctx, c, data, err)
}

func (h *ResourcesHandler) UploadFile(ctx context.Context, c *app.RequestContext) {
	pid := string(c.Query("pid"))
	category := string(c.Query("category"))
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	f, err := file.Open()
	if err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	defer f.Close()
	data := make([]byte, file.Size)
	_, _ = f.Read(data)
	row, err := h.svc.UploadFile(ctx, pid, category, file.Filename, data, file.Header.Get("Content-Type"))
	handleAdminResult(ctx, c, row, err)
}

func (h *ResourcesHandler) UploadMedia(ctx context.Context, c *app.RequestContext) {
	category := string(c.Query("category"))
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	f, err := file.Open()
	if err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	defer f.Close()
	data := make([]byte, file.Size)
	_, _ = f.Read(data)
	contentHash := string(c.Query("contentHash"))
	row, err := h.svc.UploadMedia(ctx, category, file.Filename, data, file.Header.Get("Content-Type"), contentHash)
	handleAdminResult(ctx, c, row, err)
}

func (h *ResourcesHandler) RegisterAvatar(ctx context.Context, c *app.RequestContext) {
	h.UploadMedia(ctx, c)
}

func (h *ResourcesHandler) Files(ctx context.Context, c *app.RequestContext) {
	pid := string(c.Query("pid"))
	page, _ := strconv.Atoi(string(c.Query("page")))
	pageSize, _ := strconv.Atoi(string(c.Query("pageSize")))
	data, err := h.svc.ListFiles(ctx, pid, page, pageSize)
	handleAdminResult(ctx, c, data, err)
}

func (h *ResourcesHandler) RegisterAvatars(ctx context.Context, c *app.RequestContext) {
	data, err := h.svc.RegisterAvatars(ctx)
	handleAdminResult(ctx, c, data, err)
}

func (h *ResourcesHandler) GetFile(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	data, err := h.svc.GetFile(ctx, id)
	handleAdminResult(ctx, c, data, err)
}

func (h *ResourcesHandler) DeleteFile(ctx context.Context, c *app.RequestContext) {
	id := string(c.Query("id"))
	handleAdminResult(ctx, c, nil, h.svc.DeleteFile(ctx, id))
}

func (h *ResourcesHandler) CreateFolder(ctx context.Context, c *app.RequestContext) {
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.CreateFolder(ctx, strField(body, "pid"), strField(body, "name"))
	handleAdminResult(ctx, c, data, err)
}

func (h *ResourcesHandler) UpdateFile(ctx context.Context, c *app.RequestContext) {
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	id := strField(body, "id")
	delete(body, "id")
	data, err := h.svc.UpdateFile(ctx, id, body)
	handleAdminResult(ctx, c, data, err)
}
