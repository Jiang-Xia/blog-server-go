package repo

import (
	"context"
	"strings"
	"time"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/file"
	"github.com/google/uuid"
)

// FileRepo 文件资源表读写（x_file）。
type FileRepo struct {
	client *ent.Client
}

// NewFileRepo 构造 FileRepo。
func NewFileRepo(client *ent.Client) *FileRepo {
	return &FileRepo{client: client}
}

// FileFilter 文件列表筛选。
type FileFilter struct {
	PID      string
	Page     int
	PageSize int
}

// Create 写入文件记录。
func (r *FileRepo) Create(ctx context.Context, row *ent.File) (*ent.File, error) {
	if row.ID == "" {
		row.ID = uuid.NewString()
	}
	b := r.client.File.Create().
		SetID(row.ID).
		SetPid(row.Pid).
		SetIsFolder(row.IsFolder).
		SetOriginalname(row.Originalname).
		SetFilename(row.Filename).
		SetType(row.Type).
		SetSize(row.Size).
		SetURL(row.URL).
		SetCreateAt(row.CreateAt)
	return b.Save(ctx)
}

// GetByID 按 id 查询。
func (r *FileRepo) GetByID(ctx context.Context, id string) (*ent.File, error) {
	return r.client.File.Query().Where(file.IDEQ(id)).Only(ctx)
}

// FindByFilenameLike 按文件名模糊查重（大文件 hash）。
func (r *FileRepo) FindByFilenameLike(ctx context.Context, hash string) (*ent.File, error) {
	return r.client.File.Query().
		Where(file.FilenameContains(hash)).
		First(ctx)
}

// List 分页文件列表。
func (r *FileRepo) List(ctx context.Context, f FileFilter) ([]*ent.File, int, error) {
	page, pageSize := f.Page, f.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	q := r.client.File.Query()
	if f.PID != "" {
		q = q.Where(file.PidEQ(f.PID))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.Order(ent.Desc(file.FieldCreateAt)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	return rows, total, err
}

// ListByPID 目录下全部文件。
func (r *FileRepo) ListByPID(ctx context.Context, pid string) ([]*ent.File, error) {
	return r.client.File.Query().Where(file.PidEQ(pid)).All(ctx)
}

// DeleteByID 删除文件记录。
func (r *FileRepo) DeleteByID(ctx context.Context, id string) error {
	return r.client.File.DeleteOneID(id).Exec(ctx)
}

// UpdateMeta 更新文件元数据。
func (r *FileRepo) UpdateMeta(ctx context.Context, id string, fields map[string]interface{}) (*ent.File, error) {
	up := r.client.File.UpdateOneID(id)
	if v, ok := fields["originalname"].(string); ok {
		up.SetOriginalname(v)
	}
	if v, ok := fields["filename"].(string); ok {
		up.SetFilename(v)
	}
	if v, ok := fields["url"].(string); ok {
		up.SetURL(v)
	}
	if v, ok := fields["pid"].(string); ok {
		up.SetPid(v)
	}
	return up.Save(ctx)
}

// CreateFolder 创建虚拟文件夹。
func (r *FileRepo) CreateFolder(ctx context.Context, pid, name string) (*ent.File, error) {
	now := time.Now()
	return r.Create(ctx, &ent.File{
		ID:           uuid.NewString(),
		Pid:          pid,
		IsFolder:     1,
		Originalname: name,
		Filename:     name,
		Type:         "folder",
		Size:         0,
		URL:          "",
		CreateAt:     now,
	})
}

// NormalizePublicURL 将磁盘路径转为 /static/ URL。
func NormalizePublicURL(diskPath, uploadRoot, publicPrefix string) string {
	rel := strings.TrimPrefix(diskPath, uploadRoot)
	rel = strings.ReplaceAll(rel, "\\", "/")
	if !strings.HasPrefix(rel, "/") {
		rel = "/" + rel
	}
	return publicPrefix + strings.TrimPrefix(rel, "/")
}
