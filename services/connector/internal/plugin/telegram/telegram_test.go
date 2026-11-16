package telegram

import (
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/ingress"
)

func TestParsePrivateTextAndMedia(t *testing.T) {
	now := time.Now().UTC()
	items, err := TelegramPlugin{}.Parse([]byte(`{"update_id":10,"message":{"message_id":7,"date":1700000000,"from":{"id":42,"is_bot":false},"chat":{"id":42,"type":"private"},"text":"Xin chào","reply_to_message":{"message_id":5}}}`), now)
	if err != nil || len(items) != 1 || items[0].Reason != "" {
		t.Fatal("private text", err, items)
	}
	e := items[0].Event
	if e.Platform != "telegram" || e.Type != ingress.TypeMessage || e.Text != "Xin chào" || e.ConversationKey != "42" || e.Sender != "42" || e.MessageID != "7" || e.SourceIdentity != "message:42:7" || e.ReplyTo != "5" || e.Direction != "inbound" || !e.SourceAt.Equal(time.Unix(1700000000, 0)) {
		t.Fatal("text mapping", e)
	}
	for body, kind := range map[string]string{
		`"photo":[{"file_id":"f"}],"caption":"look"`: "image",
		`"voice":{"file_id":"f"}`:                    "audio",
		`"audio":{"file_id":"f"}`:                    "audio",
		`"video":{"file_id":"f"}`:                    "video",
		`"document":{"file_id":"f"}`:                 "file",
		`"sticker":{"file_id":"f"}`:                  "sticker",
	} {
		items, err := TelegramPlugin{}.Parse([]byte(`{"update_id":11,"message":{"message_id":8,"date":1700000000,"from":{"id":42},"chat":{"id":42,"type":"private"},`+body+`}}`), now)
		if err != nil || len(items) != 1 || items[0].Reason != "" {
			t.Fatal("media", body, err, items)
		}
		e := items[0].Event
		if e.Type != ingress.TypeUnsupportedMedia || e.Text != "" || len(e.Attachments) != 1 || e.Attachments[0].Type != kind || !ingress.MediaKinds[kind] {
			t.Fatal("media mapping", body, e)
		}
	}
	items, _ = TelegramPlugin{}.Parse([]byte(`{"update_id":12,"message":{"message_id":9,"date":1700000000,"from":{"id":42},"chat":{"id":42,"type":"private"},"location":{"latitude":1,"longitude":2}}}`), now)
	if len(items) != 1 || items[0].Reason != "unsupported_event" {
		t.Fatal("unsupported private content must be quarantined", items)
	}
}

func TestParseDropsGroupsChannelsAndNonMessages(t *testing.T) {
	for _, body := range []string{
		`{"update_id":1,"message":{"message_id":1,"date":1700000000,"from":{"id":1},"chat":{"id":-100,"type":"group"},"text":"hi"}}`,
		`{"update_id":2,"message":{"message_id":1,"date":1700000000,"from":{"id":1},"chat":{"id":-100,"type":"supergroup"},"text":"hi"}}`,
		`{"update_id":3,"message":{"message_id":1,"date":1700000000,"chat":{"id":-100,"type":"channel"},"text":"hi"}}`,
		`{"update_id":4,"channel_post":{"message_id":1,"date":1700000000,"chat":{"id":-100,"type":"channel"},"text":"hi"}}`,
		`{"update_id":5,"edited_message":{"message_id":1,"date":1700000000,"chat":{"id":1,"type":"private"},"text":"hi"}}`,
	} {
		items, err := TelegramPlugin{}.Parse([]byte(body), time.Now())
		if err != nil || len(items) != 0 {
			t.Fatal("must drop before outbox", body, items)
		}
	}
	for _, body := range []string{`not json`, `{"message":{}}`} {
		if _, err := (TelegramPlugin{}).Parse([]byte(body), time.Now()); err == nil {
			t.Fatal("accepted invalid update", body)
		}
	}
}

func TestVerifySecretHeader(t *testing.T) {
	p := TelegramPlugin{}
	if p.Verify(nil, "secret-value", []byte("secret-value")) != nil {
		t.Fatal("rejected matching secret")
	}
	for _, header := range []string{"", "secret-valuX", "secret-value-longer"} {
		if p.Verify(nil, header, []byte("secret-value")) == nil {
			t.Fatal("accepted wrong secret", header)
		}
	}
	if p.Verify(nil, "", nil) == nil {
		t.Fatal("accepted missing stored secret")
	}
}
