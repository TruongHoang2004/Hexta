// Package persistencemap converts channel connection, OAuth state and outbound rows to core values.
package persistencemap

import (
	"encoding/json"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/messagingmodel"
)

func optional(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
func text(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func Connection(row messagingmodel.Connection) types.ChannelConnection {
	return types.ChannelConnection{ID: row.ID, TenantID: row.TenantID, Platform: row.Platform, AccountID: row.AccountID, AccountName: row.AccountName, AccountUsername: row.AccountUsername, Status: row.Status, Version: row.Version, Enabled: row.Enabled, LastErrorCode: text(row.LastErrorCode), LastErrorMessage: text(row.LastErrorMessage), UpdatedAt: row.UpdatedAt}
}
func ConnectionRow(value types.ChannelConnection) messagingmodel.Connection {
	return messagingmodel.Connection{ID: value.ID, TenantID: value.TenantID, Platform: value.Platform, AccountID: value.AccountID, AccountName: value.AccountName, AccountUsername: value.AccountUsername, Status: value.Status, Version: value.Version, Enabled: value.Enabled, LastErrorCode: optional(value.LastErrorCode), LastErrorMessage: optional(value.LastErrorMessage), UpdatedAt: value.UpdatedAt}
}

type storedPage struct {
	PageID   string `json:"page_id"`
	PageName string `json:"page_name"`
}

// Stored pages keep identity/name only; availability is recomputed from current bindings when used.
func OAuthState(row messagingmodel.OAuthState) types.ChannelOAuthState {
	value := types.ChannelOAuthState{ID: row.ID, TenantID: row.TenantID, ActorID: row.ActorID, Platform: row.Platform, StateHash: row.StateHash, RedirectURI: row.RedirectURI, ExpiresAt: row.ExpiresAt, CreatedAt: row.CreatedAt, ConsumedAt: row.ConsumedAt, SessionExpiresAt: row.OAuthSessionExpiresAt, OAuthSessionID: text(row.OAuthSessionID)}
	if row.Pages != nil {
		var pages []storedPage
		if json.Unmarshal([]byte(*row.Pages), &pages) == nil {
			for _, p := range pages {
				value.Pages = append(value.Pages, types.ChannelPage{PageID: p.PageID, PageName: p.PageName})
			}
		}
	}
	return value
}
func OAuthStateRow(value types.ChannelOAuthState) messagingmodel.OAuthState {
	return messagingmodel.OAuthState{ID: value.ID, TenantID: value.TenantID, ActorID: value.ActorID, Platform: value.Platform, StateHash: value.StateHash, RedirectURI: value.RedirectURI, ExpiresAt: value.ExpiresAt, CreatedAt: value.CreatedAt, ConsumedAt: value.ConsumedAt, OAuthSessionExpiresAt: value.SessionExpiresAt, OAuthSessionID: optional(value.OAuthSessionID), Pages: Pages(value.Pages)}
}
func Pages(pages []types.ChannelPage) *string {
	if pages == nil {
		return nil
	}
	stored := make([]storedPage, len(pages))
	for i, p := range pages {
		stored[i] = storedPage{PageID: p.PageID, PageName: p.PageName}
	}
	data, _ := json.Marshal(stored)
	result := string(data)
	return &result
}

func Intent(row messagingmodel.OutboundIntent) types.OutboundIntent {
	return types.OutboundIntent{ID: row.ID, TenantID: row.TenantID, ConnectionID: row.ConnectionID, MessageID: row.MessageID, ClientMessageID: row.ClientMessageID, RequestHash: row.RequestHash, RecipientID: row.RecipientID, Status: row.Status, ErrorCategory: text(row.ErrorCategory), ErrorCode: text(row.ErrorCode), CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}
func IntentRow(value types.OutboundIntent) messagingmodel.OutboundIntent {
	return messagingmodel.OutboundIntent{ID: value.ID, TenantID: value.TenantID, ConnectionID: value.ConnectionID, MessageID: value.MessageID, ClientMessageID: value.ClientMessageID, RequestHash: value.RequestHash, RecipientID: value.RecipientID, Status: value.Status, ErrorCategory: optional(value.ErrorCategory), ErrorCode: optional(value.ErrorCode), CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt}
}
func Attempt(row messagingmodel.OutboundAttempt) types.OutboundAttempt {
	return types.OutboundAttempt{ID: row.ID, TenantID: row.TenantID, IntentID: row.IntentID, Status: row.Status, ProviderMessageID: text(row.ProviderMessageID), ConnectionVersion: row.ConnectionVersion, CreatedAt: row.CreatedAt, CompletedAt: row.CompletedAt}
}
func AttemptRow(value types.OutboundAttempt) messagingmodel.OutboundAttempt {
	return messagingmodel.OutboundAttempt{ID: value.ID, TenantID: value.TenantID, IntentID: value.IntentID, Status: value.Status, ProviderMessageID: optional(value.ProviderMessageID), ConnectionVersion: value.ConnectionVersion, CreatedAt: value.CreatedAt, CompletedAt: value.CompletedAt}
}
