// Package telegram is the Telegram Bot API adapter: per-connection webhook verification, private-chat update
// normalization and Bot API calls. Bot Tokens and webhook secrets never appear in returned errors or events.
package telegram

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/ingress"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin"
)

// AppID is the connector app identity shared by all Telegram connections; each bot has its own credential.
const AppID = "telegram"

// SecretHeader carries the per-connection secret_token registered through setWebhook.
const SecretHeader = "X-Telegram-Bot-Api-Secret-Token"

type TelegramPlugin struct{ plugin.BasePlugin }

func (TelegramPlugin) Platform() string                  { return "telegram" }
func (TelegramPlugin) Capabilities() plugin.Capabilities { return plugin.Capabilities{Webhook: true} }

// Verify compares the secret header with the connection's webhook secret in constant time.
func (TelegramPlugin) Verify(_ []byte, header string, secret []byte) error {
	if len(secret) == 0 || subtle.ConstantTimeCompare([]byte(header), secret) != 1 {
		return errors.New("invalid secret token")
	}
	return nil
}

func (TelegramPlugin) Challenge(url.Values, string) (string, error) { return "", plugin.ErrUnsupported }

// mediaKinds maps Telegram message fields to placeholder kinds; media is never downloaded.
var mediaKinds = []struct {
	field, kind string
}{{"photo", "image"}, {"video", "video"}, {"animation", "video"}, {"video_note", "video"}, {"audio", "audio"}, {"voice", "audio"}, {"document", "file"}, {"sticker", "sticker"}}

// Parse normalizes one webhook Update. Updates name the chat but not the bot, so AccountID and Recipient stay empty
// and the inbound service binds them from the per-connection webhook URL. Non-message updates and non-private chats
// return no items: they are dropped before the outbox and acknowledged so Telegram does not retry them.
func (TelegramPlugin) Parse(body []byte, received time.Time) ([]plugin.Item, error) {
	var update struct {
		UpdateID *int64          `json:"update_id"`
		Message  json.RawMessage `json:"message"`
	}
	if json.Unmarshal(body, &update) != nil || update.UpdateID == nil {
		return nil, errors.New("invalid update")
	}
	if len(update.Message) == 0 || string(update.Message) == "null" {
		return nil, nil
	}
	var fields map[string]json.RawMessage
	var m struct {
		MessageID int64 `json:"message_id"`
		Date      int64 `json:"date"`
		From      *struct {
			ID int64 `json:"id"`
		} `json:"from"`
		Chat struct {
			ID   int64  `json:"id"`
			Type string `json:"type"`
		} `json:"chat"`
		Text  *string `json:"text"`
		Reply *struct {
			MessageID int64 `json:"message_id"`
		} `json:"reply_to_message"`
	}
	if json.Unmarshal(update.Message, &fields) != nil || json.Unmarshal(update.Message, &m) != nil {
		return []plugin.Item{{Reason: "malformed_item"}}, nil
	}
	if m.Chat.Type != "private" {
		return nil, nil
	}
	chat := strconv.FormatInt(m.Chat.ID, 10)
	sender := chat
	if m.From != nil && m.From.ID != 0 {
		sender = strconv.FormatInt(m.From.ID, 10)
	}
	e := ingress.Event{Platform: "telegram", Sender: sender, ConversationKey: chat, Direction: "inbound", Type: ingress.TypeMessage, SourceAt: time.Unix(m.Date, 0).UTC(), ReceivedAt: received, Acquisition: "webhook", Purpose: "live"}
	item := plugin.Item{}
	if m.MessageID > 0 {
		e.MessageID = strconv.FormatInt(m.MessageID, 10)
		// message_id is unique only within a chat.
		e.SourceIdentity = "message:" + chat + ":" + e.MessageID
	}
	if m.Reply != nil && m.Reply.MessageID > 0 {
		e.ReplyTo = strconv.FormatInt(m.Reply.MessageID, 10)
	}
	for _, media := range mediaKinds {
		if _, ok := fields[media.field]; ok {
			e.Attachments = append(e.Attachments, ingress.Attachment{Type: media.kind})
		}
	}
	switch {
	case len(e.Attachments) > 0:
		// Captions are dropped with the media; only the placeholder kinds are kept.
		e.Type = ingress.TypeUnsupportedMedia
	case m.Text != nil:
		e.Text = *m.Text
		if len([]rune(e.Text)) > 4000 {
			item.Reason = "text_too_long"
		}
	default:
		item.Reason = "unsupported_event"
	}
	if e.MessageID == "" || m.Date <= 0 || m.Chat.ID == 0 || (e.Type == ingress.TypeMessage && e.Text == "" && item.Reason == "") {
		item.Reason = "invalid_message"
	}
	item.Event = e
	return []plugin.Item{item}, nil
}

var _ plugin.Plugin = TelegramPlugin{}
var _ plugin.Webhook = TelegramPlugin{}
