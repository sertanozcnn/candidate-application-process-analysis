package password

import (
	"errors"
	"strings"
	"testing"
)

func TestHashAndVerify(t *testing.T) {
	hash, err := Hash("local-admin-password")
	if err != nil {
		t.Fatalf("Hash returned error: %v", err)
	}

	if !strings.HasPrefix(hash, "$argon2id$v=19$m=19456,t=2,p=1$") {
		t.Fatalf("hash format or params are unexpected: %s", hash)
	}

	matched, err := Verify("local-admin-password", hash)
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
	if !matched {
		t.Fatal("Verify did not match the original password")
	}
}

func TestVerifyRejectsWrongPassword(t *testing.T) {
	hash, err := Hash("local-admin-password")
	if err != nil {
		t.Fatalf("Hash returned error: %v", err)
	}

	matched, err := Verify("wrong-password", hash)
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
	if matched {
		t.Fatal("Verify matched a wrong password")
	}
}

func TestHashUsesDifferentSalt(t *testing.T) {
	first, err := Hash("local-admin-password")
	if err != nil {
		t.Fatalf("first Hash returned error: %v", err)
	}

	second, err := Hash("local-admin-password")
	if err != nil {
		t.Fatalf("second Hash returned error: %v", err)
	}

	if first == second {
		t.Fatal("same password produced identical hashes")
	}
}

func TestRejectsEmptyPassword(t *testing.T) {
	_, err := Hash("")
	if !errors.Is(err, ErrEmptyPassword) {
		t.Fatalf("Hash error = %v, want ErrEmptyPassword", err)
	}

	matched, err := Verify("", "$argon2id$v=19$m=19456,t=2,p=1$c2FsdA$aGFzaA")
	if matched {
		t.Fatal("Verify matched an empty password")
	}
	if !errors.Is(err, ErrEmptyPassword) {
		t.Fatalf("Verify error = %v, want ErrEmptyPassword", err)
	}
}

func TestVerifyRejectsInvalidHash(t *testing.T) {
	matched, err := Verify("local-admin-password", "not-a-hash")
	if matched {
		t.Fatal("Verify matched an invalid hash")
	}
	if !errors.Is(err, ErrInvalidHash) {
		t.Fatalf("Verify error = %v, want ErrInvalidHash", err)
	}
}
