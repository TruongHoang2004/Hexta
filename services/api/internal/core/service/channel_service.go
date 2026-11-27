package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	stderrors "errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
)

const (
	channelOAuthTTL       = 10 * time.Minute
	channelCallbackPath   = "/channels/facebook/callback"
	channelMaxPages       = 20
	channelConnectorWait  = 40 * time.Second
	channelFacebookScopes = "pages_show_list,pages_messaging,pages_manage_metadata"
)

// IChannelConnector is the connector machine control plane. Page tokens never cross this boundary; a Telegram Bot
// Token crosses it only inbound, on connect/rotate, and is never returned.
type IChannelConnector interface {
	Exchange(ctx context.Context, code, redirectURI string) (*types.ConnectorExchange, error)
	Upsert(ctx context.Context, conn types.ConnectorConnection) (*types.ConnectorUpsertResult, error)
	VerifyTelegramBot(ctx context.Context, token string) (*types.TelegramBot, error)
}

// ChannelSettings is non-secret Meta app configuration plus the browser origins allowed as OAuth redirects.
type ChannelSettings struct {
	AppID, GraphVersion, DialogURL string
	AllowedOrigins                 []string
}

type ChannelService struct {
	store     repository.ChannelStore
	connector IChannelConnector
	settings  ChannelSettings
	now       func() time.Time
}

func NewChannelService(store repository.ChannelStore, connector IChannelConnector, settings ChannelSettings) *ChannelService {
	if settings.GraphVersion == "" {
		settings.GraphVersion = "v21.0"
	}
	if settings.DialogURL == "" {
		settings.DialogURL = "https://www.facebook.com"
	}
	settings.DialogURL = strings.TrimRight(settings.DialogURL, "/")
	return &ChannelService{store: store, connector: connector, settings: settings, now: func() time.Time { return time.Now().UTC() }}
}

func channelStateHash(state string) string {
	sum := sha256.Sum256([]byte(state))
	return hex.EncodeToString(sum[:])
}

// authorized rechecks current membership inside the transaction; owner is required for management actions.
func (s *ChannelService) authorized(ctx context.Context, tenant, actor string, owner bool, fn func(repository.ChannelStore) *apperrors.Error) *apperrors.Error {
	if actor == "" {
		return apperrors.ErrUnauthorized(ctx)
	}
	if _, err := uuid.Parse(tenant); err != nil {
		return apperrors.ErrBadRequest(ctx).SetDetail("valid tenant ID required")
	}
	return s.store.Within(ctx, func(r repository.ChannelStore) *apperrors.Error {
		scope, err := r.Scope(ctx, tenant, actor)
		if err != nil {
			return err
		}
		if !scope.Active || !scope.Member {
			return apperrors.ErrForbidden(ctx).SetDetail("active tenant membership required")
		}
		if owner && !scope.Owner {
			return apperrors.ErrForbidden(ctx).SetDetail("tenant owner required")
		}
		return fn(r)
	})
}

func (s *ChannelService) configured(ctx context.Context) *apperrors.Error {
	if s.settings.AppID == "" || s.connector == nil {
		return apperrors.ErrSystemError(ctx, "facebook channel is not configured").SetHTTPStatus(http.StatusServiceUnavailable)
	}
	return nil
}

func (s *ChannelService) validRedirect(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != channelCallbackPath {
		return false
	}
	origin := u.Scheme + "://" + u.Host
	for _, allowed := range s.settings.AllowedOrigins {
		if allowed != "" && origin == strings.TrimRight(allowed, "/") {
			return true
		}
	}
	return false
}

