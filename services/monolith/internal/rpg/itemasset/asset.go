// Package itemasset RPG 物品 icon/bg 磁盘资产读写（对齐 Nest public/rpgAssets）。
package itemasset

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
)

// Kind 资产类型：itemIcon / itemBg。
type Kind string

const (
	KindIcon Kind = "itemIcon"
	KindBg   Kind = "itemBg"
)

var allowedExt = []string{"png", "webp", "jpg", "jpeg", "svg"}

var mimeToExt = map[string]string{
	"image/png":     "png",
	"image/webp":    "webp",
	"image/jpeg":    "jpg",
	"image/jpg":     "jpg",
	"image/svg+xml": "svg",
}

// SanitizeIconKey 校验 icon 键，防止路径穿越。
func SanitizeIconKey(icon string) (string, error) {
	key := strings.TrimSpace(icon)
	if key == "" || key == "default" {
		return "", errcode.WithMessage(errcode.InvalidParam, "请先填写有效的图标 ID")
	}
	for _, r := range key {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			continue
		}
		return "", errcode.WithMessage(errcode.InvalidParam, "图标 ID 仅允许字母、数字、下划线与连字符")
	}
	return key, nil
}

// ParseKind 解析 assetType 查询参数。
func ParseKind(assetType string) (Kind, error) {
	switch strings.TrimSpace(assetType) {
	case "icon":
		return KindIcon, nil
	case "bg":
		return KindBg, nil
	case "itemIcon":
		return KindIcon, nil
	case "itemBg":
		return KindBg, nil
	default:
		return "", errcode.WithMessage(errcode.InvalidParam, "assetType 须为 icon 或 bg")
	}
}

// DiskDir 返回资产磁盘目录：{uploadRoot}/rpgAssets/{kind}。
func DiskDir(uploadRoot string, kind Kind) string {
	return filepath.Join(strings.TrimRight(uploadRoot, `/\`), "rpgAssets", string(kind))
}

// StaticURL 返回静态访问路径。
func StaticURL(staticPrefix string, kind Kind, iconKey, ext string) string {
	prefix := strings.TrimRight(staticPrefix, "/")
	if prefix == "" {
		prefix = "/static"
	}
	return fmt.Sprintf("%s/rpgAssets/%s/%s.%s", prefix, kind, iconKey, ext)
}

// Save 保存上传文件，同键覆盖旧扩展名。
func Save(uploadRoot, staticPrefix string, kind Kind, iconKey string, data []byte, filename, contentType string) (map[string]interface{}, error) {
	if len(data) == 0 {
		return nil, errcode.WithMessage(errcode.InvalidParam, "未收到上传文件")
	}
	ext, err := resolveExt(filename, contentType)
	if err != nil {
		return nil, err
	}
	dir := DiskDir(uploadRoot, kind)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	for _, oldExt := range allowedExt {
		_ = os.Remove(filepath.Join(dir, iconKey+"."+oldExt))
	}
	target := filepath.Join(dir, iconKey+"."+ext)
	if err := os.WriteFile(target, data, 0o644); err != nil {
		return nil, err
	}
	return map[string]interface{}{"url": StaticURL(staticPrefix, kind, iconKey, ext)}, nil
}

// Delete 删除同键所有扩展名资产文件。
func Delete(uploadRoot string, kind Kind, iconKey string) (map[string]interface{}, error) {
	dir := DiskDir(uploadRoot, kind)
	deleted := false
	for _, ext := range allowedExt {
		p := filepath.Join(dir, iconKey+"."+ext)
		if err := os.Remove(p); err == nil {
			deleted = true
		}
	}
	return map[string]interface{}{"deleted": deleted}, nil
}

func resolveExt(filename, contentType string) (string, error) {
	if ext, ok := mimeToExt[strings.ToLower(contentType)]; ok {
		return ext, nil
	}
	raw := strings.TrimPrefix(strings.ToLower(filepath.Ext(filename)), ".")
	if raw == "jpeg" {
		raw = "jpg"
	}
	for _, e := range allowedExt {
		if raw == e || (raw == "jpeg" && e == "jpg") {
			if raw == "jpeg" {
				return "jpg", nil
			}
			return raw, nil
		}
	}
	return "", errcode.WithMessage(errcode.InvalidParam, "仅支持 PNG / WebP / JPG / SVG 图片")
}
