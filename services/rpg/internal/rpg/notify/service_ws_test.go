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

func TestNotifyQuestComplete(t *testing.T) {
	p := &mockPusher{}
	svc := NewRpgNotifyService(p, nil, nil)
	hp := 5
	svc.NotifyQuestComplete(context.Background(), 9, QuestCompletePayload{
		QuestCode: "daily_comment",
		QuestName: "每日评论",
		ExpReward: 10,
		HpReward:  &hp,
	})
	if p.msgType != msgQuestComplete || p.uid != 9 {
		t.Fatalf("type=%s uid=%d", p.msgType, p.uid)
	}
	var payload QuestCompletePayload
	if err := json.Unmarshal(p.data, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.QuestCode != "daily_comment" || payload.HpReward == nil || *payload.HpReward != 5 {
		t.Fatalf("payload=%+v", payload)
	}
}

func TestNotifyAchievementComplete(t *testing.T) {
	p := &mockPusher{}
	svc := NewRpgNotifyService(p, nil, nil)
	svc.NotifyAchievementComplete(context.Background(), 1, AchievementCompletePayload{
		Code: "first_comment", Name: "初次评论", ExpReward: 20,
		Rarity: "common", RarityLabel: "普通",
	})
	if p.msgType != msgAchievementComplete {
		t.Fatalf("type=%s", p.msgType)
	}
}

func TestNotifySocialReceivedWithLifeChange(t *testing.T) {
	p := &mockPusher{}
	calls := []string{}
	mp := &multiPusher{p: p, types: &calls}
	svc := NewRpgNotifyService(mp, nil, nil)
	svc.NotifySocialReceived(context.Background(), 2, SocialReceivedPayload{
		FromUID: 1, Action: "egg", HpDelta: -10, CurrentLife: 70,
	})
	if len(calls) != 2 || calls[0] != msgSocialReceived || calls[1] != msgLifeChange {
		t.Fatalf("calls=%v", calls)
	}
}

func TestNotifyArticleLevelUp(t *testing.T) {
	p := &mockPusher{}
	svc := NewRpgNotifyService(p, nil, nil)
	svc.NotifyArticleLevelUp(context.Background(), 5, ArticleLevelUpPayload{
		ArticleID: 100, ArticleTitle: "测试", OldLevel: 1, NewLevel: 2,
	})
	if p.msgType != msgArticleLevelUp || p.uid != 5 {
		t.Fatalf("type=%s uid=%d", p.msgType, p.uid)
	}
}

type multiPusher struct {
	p     *mockPusher
	types *[]string
}

func (m *multiPusher) PushToUser(_ context.Context, uid uint64, msgType string, _ uint64, data interface{}) error {
	*m.types = append(*m.types, msgType)
	raw, _ := json.Marshal(data)
	m.p.uid = uid
	m.p.msgType = msgType
	m.p.data = raw
	return nil
}
