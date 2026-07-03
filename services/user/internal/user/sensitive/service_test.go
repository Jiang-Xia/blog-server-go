package sensitive

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"
)

type mockModerationSync struct {
	calls []struct{ sourceType, sourceID, status string }
	err   error
}

func (m *mockModerationSync) UpdateContentModerationStatus(_ context.Context, sourceType, sourceID, status string) error {
	m.calls = append(m.calls, struct{ sourceType, sourceID, status string }{sourceType, sourceID, status})
	return m.err
}

func TestSyncSourceStatusCallsBlog(t *testing.T) {
	mock := &mockModerationSync{}
	s := &Service{log: zap.NewNop(), blogSync: mock}
	if err := s.syncSourceStatus(context.Background(), "comment", "c1", "approved"); err != nil {
		t.Fatal(err)
	}
	if len(mock.calls) != 1 || mock.calls[0].sourceType != "comment" || mock.calls[0].sourceID != "c1" {
		t.Fatalf("calls=%+v", mock.calls)
	}
}

func TestSyncSourceStatusSkipsEmptyID(t *testing.T) {
	mock := &mockModerationSync{}
	s := &Service{log: zap.NewNop(), blogSync: mock}
	if err := s.syncSourceStatus(context.Background(), "comment", "", "approved"); err != nil {
		t.Fatal(err)
	}
	if len(mock.calls) != 0 {
		t.Fatal("expected no calls")
	}
}

func TestSyncSourceStatusPropagatesError(t *testing.T) {
	mock := &mockModerationSync{err: errors.New("grpc down")}
	s := &Service{log: zap.NewNop(), blogSync: mock}
	if err := s.syncSourceStatus(context.Background(), "reply", "r1", "rejected"); err == nil {
		t.Fatal("expected error")
	}
}
