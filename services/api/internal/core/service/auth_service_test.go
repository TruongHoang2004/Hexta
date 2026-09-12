package service

import (
	"testing"
	"time"
)

func TestAuthService_JWTClaims_EmailAndSub(t *testing.T) {
	svc := &AuthService{
		baseService: NewBaseService(),
		jwtSecret:   []byte("test-jwt-secret-key-123"),
	}

	testUserID := "user-uuid-12345"
	testEmail := "user@example.com"
	testSessionID := int64(999)

	tokenStr, err := svc.generateToken(testSessionID, testUserID, testEmail, 15*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}

	claims, err := svc.parseToken(tokenStr)
	if err != nil {
		t.Fatalf("unexpected error parsing token: %v", err)
	}

	if claims.SessionID != testSessionID {
		t.Errorf("expected SessionID %d, got %d", testSessionID, claims.SessionID)
	}
	if claims.UserID != testUserID {
		t.Errorf("expected UserID %s, got %s", testUserID, claims.UserID)
	}
	if claims.Email != testEmail {
		t.Errorf("expected Email %s, got %s", testEmail, claims.Email)
	}
	if claims.Subject != testUserID {
		t.Errorf("expected Subject %s, got %s", testUserID, claims.Subject)
	}
}
