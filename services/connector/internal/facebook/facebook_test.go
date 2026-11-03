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
