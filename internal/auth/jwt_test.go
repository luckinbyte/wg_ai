package auth

import (
	"testing"
)

func TestJWTGenerateAndValidate(t *testing.T) {
	secret := "test-secret-key"

	token, err := GenerateToken(123, secret)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	uid, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	if uid != 123 {
		t.Errorf("expected uid 123, got %d", uid)
	}
}

func TestJWTInvalidToken(t *testing.T) {
	_, err := ValidateToken("invalid-token", "secret")
	if err == nil {
		t.Error("should fail for invalid token")
	}
}

func TestJWTWrongSecret(t *testing.T) {
	token, _ := GenerateToken(123, "secret1")
	_, err := ValidateToken(token, "secret2")
	if err == nil {
		t.Error("should fail for wrong secret")
	}
}
