// resources_service 用户资源/文件上传与列举。
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	blogrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/repo"
	userrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/repo"
	"github.com/google/uuid"
)

const (
	registerAvatarFolderID = "f25ca7bc-bd12-4c42-95ef-6c1b70f05012"
	articleImageFolderID   = "d5561c87-f189-4dc1-a28d-ba862a50f01f"
	articleCoverFolderID   = "c8e4f2a1-9b3d-4f6e-a5c7-2d8e9f0a1b2c"
	bigFileMergeFolderID   = "19f66b84-8841-4cf5-8932-d11b95947d2d"
)

// ResourcesService 文件资源与代理接口。
type ResourcesService struct {
	files          *blogrepo.FileRepo
	cfg            *config.Config
	client         *http.Client
	redis          *redisutil.Store
	tongjiMu       sync.RWMutex
	tongjiMemCache string
}

// NewResourcesService 构造 ResourcesService。
func NewResourcesService(files *blogrepo.FileRepo, cfg *config.Config, redis *redisutil.Store) *ResourcesService {
	return &ResourcesService{
		files:  files,
		cfg:    cfg,
		redis:  redis,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// DailyImg Bing 每日壁纸代理。
func (s *ResourcesService) DailyImg(ctx context.Context) (map[string]interface{}, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.vvhan.com/api/bing?type=json", nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, errcode.WithMessage(errcode.InternalError, "获取每日壁纸失败")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"raw": string(body)}, nil
}

// Weather IP 归属卡片代理（简化返回 base64 占位）。
func (s *ResourcesService) Weather(ctx context.Context, ip string) (map[string]interface{}, error) {
	if ip == "" {
		ip = "127.0.0.1"
	}
	url := fmt.Sprintf("https://api.vvhan.com/api/ipCard?ip=%s&type=1", ip)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, errcode.WithMessage(errcode.InternalError, "获取天气信息失败")
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return map[string]interface{}{"data": string(data)}, nil
}

// UploadFile 管理端文件上传。
func (s *ResourcesService) UploadFile(ctx context.Context, pid, category, originalName string, data []byte, contentType string) (*ent.File, error) {
	if pid == "" {
		pid = resolveFolderID(category)
	}
	dir := s.diskDir(category)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	ext := filepath.Ext(originalName)
	filename := uuid.NewString() + ext
	diskPath := filepath.Join(dir, filename)
	if err := os.WriteFile(diskPath, data, 0o644); err != nil {
		return nil, err
	}
	url := blogrepo.NormalizePublicURL(diskPath, s.uploadRoot(), s.publicPrefix())
	return s.files.Create(ctx, &ent.File{
		Pid:          pid,
		IsFolder:     0,
		Originalname: originalName,
		Filename:     filename,
		Type:         contentType,
		Size:         len(data),
		URL:          url,
		CreateAt:     time.Now(),
	})
}

// UploadMedia 头像/封面/正文图上传（contentHash 去重）。
func (s *ResourcesService) UploadMedia(ctx context.Context, category, originalName string, data []byte, contentType, contentHash string) (*ent.File, error) {
	hash := strings.ToLower(strings.TrimSpace(contentHash))
	if len(hash) != 64 {
		sum := sha256.Sum256(data)
		hash = hex.EncodeToString(sum[:])
	}
	ext := filepath.Ext(originalName)
	if ext == "" {
		ext = ".jpg"
	}
	filename := hash + ext
	dir := s.diskDir(category)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	diskPath := filepath.Join(dir, filename)
	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		if err := os.WriteFile(diskPath, data, 0o644); err != nil {
			return nil, err
		}
	}
	if existing, err := s.files.FindByFilenameLike(ctx, hash); err == nil {
		return existing, nil
	}
	url := blogrepo.NormalizePublicURL(diskPath, s.uploadRoot(), s.publicPrefix())
	pid := resolveFolderID(category)
	return s.files.Create(ctx, &ent.File{
		Pid:          pid,
		IsFolder:     0,
		Originalname: originalName,
		Filename:     filename,
		Type:         contentType,
		Size:         len(data),
		URL:          url,
		CreateAt:     time.Now(),
	})
}

// ListFiles 文件分页列表。
func (s *ResourcesService) ListFiles(ctx context.Context, pid string, page, pageSize int) (map[string]interface{}, error) {
	rows, total, err := s.files.List(ctx, blogrepo.FileFilter{PID: pid, Page: page, PageSize: pageSize})
	if err != nil {
		return nil, err
	}
	list := make([]map[string]interface{}, 0, len(rows))
	for _, f := range rows {
		list = append(list, fileToMap(f))
	}
	return map[string]interface{}{
		"list":       list,
		"pagination": userrepo.CalcNestPagination(total, pageSize, page),
	}, nil
}

// GetFile 文件详情。
func (s *ResourcesService) GetFile(ctx context.Context, id string) (map[string]interface{}, error) {
	row, err := s.files.GetByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.NotFound, "文件不存在")
		}
		return nil, err
	}
	return fileToMap(row), nil
}

// DeleteFile 删除文件记录及磁盘文件。
func (s *ResourcesService) DeleteFile(ctx context.Context, id string) error {
	row, err := s.files.GetByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return errcode.WithMessage(errcode.NotFound, "文件不存在")
		}
		return err
	}
	if row.IsFolder == 0 && row.URL != "" {
		disk := strings.Replace(row.URL, s.publicPrefix(), s.uploadRoot(), 1)
		disk = strings.ReplaceAll(disk, "/", string(os.PathSeparator))
		_ = os.Remove(disk)
	}
	return s.files.DeleteByID(ctx, id)
}

