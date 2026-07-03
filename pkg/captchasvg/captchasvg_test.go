package captchasvg

import (
	"strings"
	"testing"
)

func TestCreate(t *testing.T) {
	result, err := Create()
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if len(result.Text) != 4 {
		t.Fatalf("text len = %d, want 4", len(result.Text))
	}
	if !strings.Contains(result.Data, "<path fill=") {
		t.Fatal("svg missing character paths")
	}

	circles := strings.Count(result.Data, "<circle")
	if circles != noiseDots {
		t.Fatalf("circle count = %d, want %d", circles, noiseDots)
	}

	curves := strings.Count(result.Data, `fill="none"`)
	if curves != int(curveCount) {
		t.Fatalf("curve count = %d, want %d", curves, curveCount)
	}
}
