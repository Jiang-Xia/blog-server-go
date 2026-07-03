package service

import "testing"

func TestUpdateContentModerationStatusUnknownType(t *testing.T) {
	svc := NewModerationService(nil, nil, nil)
	updated, err := svc.UpdateContentModerationStatus(t.Context(), "article", "1", "approved")
	if err != nil {
		t.Fatal(err)
	}
	if updated {
		t.Fatal("expected false for unknown source type")
	}
}
