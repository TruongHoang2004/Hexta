package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// JWTClaims is a generic struct that wraps a custom payload and standard JWT registered claims.
type JWTClaims[T any] struct {
	Payload T `json:"payload"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a new JWT token for a given generic payload, secret, and expiry duration (in seconds).
func GenerateJWT[T any](payload T, secret string, expiry int) (string, error) {
	claims := JWTClaims[T]{
		Payload: payload,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiry) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// DecodeJWT parses and validates a JWT token string using the provided secret and returns the generic claims.
func DecodeJWT[T any](tokenString string, secret string) (*JWTClaims[T], error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims[T]{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("can't parse token")
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token has expired")
		}
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims[T]); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
