package notify

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

type mockPusher struct {
	msgType string
	uid     uint64
	data    json.RawMessage
}

func (m *mockPusher) PushToUser(_ context.Context, uid uint64, msgType string, _ uint64, data interface{}) error {
	raw, _ := json.Marshal(data)
	m.uid = uid
	m.msgType = msgType
	m.data = raw
	return nil
}

func TestNotifyBanStatusUnban(t *testing.T) {
	p := &mockPusher{}
	svc := NewRpgNotifyService(p, nil, nil)
	svc.NotifyBanStatus(context.Background(), 42, false, nil, nil)
	if p.msgType != msgBanStatus || p.uid != 42 {
		t.Fatalf("type=%s uid=%d", p.msgType, p.uid)
	}
	var payload BanStatusPayload
	if err := json.Unmarshal(p.data, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Banned || payload.BanEndTime != nil || payload.BanReason != nil {
		t.Fatalf("payload=%+v", payload)
	}
}

func TestNotifyBanStatusBanned(t *testing.T) {
	p := &mockPusher{}
	svc := NewRpgNotifyService(p, nil, nil)
	end := time.Now().Add(72 * time.Hour)
	reason := "test"
	svc.NotifyBanStatus(context.Background(), 7, true, &end, &reason)
	var payload BanStatusPayload
	if err := json.Unmarshal(p.data, &payload); err != nil {
		t.Fatal(err)
	}
	if !payload.Banned || payload.BanReason == nil || *payload.BanReason != reason {
		t.Fatalf("payload=%+v", payload)
	}
}

func TestNotifyShieldUsed(t *testing.T) {
	p := &mockPusher{}
	svc := NewRpgNotifyService(p, nil, nil)
	svc.NotifyShieldUsed(context.Background(), 3)
	if p.msgType != msgShieldUsed {
		t.Fatalf("type=%s", p.msgType)
	}
}
