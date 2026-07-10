package scheduledtask

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
)

var backupNamePattern = regexp.MustCompile(`^myblog_backup_[\w-]+\.sql$`)

// BackupFileInfo 备份文件元信息（对齐 Nest DatabaseBackupFileInfo）。
type BackupFileInfo struct {
	FileName      string `json:"fileName"`
	FileSize      string `json:"fileSize"`
	FileSizeBytes int64  `json:"fileSizeBytes"`
	CreatedAt     string `json:"createdAt"`
}

// backupDir 返回备份目录绝对路径。
func backupDir(cfg *config.Config) string {
	if d := cfg.Backup.Dir; d != "" {
		return filepath.Clean(d)
	}
	wd, _ := os.Getwd()
	return filepath.Join(wd, "backups")
}

// ValidateBackupFileName 校验备份文件名白名单（防路径穿越）。
func ValidateBackupFileName(fileName string) error {
	if !backupNamePattern.MatchString(fileName) {
		return fmt.Errorf("无效的备份文件名")
	}
	return nil
}

// ResolveBackupPath 解析备份文件绝对路径。
func ResolveBackupPath(cfg *config.Config, fileName string) (string, error) {
	if err := ValidateBackupFileName(fileName); err != nil {
		return "", err
	}
	dir := backupDir(cfg)
	path := filepath.Join(dir, fileName)
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if !stringsHasPrefix(absPath, absDir+string(os.PathSeparator)) && absPath != absDir {
		return "", fmt.Errorf("无效的备份文件路径")
	}
	if _, err := os.Stat(absPath); err != nil {
		return "", fmt.Errorf("备份文件不存在")
	}
	return absPath, nil
}

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// ListBackupFiles 列出备份目录下合法 sql 文件。
func ListBackupFiles(cfg *config.Config) ([]BackupFileInfo, error) {
	dir := backupDir(cfg)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return []BackupFileInfo{}, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := make([]BackupFileInfo, 0)
	for _, e := range entries {
		if e.IsDir() || !backupNamePattern.MatchString(e.Name()) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, BackupFileInfo{
			FileName:      e.Name(),
			FileSize:      formatBackupSize(info.Size()),
			FileSizeBytes: info.Size(),
			CreatedAt:     info.ModTime().UTC().Format(time.RFC3339),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt > out[j].CreatedAt
	})
	return out, nil
}

func formatBackupSize(bytes int64) string {
	if bytes > 1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(bytes)/1024/1024)
	}
	return fmt.Sprintf("%.2f KB", float64(bytes)/1024)
}
