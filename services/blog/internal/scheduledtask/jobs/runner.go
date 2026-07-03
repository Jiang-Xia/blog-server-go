// Package jobs 8 个内置定时任务业务实现。
package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/usersvc"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent/article"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent/comment"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent/link"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent/msgboard"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent/reply"
	blogrepo "github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/repo"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/crossdb"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/event"
)

// Runner 任务执行器，按 taskName 分发到具体 job。
type Runner struct {
	client   *ent.Client
	cfg      *config.Config
	articles *blogrepo.ArticleRepo
	links    *blogrepo.LinkRepo
	cross    *crossdb.CrossDB
	email    usersvc.SystemEmailSender
	events   *event.Publisher
	http     *http.Client
}

// NewRunner 构造 Runner。
func NewRunner(
	client *ent.Client,
	cfg *config.Config,
	articles *blogrepo.ArticleRepo,
	links *blogrepo.LinkRepo,
	cross *crossdb.CrossDB,
	email usersvc.SystemEmailSender,
	events *event.Publisher,
) *Runner {
	return &Runner{
		client: client, cfg: cfg, articles: articles, links: links,
		cross: cross, email: email, events: events,
		http: &http.Client{Timeout: 10 * time.Second},
	}
}

// Execute 按任务名执行并返回 JSON 可序列化结果。
func (r *Runner) Execute(ctx context.Context, taskName string) (interface{}, error) {
	switch taskName {
	case "daily_interaction_notify":
		return r.runDailyInteractionNotify(ctx)
	case "scheduled_publish":
		return r.runScheduledPublish(ctx)
	case "monthly_report":
		return r.runMonthlyReport(ctx)
	case "link_health_check":
		return r.runLinkHealthCheck(ctx)
	case "expired_data_cleanup":
		return r.runExpiredDataCleanup(ctx)
	case "database_backup":
		return r.runDatabaseBackup(ctx)
	case "draft_cleanup":
		return r.runDraftCleanup(ctx)
	case "sensitive_word_alert":
		return r.runSensitiveWordAlert(ctx)
	default:
		return map[string]string{"message": fmt.Sprintf("任务 [%s] 无对应的执行处理器", taskName)}, nil
	}
}

func (r *Runner) runDailyInteractionNotify(ctx context.Context) (interface{}, error) {
	now := time.Now()
	yStart := dayStart(now.AddDate(0, 0, -1))
	yEnd := dayEnd(now.AddDate(0, 0, -1))
	pStart := dayStart(now.AddDate(0, 0, -2))
	pEnd := dayEnd(now.AddDate(0, 0, -2))

	commentCount, _ := r.countBetween(ctx, "comment", yStart, yEnd)
	replyCount, _ := r.countBetween(ctx, "reply", yStart, yEnd)
	msgboardCount, _ := r.countBetween(ctx, "msgboard", yStart, yEnd)
	prevComment, _ := r.countBetween(ctx, "comment", pStart, pEnd)
	prevReply, _ := r.countBetween(ctx, "reply", pStart, pEnd)
	prevMsgboard, _ := r.countBetween(ctx, "msgboard", pStart, pEnd)

	dateStr := yStart.Format("2006-01-02")
	hasChange := commentCount != prevComment || replyCount != prevReply || msgboardCount != prevMsgboard
	emailSent := false
	if hasChange && r.email != nil {
		subject := fmt.Sprintf("📊 每日互动日报 - %s", dateStr)
		body := buildDailyInteractionHTML(dateStr, commentCount, replyCount, msgboardCount, prevComment, prevReply, prevMsgboard)
		sent, err := r.email.SendSystemEmail(ctx, "", subject, body)
		if err != nil {
			return nil, err
		}
		emailSent = sent
	}
	return map[string]interface{}{
		"date": dateStr, "commentCount": commentCount, "replyCount": replyCount,
		"msgboardCount": msgboardCount, "prevCommentCount": prevComment,
		"prevReplyCount": prevReply, "prevMsgboardCount": prevMsgboard, "emailSent": emailSent,
	}, nil
}

