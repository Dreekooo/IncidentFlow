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
	t.Setenv("JWT_SECRET", "test-secret")

	token, err := auth.GenerateAuthToken(42)
	if err != nil {
		t.Fatalf("GenerateAuthToken() error = %v", err)
	}

	claims, err := auth.ParseAuthToken(token)
	if err != nil {
		t.Fatalf("ParseAuthToken() error = %v", err)
	}

	if claims.UserID != 42 {
		t.Fatalf("expected user id 42, got %d", claims.UserID)
	}

	if claims.Subject != "42" {
		t.Fatalf("expected subject 42, got %s", claims.Subject)
	}
}

func TestGenerateAuthTokenMissingSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")

	_, err := auth.GenerateAuthToken(1)
	if err == nil {
		t.Fatal("expected error when JWT_SECRET is missing")
	}
}

func TestParseAuthTokenMissingSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")

	_, err := auth.ParseAuthToken("some-token")
	if err == nil {
		t.Fatal("expected error when JWT_SECRET is missing")
	}
}

func TestParseAuthTokenInvalidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	_, err := auth.ParseAuthToken("invalid-token")
	if err == nil {
		t.Fatal("expected error when parsing invalid token")
	}
}

func TestParseAuthTokenWrongSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	token, err := auth.GenerateAuthToken(1)
	if err != nil {
		t.Fatalf("GenerateAuthToken() error = %v", err)
	}

	t.Setenv("JWT_SECRET", "wrong-secret")

	_, err = auth.ParseAuthToken(token)
	if err == nil {
		t.Fatal("expected error when parsing token with wrong secret")
	}
}

func TestGenerateAuthTokenHasExpiry(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	token, err := auth.GenerateAuthToken(1)
	if err != nil {
		t.Fatalf("GenerateAuthToken() error = %v", err)
	}

	claims, err := auth.ParseAuthToken(token)
	if err != nil {
		t.Fatalf("ParseAuthToken() error = %v", err)
	}

	if claims.ExpiresAt == nil {
		t.Fatal("expected token to have expiry claim")
	}
}

func TestGenerateAuthTokenIssuer(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	
	token, err := auth.GenerateAuthToken(1)
	if err != nil {
		t.Fatalf("GenerateAuthToken() error = %v", err)
	}

	claims, err := auth.ParseAuthToken(token)
	if err != nil {
		t.Fatalf("ParseAuthToken() error = %v", err)
	}

	if claims.Issuer != "incident-flow" {
		t.Fatalf("expected issuer 'incident-flow', got '%s'", claims.Issuer)
	}
}