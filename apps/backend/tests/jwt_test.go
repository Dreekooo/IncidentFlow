package tests

import (
	"testing"

	"incident-flow/backend/internal/auth"
)

/**
 * Tests for JWT token generation and parsing.
 * These tests cover:
 * - Successful token generation and parsing
 * - Handling of missing JWT secret
 * - Handling of invalid tokens
 * - Handling of tokens signed with the wrong secret
 * - Ensuring tokens have expiry and correct issuer
 */

func TestGenerateAndParseAuthToken(t *testing.T) {
	secret := []byte("test-secret")

	token, err := auth.GenerateAuthTokenWithSecret(2, secret)
	if err != nil {
		t.Fatalf("GenerateAuthToken() error = %v", err)
	}

	claims, err := auth.ParseAuthTokenWithSecret(token, secret)
	if err != nil {
		t.Fatalf("ParseAuthToken() error = %v", err)
	}

	if claims.UserID != 2 {
		t.Fatalf("expected user id 2, got %d", claims.UserID)
	}

	if claims.Subject != "2" {
		t.Fatalf("expected subject 2, got %s", claims.Subject)
	}
}

func TestGenerateAuthTokenMissingSecret(t *testing.T) {
	_, err := auth.GenerateAuthTokenWithSecret(1, nil)
	if err == nil {
		t.Fatal("expected error when JWT_SECRET is missing")
	}
}

func TestParseAuthTokenMissingSecret(t *testing.T) {
	_, err := auth.ParseAuthTokenWithSecret("some-token", nil)
	if err == nil {
		t.Fatal("expected error when JWT_SECRET is missing")
	}
}

func TestParseAuthTokenInvalidToken(t *testing.T) {
	_, err := auth.ParseAuthTokenWithSecret("invalid-token", []byte("test-secret"))
	if err == nil {
		t.Fatal("expected error when parsing invalid token")
	}
}

func TestParseAuthTokenWrongSecret(t *testing.T) {
	secret := []byte("test-secret")
	token, err := auth.GenerateAuthTokenWithSecret(1, secret)
	if err != nil {
		t.Fatalf("GenerateAuthToken() error = %v", err)
	}

	_, err = auth.ParseAuthTokenWithSecret(token, []byte("wrong-secret"))
	if err == nil {
		t.Fatal("expected error when parsing token with wrong secret")
	}
}

func TestGenerateAuthTokenHasExpiry(t *testing.T) {
	token, err := auth.GenerateAuthTokenWithSecret(1, []byte("test-secret"))
	if err != nil {
		t.Fatalf("GenerateAuthToken() error = %v", err)
	}

	claims, err := auth.ParseAuthTokenWithSecret(token, []byte("test-secret"))
	if err != nil {
		t.Fatalf("ParseAuthToken() error = %v", err)
	}

	if claims.ExpiresAt == nil {
		t.Fatal("expected token to have expiry claim")
	}
}

func TestGenerateAuthTokenIssuer(t *testing.T) {
	token, err := auth.GenerateAuthTokenWithSecret(1, []byte("test-secret"))
	if err != nil {
		t.Fatalf("GenerateAuthToken() error = %v", err)
	}

	claims, err := auth.ParseAuthTokenWithSecret(token, []byte("test-secret"))
	if err != nil {
		t.Fatalf("ParseAuthToken() error = %v", err)
	}

	if claims.Issuer != "incident-flow" {
		t.Fatalf("expected issuer 'incident-flow', got '%s'", claims.Issuer)
	}
}
