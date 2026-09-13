package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"gitlab.com/ecommercehub1/api/config"
	"gitlab.com/ecommercehub1/api/internal/core/model"
	"gitlab.com/ecommercehub1/api/internal/core/types"
	"gitlab.com/ecommercehub1/api/internal/infrastructure/cache"
	"gitlab.com/ecommercehub1/api/internal/repository"
	"gitlab.com/ecommercehub1/shared/pkg/errors"
)

const (
	OAuthStateTTL         = 5 * time.Minute
	defaultAccessSecret   = "access-token-secret-fallback-for-dev"
	defaultRefreshSecret  = "refresh-token-secret-fallback-for-dev"
	defaultAccessExpire   = 15 * time.Minute
	defaultRefreshExpire  = 7 * 24 * time.Hour
	TokenTypeAccess       = "access"
	TokenTypeRefresh      = "refresh"
)

type IAuthService interface {
	Register(ctx context.Context, email, password string, deviceInfo, ipAddress, userAgent string) (*AuthTokens, *errors.Error)
	Login(ctx context.Context, email, password string, deviceInfo, ipAddress, userAgent string) (*AuthTokens, *errors.Error)
	GoogleLogin(ctx context.Context) (string, *errors.Error)
	GoogleCallback(ctx context.Context, code string, state string, deviceInfo, ipAddress, userAgent string) (*AuthTokens, *errors.Error)
	RefreshToken(ctx context.Context, refreshTokenStr string) (*AuthTokens, *errors.Error)
	Logout(ctx context.Context, userID string, sessionID int64) *errors.Error
	ValidateAccessToken(ctx context.Context, tokenStr string) (*types.AuthInfo, *errors.Error)
	ValidateAndConsumeOAuthState(ctx context.Context, state string) *errors.Error
	GetAccessTokenExpire() time.Duration
	GetRefreshTokenExpire() time.Duration
}

type AuthService struct {
	*baseService
	identityRepo       repository.IIdentityRepository
	sessionRepo        repository.ISessionRepository
	redisClient        *cache.RedisClient
	stateCache         sync.Map
	accessTokenSecret  []byte
	accessTokenExpire  time.Duration
	refreshTokenSecret []byte
	refreshTokenExpire time.Duration
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
	TokenType string `json:"token_type,omitempty"`
	jwt.RegisteredClaims
}

func NewAuthService(
	identityRepo *repository.IdentityRepository,
	sessionRepo *repository.SessionRepository,
	redisClient *cache.RedisClient,
	cfg *config.Config,
) *AuthService {
	return NewAuthServiceWithRepos(identityRepo, sessionRepo, redisClient, cfg)
}

func NewAuthServiceWithRepos(
	identityRepo repository.IIdentityRepository,
	sessionRepo repository.ISessionRepository,
	redisClient *cache.RedisClient,
	cfg *config.Config,
) *AuthService {
	if cfg == nil {
		cfg = config.AppConfig
	}

	accessSecret := defaultAccessSecret
	if cfg != nil && cfg.JWT.AccessTokenSecret != "" {
		accessSecret = cfg.JWT.AccessTokenSecret
	}

	accessExpire := defaultAccessExpire
	if cfg != nil && cfg.JWT.AccessTokenExpire > 0 {
		accessExpire = time.Duration(cfg.JWT.AccessTokenExpire) * time.Second
	}

	refreshSecret := defaultRefreshSecret
	if cfg != nil && cfg.JWT.RefreshTokenSecret != "" {
		refreshSecret = cfg.JWT.RefreshTokenSecret
	}

	refreshExpire := defaultRefreshExpire
	if cfg != nil && cfg.JWT.RefreshTokenExpire > 0 {
		refreshExpire = time.Duration(cfg.JWT.RefreshTokenExpire) * time.Second
	}

	return &AuthService{
		baseService:        NewBaseService(),
		identityRepo:       identityRepo,
		sessionRepo:        sessionRepo,
		redisClient:        redisClient,
		accessTokenSecret:  []byte(accessSecret),
		accessTokenExpire:  accessExpire,
		refreshTokenSecret: []byte(refreshSecret),
		refreshTokenExpire: refreshExpire,
	}
}

func (s *AuthService) GetAccessTokenExpire() time.Duration {
	return s.accessTokenExpire
}

func (s *AuthService) GetRefreshTokenExpire() time.Duration {
	return s.refreshTokenExpire
}