// connectorError maps adapter failures to safe public errors.
func connectorError(ctx context.Context, platform string, err error) *apperrors.Error {
	switch {
	case stderrors.Is(err, types.ErrConnectorCredential) && platform == types.ChannelPlatformTelegram:
		return apperrors.ErrSystemError(ctx, "telegram rejected the bot token; rotate the token").SetHTTPStatus(http.StatusUnprocessableEntity)
	case stderrors.Is(err, types.ErrConnectorCredential):
		return apperrors.ErrSystemError(ctx, "facebook authorization expired; connect again").SetHTTPStatus(http.StatusUnprocessableEntity)
	case stderrors.Is(err, types.ErrConnectorProvider):
		return apperrors.ErrSystemError(ctx, platform+" is unavailable; try again").SetHTTPStatus(http.StatusBadGateway)
	case stderrors.Is(err, types.ErrConnectorConflict):
		return apperrors.ErrConflict(ctx, "channel account", "is managed by another configuration")
	}
	return apperrors.ErrSystemError(ctx, "channel connector unavailable; try again").SetHTTPStatus(http.StatusServiceUnavailable)
}

func connectorErrorCode(err error) string {
	switch {
	case stderrors.Is(err, types.ErrConnectorCredential):
		return "credential_expired"
	case stderrors.Is(err, types.ErrConnectorProvider):
		return "provider_unavailable"
	}
	return "connector_sync_failed"
}

// StartFacebookOAuth issues a random single-use state bound to the owner, tenant and redirect URI.
func (s *ChannelService) StartFacebookOAuth(ctx context.Context, tenant, actor, redirectURI string) (*types.ChannelOAuthStart, *apperrors.Error) {
	if err := s.configured(ctx); err != nil {
		return nil, err
	}
	if !s.validRedirect(redirectURI) {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("redirect_uri must be the web app channel callback")
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, apperrors.ErrSystemError(ctx, "could not start authorization")
	}
	state := base64.RawURLEncoding.EncodeToString(raw)
	now := s.now()
	result := &types.ChannelOAuthStart{ExpiresAt: now.Add(channelOAuthTTL)}
	err := s.authorized(ctx, tenant, actor, true, func(r repository.ChannelStore) *apperrors.Error {
		return r.SaveOAuthState(ctx, types.ChannelOAuthState{ID: uuid.NewString(), TenantID: tenant, ActorID: actor, Platform: types.ChannelPlatformFacebook, StateHash: channelStateHash(state), RedirectURI: redirectURI, ExpiresAt: result.ExpiresAt, CreatedAt: now})
	})
	if err != nil {
		return nil, err
	}
	q := url.Values{"client_id": {s.settings.AppID}, "redirect_uri": {redirectURI}, "state": {state}, "response_type": {"code"}, "scope": {channelFacebookScopes}}
	result.AuthURL = s.settings.DialogURL + "/" + s.settings.GraphVersion + "/dialog/oauth?" + q.Encode()
	return result, nil
}

// availability classifies Pages against current bindings. It reveals only that a Page is taken, not by whom.
func availability(pages []types.ChannelPage, tenant string, bound []types.ChannelConnection) []types.ChannelPage {
	owner := map[string]string{}
	for _, c := range bound {
		owner[c.AccountID] = c.TenantID
	}
	result := make([]types.ChannelPage, len(pages))
	for i, p := range pages {
		p.Availability = types.PageAvailable
		if t, ok := owner[p.PageID]; ok {
			p.Availability = types.PageOwnedByOtherTenant
			if t == tenant {
				p.Availability = types.PageAlreadyConnectedThisTenant
			}
		}
		result[i] = p
	}
	return result
}

func pageIDs(pages []types.ChannelPage) []string {
	ids := make([]string, len(pages))
	for i, p := range pages {
		ids[i] = p.PageID
	}
	return ids
}

