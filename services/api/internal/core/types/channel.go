package types

import (
	"errors"
	"time"
)

const (
	ChannelPlatformFacebook = "facebook"

	ChannelStatusActive       = "active"
	ChannelStatusPaused       = "paused"
	ChannelStatusDisconnected = "disconnected"
	ChannelStatusAuthExpired  = "auth_expired"

	PageAvailable                  = "available"
	PageAlreadyConnectedThisTenant = "already_connected_this_tenant"
	PageOwnedByOtherTenant         = "owned_by_other_tenant"
)

// ChannelConnection is the API-owned authority mirror of a provider account binding (1 Page = 1 tenant).
type ChannelConnection struct {
	ID, TenantID, Platform, AccountID, AccountName, Status string
	Version                                                int64
	Enabled                                                bool
	LastErrorCode, LastErrorMessage                        string
	UpdatedAt                                              time.Time
}

type ChannelPage struct {
	PageID, PageName, Availability string
}

// ChannelOAuthState is a single-use, owner-bound OAuth flow. Only the state hash is stored.
type ChannelOAuthState struct {
	ID, TenantID, ActorID, Platform, StateHash, RedirectURI string
	ExpiresAt, CreatedAt                                    time.Time
	ConsumedAt, SessionExpiresAt                            *time.Time
	OAuthSessionID                                          string
	Pages                                                   []ChannelPage
}

type ChannelOAuthStart struct {
	AuthURL   string
	ExpiresAt time.Time
}

type ChannelOAuthPages struct {
	TenantID, OAuthSessionID string
	ExpiresAt                time.Time
	Pages                    []ChannelPage
}

// ConnectorExchange is the connector's OAuth exchange result; Page tokens never leave the connector.
type ConnectorExchange struct {
	OAuthSessionID string
	ExpiresAt      time.Time
	Pages          []ChannelPage
}

// ConnectorConnection is trusted machine configuration sent to the connector after API authorization.
type ConnectorConnection struct {
	ID, TenantID, Platform, AccountID, AppID string
	Version                                  int64
	Enabled                                  bool
	OAuthSessionID                           string
	UnsubscribeOnDisable                     bool
}

type ConnectorUpsertResult struct {
	ID         string
	Version    int64
	Enabled    bool
	Subscribed bool
}

// Connector adapter failures, mapped by services to safe public errors.
var (
	ErrConnectorUnavailable = errors.New("channel connector unavailable")
	ErrConnectorCredential  = errors.New("channel credential unavailable")
	ErrConnectorProvider    = errors.New("channel provider unavailable")
	ErrConnectorConflict    = errors.New("channel configuration conflict")
)

// Outbound delivery values. Message delivery_status follows the intent: sending until a result arrives.
const (
	DeliveryReceived = "received"
	DeliverySending  = "sending"
	DeliverySent     = "sent"
	DeliveryFailed   = "failed"
	DeliveryUnknown  = "unknown"

	IntentPending   = "pending"
	IntentCancelled = "cancelled"

	ContentText             = "text"
	ContentUnsupportedMedia = "unsupported_media"
)

type ChannelSendInput struct {
	ClientMessageID, Text            string
	ExpectedVersion, ContextRevision int64
}

type OutboundIntent struct {
	ID, TenantID, ConnectionID, MessageID, ClientMessageID, RequestHash, RecipientID, Status string
	ErrorCategory, ErrorCode                                                                 string
	CreatedAt, UpdatedAt                                                                     time.Time
}

type OutboundAttempt struct {
	ID, TenantID, IntentID, Status, ProviderMessageID string
	ConnectionVersion                                 int64
	CreatedAt                                         time.Time
	CompletedAt                                       *time.Time
}
