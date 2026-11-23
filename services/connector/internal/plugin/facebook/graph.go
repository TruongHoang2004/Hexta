package facebook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/delivery"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin"
)

// Graph calls Meta Graph API. Tokens and the app secret travel only in request bodies/headers or the
// documented OAuth exchange query; returned errors never include URLs, bodies or credentials.
type Graph struct {
	BaseURL, Version, AppID string
	AppSecret               []byte
	Client                  *http.Client
}

func NewGraph(baseURL, version, appID string, secret []byte) *Graph {
	if baseURL == "" {
		baseURL = "https://graph.facebook.com"
	}
	if version == "" {
		version = "v21.0"
	}
	return &Graph{BaseURL: strings.TrimRight(baseURL, "/"), Version: version, AppID: appID, AppSecret: secret, Client: &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}

// Classify maps Meta error codes to outbound categories. See Messenger Send API error codes.
func Classify(status, code, subcode int) *plugin.ProviderError {
	c := strconv.Itoa(code)
	if subcode != 0 {
		c += "/" + strconv.Itoa(subcode)
	}
	switch {
	case code == 190 || (code >= 200 && code < 300):
		return &plugin.ProviderError{Category: delivery.ErrorAuthExpired, Code: c}
	case code == 10 || code == 2022 || subcode == 2018278:
		return &plugin.ProviderError{Category: delivery.ErrorWindow24h, Code: c}
	case code == 4 || code == 17 || code == 32 || code == 613 || status == http.StatusTooManyRequests:
		return &plugin.ProviderError{Category: delivery.ErrorRateLimit, Code: c}
	case code == 551 || subcode == 2018001 || subcode == 2018108:
		return &plugin.ProviderError{Category: delivery.ErrorInvalidRecipient, Code: c}
	case status >= 500:
		return &plugin.ProviderError{Category: delivery.ErrorTransport, Code: c, Unknown: true}
	}
	return &plugin.ProviderError{Category: delivery.ErrorTransport, Code: c}
}

func (g *Graph) endpoint(path string) string { return g.BaseURL + "/" + g.Version + "/" + path }

func (g *Graph) do(req *http.Request, out any) error {
	res, err := g.Client.Do(req)
	if err != nil {
		// Connection/timeout failures after dispatch cannot prove the provider did nothing.
		return &plugin.ProviderError{Category: delivery.ErrorTransport, Code: "network", Unknown: true}
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return &plugin.ProviderError{Category: delivery.ErrorTransport, Code: "read", Unknown: true}
	}
	if res.StatusCode != http.StatusOK {
		var body struct {
			Error struct {
				Code    int `json:"code"`
				Subcode int `json:"error_subcode"`
			} `json:"error"`
		}
		_ = json.Unmarshal(data, &body)
		return Classify(res.StatusCode, body.Error.Code, body.Error.Subcode)
	}
	if out != nil && json.Unmarshal(data, out) != nil {
		return &plugin.ProviderError{Category: delivery.ErrorTransport, Code: "decode", Unknown: true}
	}
	return nil
}

// ExchangeCode turns an OAuth code into Page credentials. A long-lived user token yields non-expiring Page tokens.
func (g *Graph) ExchangeCode(ctx context.Context, code, redirectURI string) ([]plugin.PageCredential, error) {
	var short struct {
		Token string `json:"access_token"`
	}
	q := url.Values{"client_id": {g.AppID}, "client_secret": {string(g.AppSecret)}, "redirect_uri": {redirectURI}, "code": {code}}
	if err := g.get(ctx, "oauth/access_token?"+q.Encode(), "", &short); err != nil {
		return nil, err
	}
	var long struct {
		Token string `json:"access_token"`
	}
	q = url.Values{"grant_type": {"fb_exchange_token"}, "client_id": {g.AppID}, "client_secret": {string(g.AppSecret)}, "fb_exchange_token": {short.Token}}
	if err := g.get(ctx, "oauth/access_token?"+q.Encode(), "", &long); err != nil || long.Token == "" {
		return nil, errors.New("oauth exchange failed")
	}
	var pages struct {
		Data []plugin.PageCredential `json:"data"`
	}
	if err := g.get(ctx, "me/accounts?fields=id,name,access_token&limit=100", long.Token, &pages); err != nil {
		return nil, err
	}
	result := make([]plugin.PageCredential, 0, len(pages.Data))
	for _, p := range pages.Data {
		if p.ID != "" && p.Token != "" && len(p.ID) <= 64 {
			if len([]rune(p.Name)) > 200 {
				p.Name = string([]rune(p.Name)[:200])
			}
			result = append(result, p)
		}
	}
	return result, nil
}

func (g *Graph) get(ctx context.Context, path, token string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.endpoint(path), nil)
	if err != nil {
		return errors.New("invalid graph request")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return g.do(req, out)
}

// Subscribe registers the app's webhook for the Page's message events.
func (g *Graph) Subscribe(ctx context.Context, pageID, token string) error {
	form := url.Values{"subscribed_fields": {"messages,message_echoes"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.endpoint(url.PathEscape(pageID)+"/subscribed_apps"), strings.NewReader(form.Encode()))
	if err != nil {
		return errors.New("invalid graph request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+token)
	return g.do(req, nil)
}

func (g *Graph) Unsubscribe(ctx context.Context, pageID, token string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, g.endpoint(url.PathEscape(pageID)+"/subscribed_apps"), nil)
	if err != nil {
		return errors.New("invalid graph request")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return g.do(req, nil)
}

// SendText performs one Send API call; callers must never retry an Unknown failure.
func (g *Graph) SendText(ctx context.Context, pageID, token, recipient, text string) (string, error) {
	body, _ := json.Marshal(map[string]any{"recipient": map[string]string{"id": recipient}, "messaging_type": "RESPONSE", "message": map[string]string{"text": text}})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.endpoint(url.PathEscape(pageID)+"/messages"), bytes.NewReader(body))
	if err != nil {
		return "", errors.New("invalid graph request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	var out struct {
		MessageID string `json:"message_id"`
	}
	if err := g.do(req, &out); err != nil {
		return "", err
	}
	return out.MessageID, nil
}

var _ plugin.Provisioner = (*Graph)(nil)
var _ plugin.TextSender = (*Graph)(nil)
