package rag_test

import (
	"context"
	"fmt"
	"testing"
	"unicode/utf8"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/article"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rag"
	_ "github.com/go-sql-driver/mysql"
)

// 集成探测：对已发布文章分块后不得产生无效 UTF-8（MySQL 1366）。
func TestChunkOutputValidUTF8PublishedArticles(t *testing.T) {
	cfg, err := config.MustLoad("../../../../configs/monolith.yaml")
	if err != nil {
		t.Skip(err)
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4",
		cfg.MySQL.User, cfg.MySQL.Password, cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.Database)
	client, err := ent.Open("mysql", dsn)
	if err != nil {
		t.Skip(err)
	}
	defer client.Close()
	chunkSvc := rag.NewChunkService(cfg)
	ctx := context.Background()
	arts, err := client.Article.Query().
		Where(article.IsDeleteEQ(false), article.StatusEQ("publish")).
		All(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range arts {
		pieces := chunkSvc.SplitMarkdown(a.Content, a.Title, a.Description)
		for i, p := range pieces {
			if !utf8.ValidString(p.Content) {
				t.Errorf("invalid UTF-8 article=%d chunk=%d title=%q", a.ID, i, a.Title)
			}
		}
	}
	for _, def := range rag.RAGStaticPages {
		md, err := rag.LoadStaticPageMarkdown(def)
		if err != nil {
			t.Fatal(err)
		}
		pieces := chunkSvc.SplitMarkdown(md, def.Title, def.Description)
		for i, p := range pieces {
			if !utf8.ValidString(p.Content) {
				t.Errorf("invalid UTF-8 static=%s chunk=%d", def.Slug, i)
			}
		}
	}
}
