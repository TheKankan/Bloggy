package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ── HashPassword ──────────────────────────────────────────────────────────────

func TestHashPassword_ReturnsNonEmptyHash(t *testing.T) {
	hash, err := HashPassword("password")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if hash == "" {
		t.Fatal("expected a non-empty hash")
	}
}

func TestHashPassword_DifferentHashesForSamePassword(t *testing.T) {
	hash1, err := HashPassword("supersecret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	hash2, err := HashPassword("supersecret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// argon2id uses a random salt, so hashes must differ
	if hash1 == hash2 {
		t.Fatal("expected different hashes for the same password (random salt)")
	}
}

// ── CheckPasswordHash ─────────────────────────────────────────────────────────

func TestCheckPasswordHash_CorrectPassword(t *testing.T) {
	hash, err := HashPassword("supersecret")
	if err != nil {
		t.Fatalf("unexpected error hashing: %v", err)
	}
	match, err := CheckPasswordHash("supersecret", hash)
	if err != nil {
		t.Fatalf("unexpected error checking: %v", err)
	}
	if !match {
		t.Fatal("expected password to match hash")
	}
}

func TestCheckPasswordHash_WrongPassword(t *testing.T) {
	hash, err := HashPassword("supersecret")
	if err != nil {
		t.Fatalf("unexpected error hashing: %v", err)
	}
	match, err := CheckPasswordHash("wrongpassword", hash)
	if err != nil {
		t.Fatalf("unexpected error checking: %v", err)
	}
	if match {
		t.Fatal("expected password NOT to match hash")
	}
}

func TestCheckPasswordHash_InvalidHash(t *testing.T) {
	_, err := CheckPasswordHash("supersecret", "this-is-not-a-valid-hash")
	if err == nil {
		t.Fatal("expected an error with an invalid hash, got nil")
	}
}

// ── MakeJWT / ValidateJWT ─────────────────────────────────────────────────────

func TestMakeJWT_ReturnsValidToken(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	token, err := MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if token == "" {
		t.Fatal("expected a non-empty token")
	}
}

func TestValidateJWT_ValidToken(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	token, err := MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error making JWT: %v", err)
	}

	parsedID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("expected no error validating, got: %v", err)
	}
	if parsedID != userID {
		t.Fatalf("expected userID %v, got %v", userID, parsedID)
	}
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	// Create a token that expires immediately (negative duration = already expired)
	token, err := MakeJWT(userID, secret, -time.Second)
	if err != nil {
		t.Fatalf("unexpected error making JWT: %v", err)
	}

	_, err = ValidateJWT(token, secret)
	if err == nil {
		t.Fatal("expected an error for expired token, got nil")
	}
}

func TestValidateJWT_WrongSecret(t *testing.T) {
	userID := uuid.New()

	token, err := MakeJWT(userID, "correct-secret", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error making JWT: %v", err)
	}

	_, err = ValidateJWT(token, "wrong-secret")
	if err == nil {
		t.Fatal("expected an error with wrong secret, got nil")
	}
}

func TestValidateJWT_MalformedToken(t *testing.T) {
	_, err := ValidateJWT("this.is.not.a.jwt", "some-secret")
	if err == nil {
		t.Fatal("expected an error for malformed token, got nil")
	}
}

// ── GetBearerToken ────────────────────────────────────────────────────────────

func TestGetBearerToken_ValidHeader(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer mytoken123")

	token, err := GetBearerToken(headers)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if token != "mytoken123" {
		t.Fatalf("expected 'mytoken123', got '%s'", token)
	}
}

func TestGetBearerToken_MissingHeader(t *testing.T) {
	headers := http.Header{}

	_, err := GetBearerToken(headers)
	if err == nil {
		t.Fatal("expected an error for missing header, got nil")
	}
}

func TestGetBearerToken_MalformedHeader_NoBearer(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "mytoken123") // missing "Bearer" prefix

	_, err := GetBearerToken(headers)
	if err == nil {
		t.Fatal("expected an error for malformed header, got nil")
	}
}

func TestGetBearerToken_MalformedHeader_WrongScheme(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Basic mytoken123")

	_, err := GetBearerToken(headers)
	if err == nil {
		t.Fatal("expected an error for wrong auth scheme, got nil")
	}
}

// ── MakeRefreshToken ──────────────────────────────────────────────────────────

func TestMakeRefreshToken_ReturnsNonEmptyToken(t *testing.T) {
	token := MakeRefreshToken()
	if token == "" {
		t.Fatal("expected a non-empty refresh token")
	}
}

func TestMakeRefreshToken_TokensAreUnique(t *testing.T) {
	token1 := MakeRefreshToken()
	token2 := MakeRefreshToken()
	if token1 == token2 {
		t.Fatal("expected unique refresh tokens, got identical ones")
	}
}

func TestMakeRefreshToken_IsHexEncoded(t *testing.T) {
	token := MakeRefreshToken()
	// 32 bytes hex-encoded = 64 characters
	if len(token) != 64 {
		t.Fatalf("expected token length of 64, got %d", len(token))
	}
	for _, c := range token {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Fatalf("expected hex characters only, got '%c'", c)
		}
	}
}
