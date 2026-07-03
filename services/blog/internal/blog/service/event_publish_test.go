package service

import (
	"context"
	"testing"
	"time"

	"github.com/Jiang-Xia/blog-server-go/services/blog/ent"
	blogevent "github.com/Jiang-Xia/blog-server-go/services/blog/internal/event"
)

type mockEventPublisher struct {
	calls []struct {
		Type    string
		Payload interface{}
	}
}

func (m *mockEventPublisher) Publish(_ context.Context, eventType string, payload interface{}) {
	m.calls = append(m.calls, struct {
		Type    string
		Payload interface{}
	}{Type: eventType, Payload: payload})
}

func TestPublishSensitiveWordHitSkipsGuest(t *testing.T) {
	pub := &mockEventPublisher{}
	publishSensitiveWordHit(context.Background(), pub, 0, 10)
	if len(pub.calls) != 0 {
		t.Fatalf("guest should not publish, got %d calls", len(pub.calls))
	}
}

func TestPublishSensitiveWordHitPublishes(t *testing.T) {
	pub := &mockEventPublisher{}
	publishSensitiveWordHit(context.Background(), pub, 42, 5)
	if len(pub.calls) != 1 {
		t.Fatalf("calls=%d want 1", len(pub.calls))
	}
	if pub.calls[0].Type != blogevent.EventSensitiveWordHit {
		t.Fatalf("type=%q", pub.calls[0].Type)
	}
	p, ok := pub.calls[0].Payload.(blogevent.SensitiveWordHitPayload)
	if !ok || p.UID != 42 || p.HpPenalty != 5 {
		t.Fatalf("payload=%+v", pub.calls[0].Payload)
	}
}

func TestPublishArticleLifecyclePublished(t *testing.T) {
	pub := &mockEventPublisher{}
	now := time.Now()
	article := &ent.Article{ID: 7, UID: 3, Status: "publish", CreateTime: now, UpdateTime: now}
	publishArticleLifecycleEvents(context.Background(), pub, article, "draft", false)
	if len(pub.calls) != 1 || pub.calls[0].Type != blogevent.EventArticlePublished {
		t.Fatalf("calls=%+v", pub.calls)
	}
}

func TestPublishArticleLifecycleUpdated(t *testing.T) {
	pub := &mockEventPublisher{}
	now := time.Now()
	article := &ent.Article{ID: 7, UID: 3, Status: "publish", CreateTime: now, UpdateTime: now}
	publishArticleLifecycleEvents(context.Background(), pub, article, "publish", false)
	if len(pub.calls) != 1 || pub.calls[0].Type != blogevent.EventArticleUpdated {
		t.Fatalf("calls=%+v", pub.calls)
	}
}

func TestPublishArticleLifecycleDeleted(t *testing.T) {
	pub := &mockEventPublisher{}
	now := time.Now()
	article := &ent.Article{ID: 7, UID: 3, Status: "publish", IsDelete: true, CreateTime: now, UpdateTime: now}
	publishArticleLifecycleEvents(context.Background(), pub, article, "publish", false)
	if len(pub.calls) != 1 || pub.calls[0].Type != blogevent.EventArticleDeleted {
		t.Fatalf("calls=%+v", pub.calls)
	}
}

func TestPublishArticleLifecycleUnpublished(t *testing.T) {
	pub := &mockEventPublisher{}
	now := time.Now()
	article := &ent.Article{ID: 7, UID: 3, Status: "draft", CreateTime: now, UpdateTime: now}
	publishArticleLifecycleEvents(context.Background(), pub, article, "publish", false)
	if len(pub.calls) != 1 || pub.calls[0].Type != blogevent.EventArticleUnpublished {
		t.Fatalf("calls=%+v", pub.calls)
	}
}
