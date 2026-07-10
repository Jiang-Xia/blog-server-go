package rag_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
)

func TestEntRagIndexJobJSONCamelCase(t *testing.T) {
	errMsg := "err"
	job := &ent.RagIndexJob{
		ID: 1, ArticleID: 0, Status: "success", ChunkCount: 9,
		ErrorMsg: &errMsg, CreateAt: time.Now(), UpdateAt: time.Now(),
	}
	b, err := json.Marshal(job)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"articleId", "chunkCount", "errorMsg", "createAt", "updateAt"} {
		if _, ok := m[key]; !ok {
			t.Fatalf("missing %q in %s", key, string(b))
		}
	}
}

func TestEntRagQueryLogJSONCamelCase(t *testing.T) {
	log := &ent.RagQueryLog{
		ID: 1, UID: 2, Question: "q", LatencyMs: 10, Status: "success", CreateAt: time.Now(),
	}
	b, _ := json.Marshal(log)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	for _, key := range []string{"latencyMs", "createAt"} {
		if _, ok := m[key]; !ok {
			t.Fatalf("missing %q in %s", key, string(b))
		}
	}
}