func (s *AuthService) Register(ctx context.Context, email, password string, deviceInfo, ipAddress, userAgent string) (*AuthTokens, *errors.Error) {
	// 1. Check if email already exists
	existing, err := s.identityRepo.GetFirstByIdentifier(ctx, email)
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
	hashedPasswordStr := string(hashedPassword)
	identity := &model.AuthIdentities{
		UserID:     userID,
		Provider:   model.ProviderLocal,
		Identifier: email,
		Password:   &hashedPasswordStr,
	}

	identity, err = s.identityRepo.CreateIdentity(ctx, identity)
	if err != nil {
		return nil, err
	}

	// 4. Create session in DB
	now := time.Now()
	expiresAt := now.Add(s.refreshTokenExpire)
	session := &model.Sessions{
		UserID:     identity.UserID,
		Token:      "temp",
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
	accessToken, jwtErr := s.generateAccessToken(session.ID, identity.UserID)
	if jwtErr != nil {
		return nil, errors.ErrSystemError(ctx, "Failed to generate access token")
	}

	refreshToken, jwtErr := s.generateRefreshToken(session.ID, identity.UserID)
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
	if identity == nil || identity.Password == nil {
		return nil, errors.ErrUnauthorized(ctx).SetMessage("Invalid email or password")
	}

	// 2. Check password
	if bcryptErr := bcrypt.CompareHashAndPassword([]byte(*identity.Password), []byte(password)); bcryptErr != nil {
		return nil, errors.ErrUnauthorized(ctx).SetMessage("Invalid email or password")
	}

	// 3. Create session in DB
	now := time.Now()
	expiresAt := now.Add(s.refreshTokenExpire)
	session := &model.Sessions{
		UserID:     identity.UserID,
		Token:      "temp",
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
	accessToken, jwtErr := s.generateAccessToken(session.ID, identity.UserID)
	if jwtErr != nil {
		return nil, errors.ErrSystemError(ctx, "Failed to generate access token")
	}

	refreshToken, jwtErr := s.generateRefreshToken(session.ID, identity.UserID)
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

func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (*AuthTokens, *errors.Error) {
	// 1. Parse refresh token
	claims, err := s.parseRefreshToken(refreshTokenStr)
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
	accessToken, jwtErr := s.generateAccessToken(session.ID, session.UserID)
	if jwtErr != nil {
		return nil, errors.ErrSystemError(ctx, "Failed to generate access token")
	}

	newRefreshToken, jwtErr := s.generateRefreshToken(session.ID, session.UserID)
	if jwtErr != nil {
		return nil, errors.ErrSystemError(ctx, "Failed to generate refresh token")
	}

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		SessionID:    session.ID,
	}, nil
}

func (s *AuthService) getGoogleOAuthConfig() *oauth2.Config {
	cfg := config.AppConfig
	clientID := ""
	clientSecret := ""
	redirectURL := ""
	if cfg != nil {
		clientID = cfg.OAuth.Google.ClientID
		clientSecret = cfg.OAuth.Google.ClientSecret
		redirectURL = cfg.OAuth.Google.RedirectURL
	}

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

func (s *AuthService) GoogleLogin(ctx context.Context) (string, *errors.Error) {
	conf := s.getGoogleOAuthConfig()
	state, err := s.generateAndSaveOAuthState(ctx)
	if err != nil {
		return "", errors.ErrSystemError(ctx, "Failed to generate secure OAuth state")
	}
	url := conf.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return url, nil
}

func (s *AuthService) generateAndSaveOAuthState(ctx context.Context) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := base64.RawURLEncoding.EncodeToString(b)

	if s.redisClient != nil {
		key := fmt.Sprintf("oauth_state:%s", state)
		if err := s.redisClient.Set(ctx, key, []byte("1"), OAuthStateTTL); err != nil {
			return "", err
		}
	} else {
		s.stateCache.Store(state, time.Now().Add(OAuthStateTTL))
	}
	return state, nil
}

func (s *AuthService) ValidateAndConsumeOAuthState(ctx context.Context, state string) *errors.Error {
	if state == "" {
		return errors.ErrUnauthorized(ctx).SetMessage("Missing OAuth state parameter")
	}

	if s.redisClient != nil {
		key := fmt.Sprintf("oauth_state:%s", state)
		val, err := s.redisClient.Get(ctx, key)
		if err != nil || len(val) == 0 {
			return errors.ErrUnauthorized(ctx).SetMessage("Invalid or expired OAuth state")
		}
		_ = s.redisClient.Delete(ctx, key)
		return nil
	}

	if exp, ok := s.stateCache.LoadAndDelete(state); ok {
		if expireTime, ok := exp.(time.Time); ok && expireTime.After(time.Now()) {
			return nil
		}
	}
	return errors.ErrUnauthorized(ctx).SetMessage("Invalid or expired OAuth state")
}

func (s *AuthService) GoogleCallback(ctx context.Context, code string, state string, deviceInfo, ipAddress, userAgent string) (*AuthTokens, *errors.Error) {
	// 1. Validate and consume OAuth state (CSRF protection)
	if err := s.ValidateAndConsumeOAuthState(ctx, state); err != nil {
		return nil, err
	}

	conf := s.getGoogleOAuthConfig()

	token, err := conf.Exchange(ctx, code)
	if err != nil {
		return nil, errors.ErrUnauthorized(ctx).SetMessage("Failed to exchange code")
	}

	client := conf.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, errors.ErrSystemError(ctx, "Failed to get user info")
	}
	defer resp.Body.Close()

	var userInfo struct {
		Id    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, errors.ErrSystemError(ctx, "Failed to parse user info")
	}

	// 2. Find identity by email under Google provider
	identity, dbErr := s.identityRepo.GetCredentialByIdentifier(ctx, userInfo.Email, model.ProviderGoogle)
	if dbErr != nil {
		return nil, dbErr
	}

	if identity == nil {
		// Check for existing local account to link
		localIdentity, err := s.identityRepo.GetCredentialByIdentifier(ctx, userInfo.Email, model.ProviderLocal)
		if err != nil {
			return nil, err
		}

		userID := uuid.New().String()
		if localIdentity != nil {
			userID = localIdentity.UserID
		}

		identity = &model.AuthIdentities{
			UserID:     userID,
			Provider:   model.ProviderGoogle,
			Identifier: userInfo.Email,
			Password:   "",
		}
		identity, dbErr = s.identityRepo.CreateIdentity(ctx, identity)
		if dbErr != nil {
			return nil, dbErr
		}
	}

	// 3. Create session
	now := time.Now()
	expiresAt := now.Add(s.refreshTokenExpire)
	session := &model.Sessions{
		UserID:     identity.UserID,
		Token:      "temp",
		Provider:   model.ProviderGoogle,
		DeviceInfo: deviceInfo,
		IpAddress:  ipAddress,
		UserAgent:  userAgent,
		IsActive:   true,
		ExpiresAt:  expiresAt,
	}

	session, dbErr = s.sessionRepo.CreateSession(ctx, session)
	if dbErr != nil {
		return nil, dbErr
	}

	// 4. Generate JWT tokens
	accessToken, jwtErr := s.generateAccessToken(session.ID, identity.UserID)
	if jwtErr != nil {
		return nil, errors.ErrSystemError(ctx, "Failed to generate access token")
	}

	refreshToken, jwtErr := s.generateRefreshToken(session.ID, identity.UserID)
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

func (s *AuthService) Logout(ctx context.Context, userID string, sessionID int64) *errors.Error {
	session, err := s.sessionRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return errors.ErrNotFound(ctx, "Session", "Session not found")
	}

	if session.UserID != userID {
		return errors.ErrForbidden(ctx).SetDetail("You are not authorized to revoke this session")
	}

	return s.sessionRepo.RevokeSession(ctx, sessionID)
}

func (s *AuthService) ValidateAccessToken(ctx context.Context, tokenStr string) (*types.AuthInfo, *errors.Error) {
	claims, err := s.parseAccessToken(tokenStr)
	if err != nil {
		return nil, errors.ErrUnauthorized(ctx).SetDetail(err.Error())
	}

	session, dbErr := s.sessionRepo.GetSessionByID(ctx, claims.SessionID)
	if dbErr != nil {
		return nil, dbErr
	}
	if session == nil || !session.IsActive || session.ExpiresAt.Before(time.Now()) {
		return nil, errors.ErrUnauthorized(ctx).SetDetail("Session is expired or revoked")
	}

	return &types.AuthInfo{
		UserID:    claims.UserID,
		SessionID: claims.SessionID,
	}, nil
}

func (s *AuthService) generateAccessToken(sessionID int64, userID string) (string, error) {
	return s.generateToken(sessionID, userID, TokenTypeAccess, s.accessTokenSecret, s.accessTokenExpire)
}

func (s *AuthService) generateRefreshToken(sessionID int64, userID string) (string, error) {
	return s.generateToken(sessionID, userID, TokenTypeRefresh, s.refreshTokenSecret, s.refreshTokenExpire)
}

func (s *AuthService) generateToken(sessionID int64, userID string, tokenType string, secret []byte, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		SessionID: sessionID,
		UserID:    userID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   userID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func (s *AuthService) parseAccessToken(tokenStr string) (*JWTClaims, error) {
	return s.parseTokenWithSecret(tokenStr, s.accessTokenSecret, TokenTypeAccess)
}

func (s *AuthService) parseRefreshToken(tokenStr string) (*JWTClaims, error) {
	return s.parseTokenWithSecret(tokenStr, s.refreshTokenSecret, TokenTypeRefresh)
}

func (s *AuthService) parseTokenWithSecret(tokenStr string, secret []byte, expectedType string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		if claims.TokenType != "" && claims.TokenType != expectedType {
			return nil, fmt.Errorf("invalid token type: expected %s, got %s", expectedType, claims.TokenType)
		}
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token claims")
}
