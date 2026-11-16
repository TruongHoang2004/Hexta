package service

import (
	"context"
	stderrors "errors"
	"net/http"
	"regexp"
	"strings"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
)

// telegramTokenPattern is the @BotFather token shape: numeric bot ID, colon, secret.
var telegramTokenPattern = regexp.MustCompile(`^[0-9]{1,20}:[A-Za-z0-9_-]{30,80}$`)

func (s *ChannelService) telegramConfigured(ctx context.Context) *apperrors.Error {
	if s.connector == nil {
		return apperrors.ErrSystemError(ctx, "telegram channel is not configured").SetHTTPStatus(http.StatusServiceUnavailable)
	}
	return nil
}

func telegramToken(ctx context.Context, token string) (string, *apperrors.Error) {
	token = strings.TrimSpace(token)
	if !telegramTokenPattern.MatchString(token) {
		return "", apperrors.ErrBadRequest(ctx).SetDetail("bot_token must be a Bot Token from @BotFather")
	}
	return token, nil
}

// verifyTelegramBot authorizes the owner, then has the connector call getMe. The owner check comes first so a
// non-owner can never trigger a Telegram call. The token is passed through and never stored or logged by the API.
func (s *ChannelService) verifyTelegramBot(ctx context.Context, tenant, actor, token string) (*types.TelegramBot, *apperrors.Error) {
	if err := s.authorized(ctx, tenant, actor, true, func(repository.ChannelStore) *apperrors.Error { return nil }); err != nil {
		return nil, err
	}
	call, cancel := context.WithTimeout(ctx, channelConnectorWait)
	bot, err := s.connector.VerifyTelegramBot(call, token)
	cancel()
	switch {
	case stderrors.Is(err, types.ErrConnectorCredential):
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("telegram rejected this bot token; copy it again from @BotFather")
	case err != nil:
		return nil, connectorError(ctx, types.ChannelPlatformTelegram, err)
	case bot.ID != strings.SplitN(token, ":", 2)[0]:
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("telegram rejected this bot token; copy it again from @BotFather")
	}
	return bot, nil
}

// ConnectTelegramBot verifies a Bot Token and binds the bot to the tenant. The partial unique index enforces
// 1 Bot = 1 tenant; a bot already bound to this tenant is re-activated in place with the new token.
func (s *ChannelService) ConnectTelegramBot(ctx context.Context, tenant, actor, rawToken string) (*types.ChannelConnection, *apperrors.Error) {
	if err := s.telegramConfigured(ctx); err != nil {
		return nil, err
	}
	token, err := telegramToken(ctx, rawToken)
	if err != nil {
		return nil, err
	}
	bot, err := s.verifyTelegramBot(ctx, tenant, actor, token)
	if err != nil {
		return nil, err
	}
	now := s.now()
	var result types.ChannelConnection
	err = s.authorized(ctx, tenant, actor, true, func(r repository.ChannelStore) *apperrors.Error {
		existing, err := r.TenantAccount(ctx, tenant, types.ChannelPlatformTelegram, bot.ID)
		if err != nil {
			return err
		}
		if existing == nil || existing.Status == types.ChannelStatusDisconnected {
			// A disconnected binding keeps its history; reconnecting creates a fresh identity.
			result = types.ChannelConnection{ID: uuid.NewString(), TenantID: tenant, Platform: types.ChannelPlatformTelegram, AccountID: bot.ID, AccountName: bot.DisplayName, AccountUsername: bot.Username, Status: types.ChannelStatusActive, Version: 1, Enabled: true, UpdatedAt: now}
			if err := r.InsertConnection(ctx, result); err != nil {
				return err
			}
		} else {
			result = *existing
			result.Version++
			result.AccountName, result.AccountUsername, result.Status, result.Enabled, result.LastErrorCode, result.LastErrorMessage, result.UpdatedAt = bot.DisplayName, bot.Username, types.ChannelStatusActive, true, "", "", now
			if err := r.SaveConnection(ctx, result, existing.Version); err != nil {
				return err
			}
		}
		return r.Audit(ctx, tenant, actor, "channel.telegram.connect", "messaging_connection", result.ID, map[string]any{"bot_id": bot.ID, "version": result.Version})
	})
	if err != nil {
		return nil, err
	}
	if err := s.sync(ctx, result, types.ConnectorConnection{BotToken: token}); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *ChannelService) ListTelegramConnections(ctx context.Context, tenant, actor string) ([]types.ChannelConnection, *apperrors.Error) {
	var result []types.ChannelConnection
	err := s.authorized(ctx, tenant, actor, false, func(r repository.ChannelStore) *apperrors.Error {
		var err *apperrors.Error
		result, err = r.Connections(ctx, tenant, types.ChannelPlatformTelegram)
		return err
	})
	return result, err
}

// UpdateTelegramConnection applies pause/resume/disconnect, or rotate_token. A rotated token must belong to the same
// bot; a different bot must be connected as a new connection. Rotation re-activates the connection.
func (s *ChannelService) UpdateTelegramConnection(ctx context.Context, tenant, actor, id, action, rawToken string) (*types.ChannelConnection, *apperrors.Error) {
	if err := s.telegramConfigured(ctx); err != nil {
		return nil, err
	}
	if action != "rotate_token" {
		if strings.TrimSpace(rawToken) != "" {
			return nil, apperrors.ErrBadRequest(ctx).SetDetail("bot_token is only accepted with rotate_token")
		}
		return s.updateConnection(ctx, types.ChannelPlatformTelegram, tenant, actor, id, action)
	}
	if _, err := uuid.Parse(id); err != nil {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("invalid connection ID")
	}
	token, err := telegramToken(ctx, rawToken)
	if err != nil {
		return nil, err
	}
	bot, err := s.verifyTelegramBot(ctx, tenant, actor, token)
	if err != nil {
		return nil, err
	}
	now := s.now()
	var result types.ChannelConnection
	err = s.authorized(ctx, tenant, actor, true, func(r repository.ChannelStore) *apperrors.Error {
		c, err := r.Connection(ctx, tenant, id, true)
		if err != nil {
			return err
		}
		if c.Platform != types.ChannelPlatformTelegram {
			return apperrors.ErrNotFound(ctx, "connection", "not found")
		}
		if c.Status == types.ChannelStatusDisconnected {
			return apperrors.ErrConflict(ctx, "connection", "is disconnected; connect the bot again")
		}
		if c.AccountID != bot.ID {
			return apperrors.ErrConflict(ctx, "bot token", "belongs to a different bot; connect it as a new bot")
		}
		result = *c
		result.Version++
		result.AccountName, result.AccountUsername, result.Status, result.Enabled, result.LastErrorCode, result.LastErrorMessage, result.UpdatedAt = bot.DisplayName, bot.Username, types.ChannelStatusActive, true, "", "", now
		if err := r.SaveConnection(ctx, result, c.Version); err != nil {
			return err
		}
		return r.Audit(ctx, tenant, actor, "channel.telegram.rotate_token", "messaging_connection", id, map[string]any{"version": result.Version, "status": result.Status})
	})
	if err != nil {
		return nil, err
	}
	if err := s.sync(ctx, result, types.ConnectorConnection{BotToken: token}); err != nil {
		return nil, err
	}
	return &result, nil
}