func (r *Runner) runScheduledPublish(ctx context.Context) (interface{}, error) {
	now := time.Now()
	rows, err := r.client.Article.Query().
		Where(
			article.StatusEQ("scheduled"),
			article.ScheduledPublishAtNotNil(),
			article.ScheduledPublishAtLTE(now),
			article.IsDeleteEQ(false),
		).All(ctx)
	if err != nil {
		return nil, err
	}
	published := 0
	for _, row := range rows {
		_, err := r.client.Article.UpdateOneID(row.ID).
			SetStatus("publish").
			ClearScheduledPublishAt().
			SetUTime(now.Format(time.RFC3339)).
			Save(ctx)
		if err != nil {
			return nil, err
		}
		if r.events != nil {
			r.events.Publish(ctx, event.EventArticlePublished, event.ArticlePublishedPayload{
				UID: row.UID, ArticleID: row.ID,
			})
		}
		published++
	}
	return map[string]int{"publishedCount": published}, nil
}

func (r *Runner) runMonthlyReport(ctx context.Context) (interface{}, error) {
	lastMonth := time.Now().AddDate(0, -1, 0)
	mStart := time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, time.Local)
	mEnd := mStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
	monthStr := mStart.Format("2006-01")

	rows, err := r.client.Article.Query().
		Where(
			article.StatusEQ("publish"),
			article.CreateTimeGTE(mStart),
			article.CreateTimeLTE(mEnd),
			article.IsDeleteEQ(false),
		).All(ctx)
	if err != nil {
		return nil, err
	}
	totalViews, totalLikes := 0, 0
	for _, a := range rows {
		totalViews += a.Views
		totalLikes += a.Likes
	}
	totalComments, _ := r.client.Comment.Query().
		Where(comment.CreateTimeGTE(mStart), comment.CreateTimeLTE(mEnd)).
		Count(ctx)

	top := append([]*ent.Article{}, rows...)
	sort.Slice(top, func(i, j int) bool { return top[i].Views > top[j].Views })
	if len(top) > 5 {
		top = top[:5]
	}
	topArticles := make([]map[string]interface{}, 0, len(top))
	for _, a := range top {
		topArticles = append(topArticles, map[string]interface{}{"title": a.Title, "views": a.Views})
	}

	emailSent := false
	if r.email != nil {
		subject := fmt.Sprintf("📅 博客月报 - %s", monthStr)
		body := buildMonthlyReportHTML(monthStr, len(rows), totalViews, totalLikes, totalComments, topArticles)
		sent, err := r.email.SendSystemEmail(ctx, "", subject, body)
		if err != nil {
			return nil, err
		}
		emailSent = sent
	}
	return map[string]interface{}{
		"month": monthStr, "articleCount": len(rows), "totalViews": totalViews,
		"totalLikes": totalLikes, "totalComments": totalComments, "emailSent": emailSent,
	}, nil
}

func (r *Runner) runLinkHealthCheck(ctx context.Context) (interface{}, error) {
	links, err := r.client.Link.Query().Where(link.AgreedEQ(1)).All(ctx)
	if err != nil {
		return nil, err
	}
	details := make([]map[string]string, 0, len(links))
	okCount, downCount := 0, 0
	now := time.Now()
	for _, l := range links {
		status := "down"
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, l.URL, nil)
		if resp, err := r.http.Do(req); err == nil {
			resp.Body.Close()
			if resp.StatusCode < 400 {
				status = "ok"
			}
		}
		_, _ = r.client.Link.UpdateOneID(l.ID).
			SetLastCheckStatus(status).
			SetLastCheckTime(now).
			Save(ctx)
		details = append(details, map[string]string{"title": l.Title, "url": l.URL, "status": status})
		if status == "ok" {
			okCount++
		} else {
			downCount++
		}
	}
	emailSent := false
	if downCount > 0 && r.email != nil {
		body := buildLinkHealthHTML(details, len(details), downCount)
		subject := fmt.Sprintf("🔗 友链健康检测报告 - %d 个不可达", downCount)
		sent, err := r.email.SendSystemEmail(ctx, "", subject, body)
		if err != nil {
			return nil, err
		}
		emailSent = sent
	}
	return map[string]interface{}{
		"total": len(details), "okCount": okCount, "downCount": downCount,
		"details": details, "emailSent": emailSent,
	}, nil
}

func (r *Runner) runExpiredDataCleanup(ctx context.Context) (interface{}, error) {
	opCutoff := time.Now().AddDate(0, 0, -90)
	logCutoff := time.Now().AddDate(0, 0, -180)
	opDeleted, taskLogDeleted := 0, 0
	var err error
	if r.cross != nil {
		opDeleted, err = r.cross.DeleteOldOperationLogs(ctx, opCutoff)
		if err != nil {
			return nil, err
		}
		taskLogDeleted, err = r.cross.DeleteOldTaskLogs(ctx, logCutoff)
		if err != nil {
			return nil, err
		}
	}
	return map[string]int{"operationLogDeleted": opDeleted, "taskLogDeleted": taskLogDeleted}, nil
}

