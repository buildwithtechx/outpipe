package auth

import "testing"

func TestTokenHashAndComparison(t *testing.T) {
	token, err := NewToken("test", 32)

	if err != nil {
		t.Fatal(err)
	}

	hash := HashToken(token)

	if !EqualHash(hash, token) {
		t.Fatal("expected token hash to match")
	}

	if EqualHash(hash, token+"x") {
		t.Fatal("expected different token not to match")
	}
}
