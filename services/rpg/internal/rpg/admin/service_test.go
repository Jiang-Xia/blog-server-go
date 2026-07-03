package admin

import "testing"

func TestIntField(t *testing.T) {
	m := map[string]interface{}{"sort": float64(5), "active": true}
	if got := intField(m, "sort", 0); got != 5 {
		t.Fatalf("sort=%d", got)
	}
	if got := boolToIntField(m, "active", 0); got != 1 {
		t.Fatalf("active=%d", got)
	}
}

func TestMergeEffectJSON(t *testing.T) {
	existing := `{"maxProgress":1}`
	merged, err := mergeEffectJSON(&existing, map[string]interface{}{
		"effectJson": map[string]interface{}{"trackEvent": "sign_in", "expReward": 10},
	})
	if err != nil {
		t.Fatal(err)
	}
	if merged["trackEvent"] != "sign_in" {
		t.Fatalf("trackEvent=%v", merged["trackEvent"])
	}
	if merged["maxProgress"].(float64) != 1 {
		t.Fatalf("maxProgress=%v", merged["maxProgress"])
	}
}
