package crypto_test

import (
	"testing"

	"github.com/Jiang-Xia/blog-server-go/pkg/crypto"
)

// 与 Nest makeSalt + encryptPassword('test123', salt) 固定盐向量一致。
const (
	testLegacySalt = "dGVz"
	testLegacyHash = "Sj1mvTHJzvWRDUfEy90SHQ=="
	testPlain      = "test123"
)

func TestVerifyLegacyPBKDF2(t *testing.T) {
	if !crypto.Verify(testLegacyHash, testPlain, testLegacySalt) {
		t.Fatal("legacy PBKDF2 verify should pass")
	}
	if crypto.Verify(testLegacyHash, "wrong", testLegacySalt) {
		t.Fatal("wrong password should fail")
	}
}

func TestNeedsUpgrade(t *testing.T) {
	if !crypto.NeedsUpgrade(testLegacyHash) {
		t.Fatal("PBKDF2 hash should need upgrade")
	}
	hash, err := crypto.Hash(testPlain)
	if err != nil {
		t.Fatal(err)
	}
	if crypto.NeedsUpgrade(hash) {
		t.Fatal("bcrypt hash should not need upgrade")
	}
}

func TestUpgradeHash(t *testing.T) {
	upgraded, err := crypto.UpgradeHash(testPlain)
	if err != nil {
		t.Fatal(err)
	}
	if crypto.NeedsUpgrade(upgraded) {
		t.Fatal("upgraded hash should be bcrypt")
	}
	if !crypto.Verify(upgraded, testPlain, "") {
		t.Fatal("upgraded hash should verify plain password")
	}
}

func TestVerifyBcrypt(t *testing.T) {
	hash, err := crypto.Hash("newpass")
	if err != nil {
		t.Fatal(err)
	}
	if !crypto.Verify(hash, "newpass", "") {
		t.Fatal("bcrypt verify should pass without salt")
	}
}
