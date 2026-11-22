package mockplatform

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

// Wire builders for provider webhooks. They follow the public Messenger and Bot API formats independently of the
// connector's parser, so a mismatch shows up as a rejected or quarantined delivery.

// MediaImage asks a builder for a photo instead of text; the connector turns it into a placeholder.
const MediaImage = "image"

type fbAttachment struct {
	Type    string            `json:"type"`
	Payload map[string]string `json:"payload"`
}

type fbMessage struct {
	MID         string         `json:"mid"`
	Text        string         `json:"text,omitempty"`
	Echo        bool           `json:"is_echo,omitempty"`
	AppID       int64          `json:"app_id,omitempty"`
	Attachments []fbAttachment `json:"attachments,omitempty"`
}

type fbMessaging struct {
	Sender    map[string]string `json:"sender"`
	Recipient map[string]string `json:"recipient"`
	Timestamp int64             `json:"timestamp"`
	Message   fbMessage         `json:"message"`
}

func facebookBatch(pageID string, items ...fbMessaging) []byte {
	body, _ := json.Marshal(map[string]any{"object": "page", "entry": []map[string]any{{"id": pageID, "time": time.Now().UnixMilli(), "messaging": items}}})
	return body
}

// FacebookMessage builds a one-message Page webhook batch from a customer (PSID) to a Page.
func FacebookMessage(pageID, psid, mid, text, media string, at time.Time) []byte {
	m := fbMessage{MID: mid, Text: text}
	if media == MediaImage {
		m = fbMessage{MID: mid, Attachments: []fbAttachment{{Type: "image", Payload: map[string]string{"url": "https://example.invalid/mock.jpg"}}}}
	}
	return facebookBatch(pageID, fbMessaging{Sender: map[string]string{"id": psid}, Recipient: map[string]string{"id": pageID}, Timestamp: at.UnixMilli(), Message: m})
}

// FacebookEcho builds the message_echoes webhook Meta sends after the Page itself sent a message.
func FacebookEcho(pageID, psid, mid, text string, at time.Time) []byte {
	return facebookBatch(pageID, fbMessaging{Sender: map[string]string{"id": pageID}, Recipient: map[string]string{"id": psid}, Timestamp: at.UnixMilli(), Message: fbMessage{MID: mid, Text: text, Echo: true}})
}

// SignFacebook returns the X-Hub-Signature-256 header value for a webhook body.
func SignFacebook(secret, body []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// TelegramUpdate builds a private-chat message update; the chat ID doubles as the user ID, as in 1-1 chats.
func TelegramUpdate(updateID, chatID, messageID int64, text, media string, at time.Time) []byte {
	user := map[string]any{"id": chatID, "is_bot": false, "first_name": "Mock customer"}
	message := map[string]any{"message_id": messageID, "date": at.Unix(), "from": user, "chat": map[string]any{"id": chatID, "type": "private", "first_name": "Mock customer"}}
	if media == MediaImage {
		message["photo"] = []map[string]any{{"file_id": "mock-file", "file_unique_id": "mock-unique", "width": 90, "height": 90}}
	} else {
		message["text"] = text
	}
	body, _ := json.Marshal(map[string]any{"update_id": updateID, "message": message})
	return body
}

// TelegramSecretHeader carries the webhook secret_token registered through setWebhook.
const TelegramSecretHeader = "X-Telegram-Bot-Api-Secret-Token"
