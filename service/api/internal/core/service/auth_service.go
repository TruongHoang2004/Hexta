package service

import (
	"context"
	"time"

	"gitlab.com/ecommercehub1/api/config"
	"gitlab.com/ecommercehub1/api/internal/core/model"
	"gitlab.com/ecommercehub1/api/internal/core/types"
	"gitlab.com/ecommercehub1/api/internal/present/http/dto"
	"gitlab.com/ecommercehub1/api/internal/repository"
	"gitlab.com/ecommercehub1/api/pkg/client"
	"gitlab.com/ecommercehub1/api/utils"
	"gitlab.com/ecommercehub1/lib/pkg/errors"
	pb "gitlab.com/ecommercehub1/user/proto"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	*baseService
	userClient         *client.UserClient
	identifyRepository *repository.IdentityRepository
	sessionRepository  repository.ISessionRepository
}

func NewAuthService(
	userClient *client.UserClient,
	identifyRepository *repository.IdentityRepository,
	sessionRepository repository.ISessionRepository,
) *AuthService {
	return &AuthService{
		userClient:         userClient,
		identifyRepository: identifyRepository,
		sessionRepository:  sessionRepository,
	}
}

func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, *errors.Error) {

	gender := pb.Gender_GENDER_UNSPECIFIED
	switch req.Gender {
	case dto.GenderMale:
		gender = pb.Gender_GENDER_MALE
	case dto.GenderFemale:
		gender = pb.Gender_GENDER_FEMALE
	}
	res, rpcErr := s.userClient.CreateUser(ctx, &pb.CreateUserRequest{
		FullName:    req.FullName,
		Gender:      gender,
		DateOfBirth: req.DateOfBirth.Local().Format("2006-01-02"),
		UserName:    req.UserName,
		Email:       &req.Email,
		Phone:       &req.Phone,
	})
	if rpcErr != nil {
		return nil, s.grpToIError(ctx, rpcErr)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.ErrSystemError(ctx, err.Error())
	}

	_, iErr := s.identifyRepository.CreateIdentity(ctx, &model.AuthIdentities{
		UserID:     res.User.Id,
		Provider:   model.ProviderLocal,
		Identifier: req.UserName,
		Password:   string(hashedPassword),
	})

	if iErr != nil {
		return nil, iErr
	}

	return &dto.RegisterResponse{
		ID:          res.User.Id,
		UserName:    res.User.UserName,
		FullName:    res.User.FullName,
		Email:       res.User.GetEmail(),
		Gender:      req.Gender,
		Phone:       res.User.GetPhone(),
		DateOfBirth: req.DateOfBirth,
		CreatedAt:   res.User.CreatedAt.AsTime(),
		UpdatedAt:   res.User.UpdatedAt.AsTime(),
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest, ip, userAgent string) (*dto.LoginResponse, *errors.Error) {

	identify, iErr := s.identifyRepository.GetCredentialByIdentifier(ctx, req.UserName, model.ProviderLocal)
	if iErr != nil {
		return nil, iErr
	}

	if identify == nil {
		return nil, errors.ErrUnauthorized(ctx)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(identify.Password), []byte(req.Password)); err != nil {
		return nil, errors.ErrUnauthorized(ctx)
	}

	refreshToken, err := utils.GenerateJWT(types.RefreshTokenPayload{
		UserID: identify.UserID,
	}, config.AppConfig.JWT.RefreshTokenSecret, config.AppConfig.JWT.RefreshTokenExpire)
	if err != nil {
		return nil, errors.ErrSystemError(ctx, err.Error())
	}

	session, iErr := s.sessionRepository.CreateSession(ctx, &model.Sessions{
		UserID:    identify.UserID,
		Token:     refreshToken,
		Provider:  model.ProviderLocal,
		IpAddress: ip,
		UserAgent: userAgent,
		IsActive:  true,
		ExpiresAt: time.Now().Add(time.Duration(config.AppConfig.JWT.RefreshTokenExpire) * time.Minute),
	})
	if iErr != nil {
		return nil, iErr
	}

	accessToken, err := utils.GenerateJWT(types.AccessTokenPayload{
		UserID:    identify.UserID,
		SessionID: session.ID,
	}, config.AppConfig.JWT.AccessTokenSecret, config.AppConfig.JWT.AccessTokenExpire)
	if err != nil {
		return nil, errors.ErrSystemError(ctx, err.Error())
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, token string) (*types.AccessTokenPayload, *errors.Error) {
	claim, err := utils.DecodeJWT[types.AccessTokenPayload](token, config.AppConfig.JWT.AccessTokenSecret)
	if err != nil {
		return nil, errors.ErrUnauthorized(ctx).SetDetail(err.Error())
	}

	payload := claim.Payload

	session, iErr := s.sessionRepository.GetSessionByID(ctx, payload.SessionID)
	if iErr != nil {
		return nil, iErr.SetDetail("Session not found")
	}

	if session == nil || !session.IsActive || session.ExpiresAt.Before(time.Now()) {
		return nil, errors.ErrUnauthorized(ctx).SetDetail("Session is not valid")
	}
	return &payload, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, *errors.Error) {
	claim, err := utils.DecodeJWT[types.RefreshTokenPayload](req.RefreshToken, config.AppConfig.JWT.RefreshTokenSecret)
	if err != nil {
		return nil, errors.ErrUnauthorized(ctx).SetDetail(err.Error())
	}

	payload := claim.Payload

	session, iErr := s.sessionRepository.GetSessionByToken(ctx, req.RefreshToken)
	if iErr != nil {
		return nil, iErr
	}

	if session == nil || !session.IsActive || session.ExpiresAt.Before(time.Now()) {
		return nil, errors.ErrUnauthorized(ctx)
	}

	iErr = s.sessionRepository.RevokeSession(ctx, session.ID)
	if iErr != nil {
		return nil, iErr
	}

	refreshToken, err := utils.GenerateJWT(types.RefreshTokenPayload{
		UserID: payload.UserID,
	}, config.AppConfig.JWT.RefreshTokenSecret, config.AppConfig.JWT.RefreshTokenExpire)
	if err != nil {
		return nil, errors.ErrSystemError(ctx, err.Error())
	}

	newSession, iErr := s.sessionRepository.CreateSession(ctx, &model.Sessions{
		UserID:    payload.UserID,
		Token:     refreshToken,
		Provider:  model.ProviderLocal,
		IpAddress: session.IpAddress,
		UserAgent: session.UserAgent,
		IsActive:  true,
		ExpiresAt: time.Now().Add(time.Duration(config.AppConfig.JWT.RefreshTokenExpire) * time.Minute),
	})
	if iErr != nil {
		return nil, iErr
	}

	accessToken, err := utils.GenerateJWT(types.AccessTokenPayload{
		UserID:    payload.UserID,
		SessionID: newSession.ID,
	}, config.AppConfig.JWT.AccessTokenSecret, config.AppConfig.JWT.AccessTokenExpire)
	if err != nil {
		return nil, errors.ErrSystemError(ctx, err.Error())
	}

	return &dto.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
