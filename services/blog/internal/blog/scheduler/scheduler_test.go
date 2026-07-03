package scheduler

import (
	"testing"

	"github.com/robfig/cron/v3"
)

// TestSeedCronExpressionsParse 8 个种子 cron 须可被 robfig/cron（含秒）解析。
func TestSeedCronExpressionsParse(t *testing.T) {
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	exprs := []string{
		"0 0 10 * * *",
		"0 * * * * *",
		"0 0 10 1 * *",
		"0 0 3 * * 1",
		"0 30 2 * * *",
		"0 0 4 * * *",
		"0 0 5 * * *",
		"0 0 9 * * *",
	}
	for _, e := range exprs {
		if _, err := parser.Parse(e); err != nil {
			t.Fatalf("parse cron %q: %v", e, err)
		}
	}
}

func TestRegisterTaskInvalidCron(t *testing.T) {
	s := New(nil)
	if err := s.RegisterTask("bad", "not a cron"); err == nil {
		t.Fatal("expected error for invalid cron")
	}
}
