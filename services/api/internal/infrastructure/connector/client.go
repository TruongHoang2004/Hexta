// Package connector calls the connector's private machine control plane. It never decides tenant authority.
package connector

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
)

type Client struct {
	base, token string
	client      *http.Client
}

// NewClient returns a disabled client when the connector is unconfigured, so unrelated API use cases still start.
func NewClient(cfg *config.Config) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("connector configuration required")
	}
	base := strings.TrimRight(strings.TrimSpace(cfg.Connector.URL), "/")
	token := strings.TrimSpace(cfg.Connector.ManagementToken)
	if base == "" && token == "" {
		return &Client{}, nil
	}
	u, err := url.Parse(base)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || len(token) < 16 {
		return nil, errors.New("invalid connector configuration")
	}
	return &Client{base: base, token: token, client: &http.Client{Timeout: 45 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}

func (c *Client) call(ctx context.Context, method, path string, input, output any) error {
	if c.client == nil {
		return types.ErrConnectorUnavailable
	}
	body, err := json.Marshal(input)
	if err != nil {
		return types.ErrConnectorUnavailable
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, bytes.NewReader(body))
	if err != nil {
		return types.ErrConnectorUnavailable
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.client.Do(req)
	if err != nil {
		return types.ErrConnectorUnavailable
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return types.ErrConnectorUnavailable
	}
	switch res.StatusCode {
	case http.StatusOK:
	case http.StatusUnprocessableEntity:
		return types.ErrConnectorCredential
	case http.StatusBadGateway:
		return types.ErrConnectorProvider
	case http.StatusConflict:
		return types.ErrConnectorConflict
	default:
		return types.ErrConnectorUnavailable
	}
	if json.Unmarshal(data, output) != nil {
		return types.ErrConnectorUnavailable
	}
	return nil
}

func (c *Client) Exchange(ctx context.Context, code, redirectURI string) (*types.ConnectorExchange, error) {
	var out struct {
		OAuthSessionID string    `json:"oauth_session_id"`
		ExpiresAt      time.Time `json:"expires_at"`
		Pages          []struct {
			PageID   string `json:"page_id"`
			PageName string `json:"page_name"`
		} `json:"pages"`
	}
	if err := c.call(ctx, http.MethodPost, "/internal/facebook/oauth/exchange", map[string]string{"code": code, "redirect_uri": redirectURI}, &out); err != nil {
		return nil, err
	}
	if out.OAuthSessionID == "" || out.ExpiresAt.IsZero() || len(out.Pages) > 100 {
		return nil, types.ErrConnectorUnavailable
	}
	result := &types.ConnectorExchange{OAuthSessionID: out.OAuthSessionID, ExpiresAt: out.ExpiresAt, Pages: make([]types.ChannelPage, len(out.Pages))}
	for i, p := range out.Pages {
		result.Pages[i] = types.ChannelPage{PageID: p.PageID, PageName: p.PageName}
	}
	return result, nil
}

func (c *Client) Upsert(ctx context.Context, conn types.ConnectorConnection) (*types.ConnectorUpsertResult, error) {
	input := map[string]any{"id": conn.ID, "version": conn.Version, "tenant_id": conn.TenantID, "platform": conn.Platform, "account_id": conn.AccountID, "app_id": conn.AppID, "enabled": conn.Enabled, "credential_ref": ""}
	if conn.OAuthSessionID != "" {
		input["oauth_session_id"] = conn.OAuthSessionID
	}
	if conn.UnsubscribeOnDisable {
		input["unsubscribe_on_disable"] = true
	}
	var out struct {
		ID         string `json:"id"`
		Version    int64  `json:"version"`
		Enabled    bool   `json:"enabled"`
		Subscribed bool   `json:"subscribed"`
	}
	if err := c.call(ctx, http.MethodPut, "/internal/connections", input, &out); err != nil {
		return nil, err
	}
	return &types.ConnectorUpsertResult{ID: out.ID, Version: out.Version, Enabled: out.Enabled, Subscribed: out.Subscribed}, nil
}
