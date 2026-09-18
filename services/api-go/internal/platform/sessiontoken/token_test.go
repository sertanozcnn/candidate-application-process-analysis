package sessiontoken

import "testing"

func TestGenerateReturnsDifferentTokens(t *testing.T) {
	first, err := Generate()
	if err != nil {
		t.Fatalf("first Generate returned error: %v", err)
	}

	second, err := Generate()
	if err != nil {
		t.Fatalf("second Generate returned error: %v", err)
	}

	if first == second {
		t.Fatal("Generate returned identical tokens")
	}
	if first == "" || second == "" {
		t.Fatal("Generate returned empty token")
	}
}

func TestHashIsStable(t *testing.T) {
	first := Hash("session-token")
	second := Hash("session-token")

	if first == "" {
		t.Fatal("Hash returned empty value")
	}
	if first != second {
		t.Fatal("Hash is not stable")
	}
	if first == "session-token" {
		t.Fatal("Hash returned the raw token")
	}
}
