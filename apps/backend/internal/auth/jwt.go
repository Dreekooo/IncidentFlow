package auth

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

const (
	tokenExpiry = 15 * time.Minute
	issuer      = "incident-flow"
)

func getJWTSecret() ([]byte, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is not set")
	}

	return []byte(secret), nil
}

func GenerateAuthToken(userID int) (string, error) {
	secret, err := getJWTSecret()
	if err != nil {
		return "", err
	}

	return GenerateAuthTokenWithSecret(userID, secret)
}

func GenerateAuthTokenWithSecret(userID int, secret []byte) (string, error) {
	if len(secret) == 0 {
		return "", fmt.Errorf("JWT secret is empty")
	}

	now := time.Now()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(userID),
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenExpiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ParseAuthToken(tokenString string) (*Claims, error) {
	secret, err := getJWTSecret()
	if err != nil {
		return nil, err
	}

	return ParseAuthTokenWithSecret(tokenString, secret)
}

func ParseAuthTokenWithSecret(tokenString string, secret []byte) (*Claims, error) {
	if len(secret) == 0 {
		return nil, fmt.Errorf("JWT secret is empty")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// GenerateAuthTokenWithSecretAndExpiry generates a token with custom expiry time (for testing)
func GenerateAuthTokenWithSecretAndExpiry(userID int, secret []byte, expiresAt time.Time) (string, error) {
	if len(secret) == 0 {
		return "", fmt.Errorf("JWT secret is empty")
	}

	now := time.Now()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(userID),
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}
