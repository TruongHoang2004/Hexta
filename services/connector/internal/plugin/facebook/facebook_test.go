package facebook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"testing"
	"time"
)

func TestBatchAndAuthentication(t *testing.T) {
	p := FacebookPlugin{}
	body := []byte(`{"object":"page","entry":[{"id":"p1","messaging":[{"sender":{"id":"u"},"recipient":{"id":"p1"},"timestamp":1700000000000,"message":{"mid":"m1","text":"hello","reply_to":{"mid":"old"}}},{"sender":7},{"sender":{"id":"p1"},"recipient":{"id":"u"},"timestamp":1700000000000,"message":{"mid":"m2","text":"echo","is_echo":true}}]},{"id":"p2","messaging":[{"sender":{"id":"v"},"recipient":{"id":"p2"},"timestamp":1700000000000,"message":{"mid":"m3","text":"other"}}]}]}`)
	secret := []byte("fixture-secret")
	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if p.Verify(body, sig, secret) != nil {
		t.Fatal("valid signature")
	}
	if p.Verify(append(body, ' '), sig, secret) == nil || p.Verify(body, "", secret) == nil {
		t.Fatal("forged signature")
	}
	q := url.Values{"hub.mode": {"subscribe"}, "hub.verify_token": {"fixture"}, "hub.challenge": {"123"}}
	if got, err := p.Challenge(q, "fixture"); err != nil || got != "123" {
		t.Fatal("challenge")
	}
	if _, err := p.Challenge(q, "wrong"); err == nil {
		t.Fatal("token")
	}
	items, err := p.Parse(body, time.Now())
	if err != nil || len(items) != 4 {
		t.Fatal(err, len(items))
	}
	if items[0].Event.MessageID != "m1" || items[0].Event.ReplyTo != "old" || items[1].Reason != "malformed_item" || items[2].Event.Type != "echo" || items[3].AccountID != "p2" {
		t.Fatal("batch mapping")
	}
}

func TestMediaBecomesPlaceholderWithoutBinaryReferences(t *testing.T) {
	body := []byte(`{"object":"page","entry":[{"id":"p1","messaging":[
{"sender":{"id":"u"},"recipient":{"id":"p1"},"timestamp":1700000000000,"message":{"mid":"m1","attachments":[{"type":"image","payload":{"url":"https://cdn.example/private.jpg"}}]}},
{"sender":{"id":"u"},"recipient":{"id":"p1"},"timestamp":1700000000000,"message":{"mid":"m2","sticker_id":369239263222822,"attachments":[{"type":"image","payload":{"url":"https://cdn.example/s.png","sticker_id":369239263222822}}]}},
{"sender":{"id":"u"},"recipient":{"id":"p1"},"timestamp":1700000000000,"message":{"mid":"m3","attachments":[{"type":"audio"},{"type":"file"}]}},
{"sender":{"id":"u"},"recipient":{"id":"p1"},"timestamp":1700000000000,"message":{"mid":"m4","attachments":[{"type":"template"}]}}]}]}`)
	items, err := FacebookPlugin{}.Parse(body, time.Now())
	if err != nil || len(items) != 4 {
		t.Fatal(err, len(items))
	}
	for i, want := range [][]string{{"image"}, {"sticker"}, {"audio", "file"}} {
		e := items[i].Event
		if items[i].Reason != "" || e.Type != "unsupported_media" || e.Text != "" || len(e.Attachments) != len(want) {
			t.Fatal("media mapping", i, items[i].Reason, e.Type)
		}
		for j, kind := range want {
			if e.Attachments[j].Type != kind || e.Attachments[j].Reference != "" {
				t.Fatal("attachment kind", i, e.Attachments[j])
			}
		}
	}
	if items[3].Reason != "unsupported_attachment" {
		t.Fatal("unknown attachment kind must stay quarantined")
	}
}
