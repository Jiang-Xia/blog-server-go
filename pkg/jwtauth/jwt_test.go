package jwtauth_test

import (
	"testing"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/jwtauth"
)

func testConfig(secret string) *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret:     secret,
			LegacyTTL:  8 * time.Hour,
			AccessTTL:  30 * time.Minute,
			RefreshTTL: 7 * 24 * time.Hour,
		},
	}
}

func TestSignTripleAndVerify(t *testing.T) {
	svc := jwtauth.NewService(testConfig("unit-test-secret"))
	roles := []jwtauth.RolePayload{{ID: 1, RoleName: "super"}}
	triple, err := svc.SignTriple(1, "super", "18888888888", roles)
	if err != nil {
		t.Fatal(err)
	}
	if triple.AccessToken == "" || triple.RefreshToken == "" || triple.Token == "" {
		t.Fatal("triple tokens should not be empty")
	}
	claims, err := svc.Verify(triple.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if claims.ID != 1 || claims.Username != "18888888888" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	if len(claims.Role) != 1 || claims.Role[0].RoleName != "super" {
		t.Fatalf("role payload missing: %+v", claims.Role)
	}
}

func TestVerifyRejectsBadToken(t *testing.T) {
	svc := jwtauth.NewService(testConfig("unit-test-secret"))
	if _, err := svc.Verify("bad.token.value"); err == nil {
		t.Fatal("invalid token should fail verify")
	}
}

func TestRemainingTTL(t *testing.T) {
	svc := jwtauth.NewService(testConfig("unit-test-secret"))
	triple, err := svc.SignTriple(2, "u2", "u2", nil)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := svc.Verify(triple.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	ttl := svc.RemainingTTL(claims)
	if ttl < 1 {
		t.Fatalf("remaining ttl should be positive, got %d", ttl)
	}
}
