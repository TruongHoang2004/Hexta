package facebook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/TruongHoang2004/Hexta/packages/shared/ingress"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin"
	"net/url"
	"strings"
	"time"
)

type FacebookPlugin struct{ plugin.BasePlugin }

func (FacebookPlugin) Platform() string                  { return "facebook" }
func (FacebookPlugin) Capabilities() plugin.Capabilities { return plugin.Capabilities{Webhook: true} }
func (FacebookPlugin) Verify(body []byte, signature string, secret []byte) error {
	if len(secret) == 0 || !strings.HasPrefix(signature, "sha256=") {
		return errors.New("invalid signature")
	}
	got, err := hex.DecodeString(strings.TrimPrefix(signature, "sha256="))
	if err != nil {
		return errors.New("invalid signature")
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	if !hmac.Equal(got, mac.Sum(nil)) {
		return errors.New("invalid signature")
	}
	return nil
}
func (FacebookPlugin) Challenge(q url.Values, token string) (string, error) {
	if token == "" || q.Get("hub.mode") != "subscribe" || !hmac.Equal([]byte(q.Get("hub.verify_token")), []byte(token)) || q.Get("hub.challenge") == "" {
		return "", errors.New("invalid challenge")
	}
	return q.Get("hub.challenge"), nil
}
func (FacebookPlugin) Parse(body []byte, received time.Time) ([]plugin.Item, error) {
	var batch struct {
		Object string `json:"object"`
		Entry  []struct {
			ID        string            `json:"id"`
			Messaging []json.RawMessage `json:"messaging"`
		} `json:"entry"`
	}
	if json.Unmarshal(body, &batch) != nil || batch.Object != "page" || len(batch.Entry) == 0 {
		return nil, errors.New("invalid batch")
	}
	result := []plugin.Item{}
	for _, entry := range batch.Entry {
		if len(result) >= 1000 {
			return nil, errors.New("batch limit")
		}
		if len(entry.Messaging) == 0 {
			result = append(result, plugin.Item{AccountID: entry.ID, Reason: "unsupported_entry"})
		}
		for _, raw := range entry.Messaging {
			if len(result) >= 1000 {
				return nil, errors.New("batch limit")
			}
			var m struct {
				Sender struct {
					ID string `json:"id"`
				} `json:"sender"`
				Recipient struct {
					ID string `json:"id"`
				} `json:"recipient"`
				Timestamp int64 `json:"timestamp"`
				Message   *struct {
					MID   string `json:"mid"`
					Text  string `json:"text"`
					Echo  bool   `json:"is_echo"`
					Reply struct {
						MID string `json:"mid"`
					} `json:"reply_to"`
					Attachments []struct {
						Type string `json:"type"`
					} `json:"attachments"`
				} `json:"message"`
				Reaction json.RawMessage `json:"reaction"`
				Read     json.RawMessage `json:"read"`
				Delivery json.RawMessage `json:"delivery"`
			}
			item := plugin.Item{AccountID: entry.ID}
			if json.Unmarshal(raw, &m) != nil {
				item.Reason = "malformed_item"
				result = append(result, item)
				continue
			}
			e := ingress.Event{Platform: "facebook", AccountID: entry.ID, Sender: m.Sender.ID, Recipient: m.Recipient.ID, ConversationKey: m.Sender.ID, Direction: "inbound", Type: "message", SourceAt: time.UnixMilli(m.Timestamp), ReceivedAt: received, Acquisition: "webhook", Purpose: "live"}
			if m.Message == nil {
				switch {
				case len(m.Reaction) > 0:
					item.Reason = "unsupported_reaction"
				case len(m.Read) > 0:
					item.Reason = "unsupported_read"
				case len(m.Delivery) > 0:
					item.Reason = "unsupported_delivery"
				default:
					item.Reason = "unsupported_event"
				}
			} else {
				e.MessageID = m.Message.MID
				e.SourceIdentity = "message:" + m.Message.MID
				e.Text = m.Message.Text
				e.ReplyTo = m.Message.Reply.MID
				if m.Message.Echo {
					e.Direction = "outbound"
					e.Type = "echo"
					e.ConversationKey = m.Recipient.ID
				}
				for _, a := range m.Message.Attachments {
					e.Attachments = append(e.Attachments, ingress.Attachment{Type: a.Type})
				}
				if len(e.Attachments) > 0 {
					item.Reason = "unsupported_attachment"
				}
				if e.MessageID == "" || e.Text == "" || len([]rune(e.Text)) > 4000 || m.Timestamp <= 0 || e.Sender == "" || len(e.Sender) > 255 || e.Recipient == "" || len(e.Recipient) > 255 || entry.ID == "" || (!m.Message.Echo && e.Recipient != entry.ID) || (m.Message.Echo && e.Sender != entry.ID) {
					item.Reason = "invalid_message"
				}
			}
			item.Event = e
			result = append(result, item)
		}
	}
	return result, nil
}

var _ plugin.Plugin = FacebookPlugin{}
var _ plugin.Webhook = FacebookPlugin{}