// CompleteFacebookOAuth consumes the state before exchanging the code, so a replayed callback cannot exchange twice.
func (s *ChannelService) CompleteFacebookOAuth(ctx context.Context, actor, code, state, providerError string) (*types.ChannelOAuthPages, *apperrors.Error) {
	if err := s.configured(ctx); err != nil {
		return nil, err
	}
	if actor == "" {
		return nil, apperrors.ErrUnauthorized(ctx)
	}
	if state == "" || len(state) > 128 || (providerError == "" && (code == "" || len(code) > 2048)) {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("code and state required")
	}
	var flow *types.ChannelOAuthState
	err := s.store.Within(ctx, func(r repository.ChannelStore) *apperrors.Error {
		var err *apperrors.Error
		flow, err = r.ConsumeOAuthState(ctx, channelStateHash(state), s.now())
		return err
	})
	if err != nil {
		return nil, err
	}
	if flow == nil || flow.ActorID != actor {
		return nil, apperrors.ErrForbidden(ctx).SetDetail("authorization state is invalid, expired or already used")
	}
	if providerError != "" {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("facebook authorization was cancelled or denied")
	}
	// The owner must still hold authority after the provider round trip.
	if err := s.authorized(ctx, flow.TenantID, actor, true, func(repository.ChannelStore) *apperrors.Error { return nil }); err != nil {
		return nil, err
	}
	call, cancel := context.WithTimeout(ctx, channelConnectorWait)
	exchange, exchangeErr := s.connector.Exchange(call, code, flow.RedirectURI)
	cancel()
	if exchangeErr != nil {
		return nil, connectorError(ctx, types.ChannelPlatformFacebook, exchangeErr)
	}
	result := &types.ChannelOAuthPages{TenantID: flow.TenantID, OAuthSessionID: exchange.OAuthSessionID, ExpiresAt: exchange.ExpiresAt}
	err = s.authorized(ctx, flow.TenantID, actor, true, func(r repository.ChannelStore) *apperrors.Error {
		if err := r.AttachOAuthSession(ctx, flow.ID, exchange.OAuthSessionID, exchange.ExpiresAt, exchange.Pages); err != nil {
			return err
		}
		bound, err := r.BoundAccounts(ctx, types.ChannelPlatformFacebook, pageIDs(exchange.Pages))
		if err != nil {
			return err
		}
		result.Pages = availability(exchange.Pages, flow.TenantID, bound)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ListConnections returns the tenant's current connections of every type, or of one platform when platform is set.
// Any active member may read them (Inbox banners and composer locks); credentials are never part of the result.
func (s *ChannelService) ListConnections(ctx context.Context, tenant, actor, platform string) ([]types.ChannelConnection, *apperrors.Error) {
	if _, ok := types.ChannelPlatforms[platform]; platform != "" && !ok {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("unknown connection platform")
	}
	result := []types.ChannelConnection{}
	err := s.authorized(ctx, tenant, actor, false, func(r repository.ChannelStore) *apperrors.Error {
		list, err := r.Connections(ctx, tenant, platform)
		if list != nil {
			result = list
		}
		return err
	})
	return result, err
}

func (s *ChannelService) ListFacebookConnections(ctx context.Context, tenant, actor string) ([]types.ChannelConnection, *apperrors.Error) {
	return s.ListConnections(ctx, tenant, actor, types.ChannelPlatformFacebook)
}

// UpdateConnection applies an owner action to any connection: pause, resume and disconnect for every platform, and
// rotate_token where the platform supports credential rotation. bot_token is accepted only with rotate_token.
func (s *ChannelService) UpdateConnection(ctx context.Context, tenant, actor, id, action, credential string) (*types.ChannelConnection, *apperrors.Error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("invalid connection ID")
	}
	var current *types.ChannelConnection
	err := s.authorized(ctx, tenant, actor, true, func(r repository.ChannelStore) *apperrors.Error {
		var err *apperrors.Error
		current, err = r.Connection(ctx, tenant, id, false)
		return err
	})
	if err != nil {
		return nil, err
	}
	if action == types.ChannelActionRotateToken && !types.ChannelPlatforms[current.Platform].RotatesCredential {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("this connection type does not support rotate_token")
	}
	switch current.Platform {
	case types.ChannelPlatformTelegram:
		return s.UpdateTelegramConnection(ctx, tenant, actor, id, action, credential)
	case types.ChannelPlatformFacebook:
		if strings.TrimSpace(credential) != "" {
			return nil, apperrors.ErrBadRequest(ctx).SetDetail("bot_token is only accepted with rotate_token")
		}
		return s.UpdateFacebookConnection(ctx, tenant, actor, id, action)
	default:
		return nil, apperrors.ErrNotFound(ctx, "connection", "not found")
	}
}

// sync pushes committed configuration to the connector. A failure is recorded on the row; retrying the same
// action re-syncs because connector versions only need to increase. extra carries the per-call credential fields
// (OAuth session, Bot Token, disconnect wipe).
func (s *ChannelService) sync(ctx context.Context, c types.ChannelConnection, extra types.ConnectorConnection) *apperrors.Error {
	app := s.settings.AppID
	if c.Platform == types.ChannelPlatformTelegram {
		app = types.ChannelAppTelegram
	}
	call, cancel := context.WithTimeout(ctx, channelConnectorWait)
	_, err := s.connector.Upsert(call, types.ConnectorConnection{ID: c.ID, TenantID: c.TenantID, Platform: c.Platform, AccountID: c.AccountID, AppID: app, Version: c.Version, Enabled: c.Enabled, OAuthSessionID: extra.OAuthSessionID, UnsubscribeOnDisable: extra.UnsubscribeOnDisable, BotToken: extra.BotToken})
	cancel()
	if err == nil {
		return nil
	}
	_ = s.store.Within(ctx, func(r repository.ChannelStore) *apperrors.Error {
		return r.RecordConnectionError(ctx, c.TenantID, c.ID, c.Version, connectorErrorCode(err), "channel sync failed; retry the action")
	})
	return connectorError(ctx, c.Platform, err)
}

// ConnectFacebookPages binds selected Pages from the owner's own OAuth session. The partial unique index enforces
// 1 Page = 1 tenant; a Page already bound to this tenant is re-authorized in place (reconnect).
func (s *ChannelService) ConnectFacebookPages(ctx context.Context, tenant, actor, session string, pages []string) ([]types.ChannelConnection, *apperrors.Error) {
	if err := s.configured(ctx); err != nil {
		return nil, err
	}
	if _, err := uuid.Parse(session); err != nil || len(pages) == 0 || len(pages) > channelMaxPages {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("oauth_session_id and 1–20 page_ids required")
	}
	seen := map[string]bool{}
	for _, p := range pages {
		if p == "" || seen[p] {
			return nil, apperrors.ErrBadRequest(ctx).SetDetail("page_ids must be unique")
		}
		seen[p] = true
	}
	now := s.now()
	var result []types.ChannelConnection
	err := s.authorized(ctx, tenant, actor, true, func(r repository.ChannelStore) *apperrors.Error {
		flow, err := r.OAuthSessionState(ctx, tenant, actor, session, now)
		if err != nil {
			return err
		}
		if flow == nil {
			return apperrors.ErrSystemError(ctx, "facebook authorization expired; connect again").SetHTTPStatus(http.StatusUnprocessableEntity)
		}
		names := map[string]string{}
		for _, p := range flow.Pages {
			names[p.PageID] = p.PageName
		}
		for _, page := range pages {
			name, ok := names[page]
			if !ok {
				return apperrors.ErrBadRequest(ctx).SetDetail("page was not granted in this authorization")
			}
			existing, err := r.TenantAccount(ctx, tenant, types.ChannelPlatformFacebook, page)
			if err != nil {
				return err
			}
			if existing == nil || existing.Status == types.ChannelStatusDisconnected {
				// A disconnected binding keeps its history; reconnecting creates a fresh identity.
				c := types.ChannelConnection{ID: uuid.NewString(), TenantID: tenant, Platform: types.ChannelPlatformFacebook, AccountID: page, AccountName: name, Status: types.ChannelStatusActive, Version: 1, Enabled: true, UpdatedAt: now}
				if err := r.InsertConnection(ctx, c); err != nil {
					return err
				}
				result = append(result, c)
				continue
			}
			c := *existing
			c.Version++
			c.AccountName, c.Status, c.Enabled, c.LastErrorCode, c.LastErrorMessage, c.UpdatedAt = name, types.ChannelStatusActive, true, "", "", now
			if err := r.SaveConnection(ctx, c, existing.Version); err != nil {
				return err
			}
			result = append(result, c)
		}
		return r.Audit(ctx, tenant, actor, "channel.facebook.connect", "messaging_connection", tenant, map[string]any{"page_ids": pages})
	})
	if err != nil {
		return nil, err
	}
	for i, c := range result {
		if err := s.sync(ctx, c, types.ConnectorConnection{OAuthSessionID: session}); err != nil {
			return nil, err
		}
		result[i].LastErrorCode, result[i].LastErrorMessage = "", ""
	}
	return result, nil
}

// UpdateFacebookConnection applies pause/resume/disconnect. Repeating an action already in effect re-syncs the
// connector without a version change, which recovers from a previously failed sync.
func (s *ChannelService) UpdateFacebookConnection(ctx context.Context, tenant, actor, id, action string) (*types.ChannelConnection, *apperrors.Error) {
	if err := s.configured(ctx); err != nil {
		return nil, err
	}
	return s.updateConnection(ctx, types.ChannelPlatformFacebook, tenant, actor, id, action)
}

// updateConnection applies pause/resume/disconnect to one platform's connection. Pause and disconnect cancel the
// connection's pending outbound intents in the same transaction.
func (s *ChannelService) updateConnection(ctx context.Context, platform, tenant, actor, id, action string) (*types.ChannelConnection, *apperrors.Error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("invalid connection ID")
	}
	target := map[string]string{"pause": types.ChannelStatusPaused, "resume": types.ChannelStatusActive, "disconnect": types.ChannelStatusDisconnected}[action]
	if target == "" {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("action must be pause, resume or disconnect")
	}
	now := s.now()
	var result types.ChannelConnection
	err := s.authorized(ctx, tenant, actor, true, func(r repository.ChannelStore) *apperrors.Error {
		c, err := r.Connection(ctx, tenant, id, true)
		if err != nil {
			return err
		}
		if c.Platform != platform {
			return apperrors.ErrNotFound(ctx, "connection", "not found")
		}
		result = *c
		if c.Status == target {
			return nil
		}
		switch action {
		case "pause":
			if c.Status != types.ChannelStatusActive && c.Status != types.ChannelStatusAuthExpired {
				return apperrors.ErrConflict(ctx, "connection", "cannot be paused in its current state")
			}
		case "resume":
			if c.Status != types.ChannelStatusPaused {
				return apperrors.ErrConflict(ctx, "connection", "is not paused")
			}
		}
		result.Version++
		result.Status, result.Enabled, result.LastErrorCode, result.LastErrorMessage, result.UpdatedAt = target, target == types.ChannelStatusActive, "", "", now
		if err := r.SaveConnection(ctx, result, c.Version); err != nil {
			return err
		}
		if !result.Enabled {
			if err := r.CancelPendingIntents(ctx, tenant, id, "connection_"+target, now); err != nil {
				return err
			}
		}
		return r.Audit(ctx, tenant, actor, "channel."+platform+"."+action, "messaging_connection", id, map[string]any{"version": result.Version, "status": result.Status})
	})
	if err != nil {
		return nil, err
	}
	if err := s.sync(ctx, result, types.ConnectorConnection{UnsubscribeOnDisable: action == "disconnect"}); err != nil {
		return nil, err
	}
	return &result, nil
}