// CreateFolder 创建虚拟文件夹。
func (s *ResourcesService) CreateFolder(ctx context.Context, pid, name string) (*ent.File, error) {
	return s.files.CreateFolder(ctx, pid, name)
}

// UpdateFile 更新文件元数据。
func (s *ResourcesService) UpdateFile(ctx context.Context, id string, fields map[string]interface{}) (*ent.File, error) {
	return s.files.UpdateMeta(ctx, id, fields)
}

// RegisterAvatars 注册可选头像列表。
func (s *ResourcesService) RegisterAvatars(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := s.files.ListByPID(ctx, registerAvatarFolderID)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, f := range rows {
		if f.IsFolder == 0 {
			out = append(out, fileToMap(f))
		}
	}
	return out, nil
}

// CheckBigFile 大文件分片续传检查。
func (s *ResourcesService) CheckBigFile(ctx context.Context, hash string) (map[string]interface{}, error) {
	if row, err := s.files.FindByFilenameLike(ctx, hash); err == nil {
		return map[string]interface{}{"uploaded": true, "file": fileToMap(row)}, nil
	}
	chunkDir := filepath.Join(s.uploadRoot(), "chunks", hash)
	indices := []int{}
	entries, _ := os.ReadDir(chunkDir)
	for _, e := range entries {
		var idx int
		if _, err := fmt.Sscanf(e.Name(), "%d", &idx); err == nil {
			indices = append(indices, idx)
		}
	}
	return map[string]interface{}{"uploaded": false, "chunkList": indices}, nil
}

// MergeBigFile 合并分片。
func (s *ResourcesService) MergeBigFile(ctx context.Context, hash, fileName string, chunks int) (*ent.File, error) {
	chunkDir := filepath.Join(s.uploadRoot(), "chunks", hash)
	entries, err := os.ReadDir(chunkDir)
	if err != nil {
		return nil, errcode.WithMessage(errcode.InternalError, "文件不存在！")
	}
	if len(entries) != chunks {
		return nil, errcode.WithMessage(errcode.InternalError, "前后切片数量不一致，禁止合并")
	}
	monthDir := time.Now().Format("2006-01")
	targetDir := filepath.Join(s.uploadRoot(), monthDir)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return nil, errcode.WithMessage(errcode.InternalError, "硬盘内存不足了~")
	}
	targetPath := filepath.Join(targetDir, hash+"-"+fileName)
	out, err := os.Create(targetPath)
	if err != nil {
		return nil, err
	}
	defer out.Close()
	for i := 0; i < chunks; i++ {
		partPath := filepath.Join(chunkDir, fmt.Sprintf("%d", i))
		part, err := os.ReadFile(partPath)
		if err != nil {
			return nil, errcode.WithMessage(errcode.InternalError, "文件不存在！")
		}
		if _, err := out.Write(part); err != nil {
			return nil, err
		}
	}
	_ = os.RemoveAll(chunkDir)
	url := blogrepo.NormalizePublicURL(targetPath, s.uploadRoot(), s.publicPrefix())
	return s.files.Create(ctx, &ent.File{
		Pid:          bigFileMergeFolderID,
		IsFolder:     0,
		Originalname: fileName,
		Filename:     hash + "-" + fileName,
		Type:         "application/octet-stream",
		Size:         0,
		URL:          url,
		CreateAt:     time.Now(),
	})
}

// SaveBigFileChunk 保存分片（uploadBigFile）。
func (s *ResourcesService) SaveBigFileChunk(hash string, index int, data []byte) error {
	dir := filepath.Join(s.uploadRoot(), "chunks", hash)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, fmt.Sprintf("%d", index)), data, 0o644)
}

func (s *ResourcesService) uploadRoot() string {
	root := strings.TrimSpace(s.cfg.Storage.UploadPath)
	if root == "" {
		root = "./public/uploads/"
	}
	if !strings.HasSuffix(root, string(os.PathSeparator)) && !strings.HasSuffix(root, "/") {
		root += string(os.PathSeparator)
	}
	return root
}

func (s *ResourcesService) publicPrefix() string {
	p := strings.TrimSpace(s.cfg.Storage.PublicPrefix)
	if p == "" {
		return "/static/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if !strings.HasSuffix(p, "/") {
		p += "/"
	}
	return p
}

func (s *ResourcesService) diskDir(category string) string {
	base := s.uploadRoot()
	switch category {
	case "avatar":
		return filepath.Join(base, "avatar")
	default:
		return filepath.Join(base, time.Now().Format("2006-01"))
	}
}

func resolveFolderID(category string) string {
	switch category {
	case "avatar":
		return registerAvatarFolderID
	case "cover":
		return articleCoverFolderID
	case "article":
		return articleImageFolderID
	default:
		return "0"
	}
}

func fileToMap(f *ent.File) map[string]interface{} {
	return map[string]interface{}{
		"id": f.ID, "pid": f.Pid, "isFolder": f.IsFolder,
		"originalname": f.Originalname, "filename": f.Filename,
		"type": f.Type, "size": f.Size, "url": f.URL, "createAt": f.CreateAt,
	}
}
