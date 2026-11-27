package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/delivery"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin"
)

// tokenPattern is the BotFather token shape: numeric bot ID, colon, secret.
var tokenPattern = regexp.MustCompile(`^[0-9]{1,20}:[A-Za-z0-9_-]{30,80}$`)

var botIDPattern = regexp.MustCompile(`^[0-9]{1,20}$`)

// ValidBotID reports whether an account ID is a numeric Telegram bot ID.
func ValidBotID(id string) bool { return botIDPattern.MatchString(id) }

// ErrInvalidToken means the token is malformed or Telegram rejected it (401/404).
var ErrInvalidToken = plugin.ErrInvalidCredential

// ValidToken reports whether a token has the BotFather shape. It does not prove the token is live.
func ValidToken(token string) bool { return tokenPattern.MatchString(token) }

// TokenBotID returns the bot ID embedded in a well-formed token.
func TokenBotID(token string) string {
	id, _, _ := strings.Cut(token, ":")
	return id
}

// BotAPI calls the Telegram Bot API. The token is part of the request path, so neither URLs nor transport errors
// are ever returned or logged: every failure is a sanitized ProviderError or ErrInvalidToken.
type BotAPI struct {
	BaseURL string
	Client  *http.Client
}

func NewBotAPI(baseURL string) *BotAPI {
	if baseURL == "" {
		baseURL = "https://api.telegram.org"
	}
	return &BotAPI{BaseURL: strings.TrimRight(baseURL, "/"), Client: &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}

// Classify maps Bot API failures to outbound categories. Telegram reports errors as an HTTP status plus a description.
func Classify(status int, description string) *plugin.ProviderError {
	code := strconv.Itoa(status)
	d := strings.ToLower(description)
	switch {
	case status == http.StatusUnauthorized || status == http.StatusNotFound:
		return &plugin.ProviderError{Category: delivery.ErrorAuthExpired, Code: code}
	case status == http.StatusForbidden:
		// "bot was blocked by the user", "user is deactivated", "bot can't initiate conversation with a user".
		return &plugin.ProviderError{Category: delivery.ErrorBotBlocked, Code: code}
	case status == http.StatusBadRequest && strings.Contains(d, "chat not found"):
		return &plugin.ProviderError{Category: delivery.ErrorChatNotFound, Code: code}
	case status == http.StatusTooManyRequests:
		return &plugin.ProviderError{Category: delivery.ErrorRateLimit, Code: code}
	case status >= 500:
		return &plugin.ProviderError{Category: delivery.ErrorTransport, Code: code, Unknown: true}
	}
	return &plugin.ProviderError{Category: delivery.ErrorTransport, Code: code}
}

func (b *BotAPI) call(ctx context.Context, token, method string, input, output any) error {
	if !ValidToken(token) {
		return ErrInvalidToken
	}
	body, err := json.Marshal(input)
	if err != nil {
		return errors.New("invalid bot api request")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.BaseURL+"/bot"+token+"/"+method, bytes.NewReader(body))
	if err != nil {
		return errors.New("invalid bot api request")
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := b.Client.Do(req)
	if err != nil {
		// Connection/timeout failures after dispatch cannot prove Telegram did nothing.
		return &plugin.ProviderError{Category: delivery.ErrorTransport, Code: "network", Unknown: true}
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return &plugin.ProviderError{Category: delivery.ErrorTransport, Code: "read", Unknown: true}
	}
	var envelope struct {
		OK          bool            `json:"ok"`
		Description string          `json:"description"`
		Result      json.RawMessage `json:"result"`
	}
	_ = json.Unmarshal(data, &envelope)
	if res.StatusCode != http.StatusOK || !envelope.OK {
		if res.StatusCode == http.StatusOK {
			return &plugin.ProviderError{Category: delivery.ErrorTransport, Code: "decode", Unknown: true}
		}
		return Classify(res.StatusCode, envelope.Description)
	}
	if output != nil && json.Unmarshal(envelope.Result, output) != nil {
		return &plugin.ProviderError{Category: delivery.ErrorTransport, Code: "decode", Unknown: true}
	}
	return nil
}

// GetMe verifies a token and returns the bot identity. A rejected token returns ErrInvalidToken.
func (b *BotAPI) GetMe(ctx context.Context, token string) (plugin.BotIdentity, error) {
	var me struct {
		ID        int64  `json:"id"`
		IsBot     bool   `json:"is_bot"`
		FirstName string `json:"first_name"`
		Username  string `json:"username"`
	}
	err := b.call(ctx, token, "getMe", struct{}{}, &me)
	var pe *plugin.ProviderError
	if errors.As(err, &pe) && pe.Category == delivery.ErrorAuthExpired {
		return plugin.BotIdentity{}, ErrInvalidToken
	}
	if err != nil {
		return plugin.BotIdentity{}, err
	}
	id := strconv.FormatInt(me.ID, 10)
	if !me.IsBot || id != TokenBotID(token) {
		return plugin.BotIdentity{}, ErrInvalidToken
	}
	name := me.FirstName
	if len([]rune(name)) > 200 {
		name = string([]rune(name)[:200])
	}
	if len(me.Username) > 64 {
		me.Username = me.Username[:64]
	}
	return plugin.BotIdentity{ID: id, DisplayName: name, Username: me.Username}, nil
}

// SetWebhook registers the per-connection HTTPS callback with its secret_token, for private message updates only.
func (b *BotAPI) SetWebhook(ctx context.Context, token, callback, secret string) error {
	return b.call(ctx, token, "setWebhook", map[string]any{"url": callback, "secret_token": secret, "allowed_updates": []string{"message"}}, nil)
}

// DeleteWebhook stops delivery and drops updates queued while the bot is paused or disconnected.
func (b *BotAPI) DeleteWebhook(ctx context.Context, token string) error {
	return b.call(ctx, token, "deleteWebhook", map[string]any{"drop_pending_updates": true}, nil)
}

// SendText performs one sendMessage call; callers must never retry an Unknown failure.
func (b *BotAPI) SendText(ctx context.Context, _, token, chat, text string) (string, error) {
	var sent struct {
		MessageID int64 `json:"message_id"`
	}
	if err := b.call(ctx, token, "sendMessage", map[string]string{"chat_id": chat, "text": text}, &sent); err != nil {
		if errors.Is(err, ErrInvalidToken) {
			return "", &plugin.ProviderError{Category: delivery.ErrorAuthExpired, Code: "invalid_token"}
		}
		return "", err
	}
	return strconv.FormatInt(sent.MessageID, 10), nil
}

var _ plugin.TextSender = (*BotAPI)(nil)
var _ plugin.BotProvisioner = (*BotAPI)(nil)
