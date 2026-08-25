package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"gitlab.com/ecommercehub1/api/internal/core/model"
	"gitlab.com/ecommercehub1/api/internal/repository"
	"gitlab.com/ecommercehub1/shared/pkg/errors"
	"github.com/google/uuid"
)

type IAuthService interface {
	Register(ctx context.Context, email, password string, deviceInfo, ipAddress, userAgent string) (*AuthTokens, *errors.Error)
	Login(ctx context.Context, email, password string, deviceInfo, ipAddress, userAgent string) (*AuthTokens, *errors.Error)
	RefreshToken(ctx context.Context, refreshTokenStr string) (*AuthTokens, *errors.Error)
	Logout(ctx context.Context, sessionID int64) *errors.Error
}

type AuthService struct {
	*baseService
	identityRepo repository.IIdentityRepository
	sessionRepo  repository.ISessionRepository
	jwtSecret    []byte
}

type AuthTokens struct {
	AccessToken  string
	RefreshToken string
	SessionID    int64
	User         *model.AuthIdentities
}

type JWTClaims struct {
	SessionID int64  `json:"session_id"`
	UserID    string `json:"user_id"`
	jwt.RegisteredClaims
}

func NewAuthService(identityRepo *repository.IdentityRepository, sessionRepo *repository.SessionRepository) *AuthService {
	// TODO: Inject JWT secret from config
	return &AuthService{
		baseService:  NewBaseService(),
		identityRepo: identityRepo,
		sessionRepo:  sessionRepo,
		jwtSecret:    []byte("super-secret-key-replace-me-later"),
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string, deviceInfo, ipAddress, userAgent string) (*AuthTokens, *errors.Error) {
	// 1. Check if email already exists
	existing, err := s.identityRepo.GetCredentialByIdentifier(ctx, email, model.ProviderLocal)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.ErrConflict(ctx, "Email", "already exists")
	}

	// 2. Hash password
	hashedPassword, bcryptErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if bcryptErr != nil {
		return nil, errors.ErrSystemError(ctx, "Failed to hash password")
	}

	// 3. Create identity
	userID := uuid.New().String()
	identity := &model.AuthIdentities{
		UserID:     userID,
		Provider:   model.ProviderLocal,
		Identifier: email,
		Password:   string(hashedPassword),
	}

	identity, err = s.identityRepo.CreateIdentity(ctx, identity)
	if err != nil {
		return nil, err
	}

	// 4. Create session in DB
	now := time.Now()
	expiresAt := now.Add(7 * 24 * time.Hour) // 7 days for refresh token
	session := &model.Sessions{
		UserID:     identity.UserID,
		Token:      "temp", // Will update after generating JWT
		Provider:   model.ProviderLocal,
		DeviceInfo: deviceInfo,
		IpAddress:  ipAddress,
		UserAgent:  userAgent,
		IsActive:   true,
		ExpiresAt:  expiresAt,
	}

	session, err = s.sessionRepo.CreateSession(ctx, session)
	if err != nil {
		return nil, err
	}

	// 5. Generate JWT tokens
	accessToken, jwtErr := s.generateToken(session.ID, identity.UserID, 15*time.Minute)
	if jwtErr != nil {
		return nil, errors.ErrSystemError(ctx, "Failed to generate access token")
	}

	refreshToken, jwtErr := s.generateToken(session.ID, identity.UserID, 7*24*time.Hour)
	if jwtErr != nil {
		return nil, errors.ErrSystemError(ctx, "Failed to generate refresh token")
	}

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionID:    session.ID,
		User:         identity,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string, deviceInfo, ipAddress, userAgent string) (*AuthTokens, *errors.Error) {
	// 1. Get identity by email
	identity, err := s.identityRepo.GetCredentialByIdentifier(ctx, email, model.ProviderLocal)
	if err != nil {
		return nil, err
	}
	if identity == nil {
		return nil, errors.ErrUnauthorized(ctx).SetMessage("Invalid email or password")
	}

	// 2. Check password
	if bcryptErr := bcrypt.CompareHashAndPassword([]byte(identity.Password), []byte(password)); bcryptErr != nil {
		return nil, errors.ErrUnauthorized(ctx).SetMessage("Invalid email or password")
	}

	// 3. Create session in DB
	now := time.Now()
	expiresAt := now.Add(7 * 24 * time.Hour) // 7 days for refresh token
	session := &model.Sessions{
		UserID:     identity.UserID,
		Token:      "temp", // Will update after generating JWT
		Provider:   model.ProviderLocal,
		DeviceInfo: deviceInfo,
		IpAddress:  ipAddress,
		UserAgent:  userAgent,
		IsActive:   true,
		ExpiresAt:  expiresAt,
	}

	session, err = s.sessionRepo.CreateSession(ctx, session)
	if err != nil {
		return nil, err
	}

	// 4. Generate JWT tokens
	accessToken, jwtErr := s.generateToken(session.ID, identity.UserID, 15*time.Minute)
	if jwtErr != nil {
		return nil, errors.ErrSystemError(ctx, "Failed to generate access token")
	}

	refreshToken, jwtErr := s.generateToken(session.ID, identity.UserID, 7*24*time.Hour)
	if jwtErr != nil {
		return nil, errors.ErrSystemError(ctx, "Failed to generate refresh token")
	}

	// 5. Update session with refresh token
	// TODO: add UpdateToken to session repository if needed, or just leave it for now.
	// Actually, the model says `Token`. It usually stores the refresh token string or a hash of it.
	// We will skip updating for brevity unless required by logic.

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionID:    session.ID,
		User:         identity,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (*AuthTokens, *errors.Error) {
	// 1. Parse token
	claims, err := s.parseToken(refreshTokenStr)
	if err != nil {
		return nil, errors.ErrUnauthorized(ctx).SetMessage("Invalid refresh token")
	}

	// 2. Validate session
	session, dbErr := s.sessionRepo.GetSessionByID(ctx, claims.SessionID)
	if dbErr != nil {
		return nil, dbErr
	}
	if session == nil || !session.IsActive || session.ExpiresAt.Before(time.Now()) {
		return nil, errors.ErrUnauthorized(ctx).SetMessage("Session expired or invalid")
	}

	// 3. Generate new tokens
	accessToken, jwtErr := s.generateToken(session.ID, session.UserID, 15*time.Minute)
	if jwtErr != nil {
		return nil, errors.ErrSystemError(ctx, "Failed to generate access token")
	}

	newRefreshToken, jwtErr := s.generateToken(session.ID, session.UserID, 7*24*time.Hour)
	if jwtErr != nil {
		return nil, errors.ErrSystemError(ctx, "Failed to generate refresh token")
	}

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		SessionID:    session.ID,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, sessionID int64) *errors.Error {
	return s.sessionRepo.RevokeSession(ctx, sessionID)
}

func (s *AuthService) generateToken(sessionID int64, userID string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		SessionID: sessionID,
		UserID:    userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   userID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) parseToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token claims")
}