func (r *Runner) runDatabaseBackup(ctx context.Context) (interface{}, error) {
	_ = ctx
	dir := scheduledtaskBackupDir(r.cfg)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	ts := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("myblog_backup_%s.sql", ts)
	backupPath := filepath.Join(dir, fileName)

	m := r.cfg.MySQL
	dump := r.cfg.Backup.MysqldumpPath
	if dump == "" {
		dump = "mysqldump"
	}
	host, port, user, pass, db := m.Host, m.Port, m.User, m.Password, m.Database
	if port == 0 {
		port = 3306
	}
	outFile, err := os.Create(backupPath)
	if err != nil {
		return nil, err
	}
	defer outFile.Close()

	args := []string{
		fmt.Sprintf("-h%s", host),
		fmt.Sprintf("-P%d", port),
		fmt.Sprintf("-u%s", user),
		fmt.Sprintf("-p%s", pass),
		db,
	}
	cmd := exec.Command(dump, args...)
	cmd.Stdout = outFile
	if err := cmd.Run(); err != nil {
		_ = os.Remove(backupPath)
		return nil, fmt.Errorf("mysqldump failed: %w", err)
	}
	info, err := os.Stat(backupPath)
	if err != nil {
		return nil, err
	}
	cutoff := time.Now().AddDate(0, 0, -30)
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), "myblog_backup_") || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		fi, err := e.Info()
		if err != nil || fi.ModTime().After(cutoff) {
			continue
		}
		_ = os.Remove(filepath.Join(dir, e.Name()))
	}
	return map[string]string{
		"fileName": fileName, "fileSize": scheduledtaskFormatSize(info.Size()), "backupPath": backupPath,
	}, nil
}

func (r *Runner) runDraftCleanup(ctx context.Context) (interface{}, error) {
	cutoff := time.Now().AddDate(0, 0, -90)
	rows, err := r.client.Article.Query().
		Where(
			article.StatusEQ("draft"),
			article.IsDeleteEQ(false),
			article.UpdateTimeLT(cutoff),
		).
		Select(article.FieldID, article.FieldTitle, article.FieldUpdateTime).
		All(ctx)
	if err != nil {
		return nil, err
	}
	drafts := make([]map[string]interface{}, 0, len(rows))
	for _, d := range rows {
		drafts = append(drafts, map[string]interface{}{"id": d.ID, "title": d.Title, "updateTime": d.UpdateTime})
		_, err := r.client.Article.UpdateOneID(d.ID).SetIsDelete(true).Save(ctx)
		if err != nil {
			return nil, err
		}
	}
	return map[string]interface{}{"cleanedCount": len(rows), "drafts": drafts}, nil
}

func (r *Runner) runSensitiveWordAlert(ctx context.Context) (interface{}, error) {
	if r.cross == nil {
		return map[string]interface{}{"totalHits": 0, "emailSent": false}, nil
	}
	hits, err := r.cross.QueryPendingSensitiveHits(ctx)
	if err != nil {
		total, countErr := r.cross.CountPendingSensitiveHits(ctx)
		if countErr != nil {
			return nil, countErr
		}
		return map[string]interface{}{"totalHits": total, "emailSent": false}, nil
	}
	totalHits := len(hits)
	emailSent := false
	if totalHits > 0 && r.email != nil {
		wordMap := map[string]int{}
		sourceMap := map[string]int{}
		for _, h := range hits {
			for _, w := range strings.Split(h.HitWords, ",") {
				w = strings.TrimSpace(w)
				if w != "" {
					wordMap[w]++
				}
			}
			sourceMap[h.SourceType]++
		}
		topWords := topWordEntries(wordMap, 10)
		sourceBreakdown := sourceEntries(sourceMap)
		body := buildSensitiveWordHTML(totalHits, topWords, sourceBreakdown)
		sent, err := r.email.SendSystemEmail(ctx, "", "⚠️ 敏感词命中告警", body)
		if err != nil {
			return nil, err
		}
		emailSent = sent
	}
	return map[string]interface{}{"totalHits": totalHits, "emailSent": emailSent}, nil
}

