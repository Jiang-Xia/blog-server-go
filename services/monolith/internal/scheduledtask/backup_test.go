package scheduledtask

import (
	"testing"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
)

func TestValidateBackupFileName(t *testing.T) {
	cases := []struct {
		name    string
		file    string
		wantErr bool
	}{
		{"valid", "myblog_backup_20260703_120000.sql", false},
		{"path traversal", "../myblog_backup_x.sql", true},
		{"wrong prefix", "backup.sql", true},
		{"wrong ext", "myblog_backup_2026.txt", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateBackupFileName(tc.file)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateBackupFileName(%q) err=%v wantErr=%v", tc.file, err, tc.wantErr)
			}
		})
	}
}

func TestResolveBackupPathRejectsTraversal(t *testing.T) {
	cfg := &config.Config{}
	_, err := ResolveBackupPath(cfg, "../../etc/passwd")
	if err == nil {
		t.Fatal("expected error for invalid filename")
	}
}

func TestSeedTasksCount(t *testing.T) {
	if len(SeedTasks) != 8 {
		t.Fatalf("expected 8 seed tasks, got %d", len(SeedTasks))
	}
	names := map[string]bool{}
	for _, s := range SeedTasks {
		if s.Cron == "" || s.Name == "" {
			t.Fatalf("invalid seed: %+v", s)
		}
		names[s.Name] = true
	}
	for _, required := range []string{"scheduled_publish", "database_backup", "daily_interaction_notify"} {
		if !names[required] {
			t.Fatalf("missing seed %s", required)
		}
	}
}