func (r *Runner) countBetween(ctx context.Context, kind string, start, end time.Time) (int, error) {
	switch kind {
	case "comment":
		return r.client.Comment.Query().
			Where(comment.CreateTimeGTE(start), comment.CreateTimeLTE(end)).Count(ctx)
	case "reply":
		return r.client.Reply.Query().
			Where(reply.CreateTimeGTE(start), reply.CreateTimeLTE(end)).Count(ctx)
	case "msgboard":
		return r.client.Msgboard.Query().
			Where(msgboard.CreateTimeGTE(start), msgboard.CreateTimeLTE(end)).Count(ctx)
	default:
		return 0, nil
	}
}

func dayStart(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func dayEnd(t time.Time) time.Time {
	return dayStart(t).Add(24*time.Hour - time.Nanosecond)
}

func scheduledtaskBackupDir(cfg *config.Config) string {
	if d := cfg.Backup.Dir; d != "" {
		return filepath.Clean(d)
	}
	wd, _ := os.Getwd()
	return filepath.Join(wd, "backups")
}

func scheduledtaskFormatSize(n int64) string {
	if n > 1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(n)/1024/1024)
	}
	return fmt.Sprintf("%.2f KB", float64(n)/1024)
}

func topWordEntries(m map[string]int, limit int) []map[string]interface{} {
	type pair struct {
		word  string
		count int
	}
	arr := make([]pair, 0, len(m))
	for w, c := range m {
		arr = append(arr, pair{w, c})
	}
	sort.Slice(arr, func(i, j int) bool { return arr[i].count > arr[j].count })
	if len(arr) > limit {
		arr = arr[:limit]
	}
	out := make([]map[string]interface{}, 0, len(arr))
	for _, p := range arr {
		out = append(out, map[string]interface{}{"word": p.word, "count": p.count})
	}
	return out
}

func sourceEntries(m map[string]int) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(m))
	for k, v := range m {
		out = append(out, map[string]interface{}{"sourceType": k, "count": v})
	}
	return out
}

func buildDailyInteractionHTML(date string, c, r, m, pc, pr, pm int) string {
	diff := func(cur, prev int) string {
		d := cur - prev
		if d > 0 {
			return fmt.Sprintf(`<span style="color:#52c41a">+%d</span>`, d)
		}
		if d < 0 {
			return fmt.Sprintf(`<span style="color:#f5222d">%d</span>`, d)
		}
		return `<span style="color:#999">0</span>`
	}
	return fmt.Sprintf(`<div style="max-width:600px;margin:0 auto;padding:20px;font-family:Arial,sans-serif;">
<h2>📊 每日互动数据日报</h2><p>日期：%s</p>
<table style="width:100%%;border-collapse:collapse;">
<tr><th>类型</th><th>昨日</th><th>前日</th><th>变化</th></tr>
<tr><td>评论</td><td>%d</td><td>%d</td><td>%s</td></tr>
<tr><td>回复</td><td>%d</td><td>%d</td><td>%s</td></tr>
<tr><td>留言</td><td>%d</td><td>%d</td><td>%s</td></tr>
</table></div>`, date, c, pc, diff(c, pc), r, pr, diff(r, pr), m, pm, diff(m, pm))
}

func buildMonthlyReportHTML(month string, articleCount, views, likes, comments int, top []map[string]interface{}) string {
	rows := ""
	for i, a := range top {
		rows += fmt.Sprintf("<tr><td>%d</td><td>%v</td><td>%v</td></tr>", i+1, a["title"], a["views"])
	}
	return fmt.Sprintf(`<div style="max-width:600px;margin:0 auto;padding:20px;">
<h2>📅 %s 博客月报</h2>
<p>新增文章：%d | 阅读：%d | 点赞：%d | 评论：%d</p>
<table>%s</table></div>`, month, articleCount, views, likes, comments, rows)
}

func buildLinkHealthHTML(details []map[string]string, total, down int) string {
	rows := ""
	for _, d := range details {
		rows += fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td></tr>", d["title"], d["url"], d["status"])
	}
	return fmt.Sprintf(`<div><h2>🔗 友链健康检测</h2><p>共 %d 个，%d 不可达</p><table>%s</table></div>`, total, down, rows)
}

func buildSensitiveWordHTML(total int, topWords, sources []map[string]interface{}) string {
	raw, _ := json.Marshal(map[string]interface{}{
		"totalHits": total, "topWords": topWords, "sourceBreakdown": sources,
	})
	return fmt.Sprintf(`<div><h2>⚠️ 敏感词告警</h2><pre>%s</pre></div>`, string(raw))
}
