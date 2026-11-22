package mockplatform

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

var testSecret = []byte("fixture-app-secret-0123456789")

const testToken = "4242:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

type delivery struct {
	path    string
	headers http.Header
	body    []byte
}

// connectorStub records webhook deliveries and answers with status.
type connectorStub struct {
	mu     sync.Mutex
	got    []delivery
	status int
}

func (c *connectorStub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	c.mu.Lock()
	c.got = append(c.got, delivery{r.URL.Path, r.Header.Clone(), body})
	c.mu.Unlock()
	w.WriteHeader(c.status)
}

func (c *connectorStub) deliveries() []delivery {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]delivery(nil), c.got...)
}

func setup(t *testing.T, echo bool) (*httptest.Server, *connectorStub) {
	t.Helper()
	stub := &connectorStub{status: http.StatusOK}
	connector := httptest.NewServer(stub)
	t.Cleanup(connector.Close)
	mock := httptest.NewServer(New(Config{AppID: "mock-app", AppSecret: testSecret, ConnectorURL: connector.URL, Pages: 2, Echo: echo}).Handler())
	t.Cleanup(mock.Close)
	return mock, stub
}

func call(t *testing.T, method, target, token string, body any, out any) int {
	t.Helper()
	var reader io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reader = bytes.NewReader(data)
	}
	req, _ := http.NewRequest(method, target, reader)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if out != nil {
		_ = json.NewDecoder(res.Body).Decode(out)
	}
	return res.StatusCode
}

type graphErr struct {
	Error struct {
		Code    int `json:"code"`
		Subcode int `json:"error_subcode"`
	} `json:"error"`
}

// connectPage runs the connector's OAuth exchange and subscribes the first Page, returning its ID and token.
func connectPage(t *testing.T, mock string) (string, string) {
	t.Helper()
	redirect := "http://localhost:3000/channels/facebook/callback"
	res, err := http.Get(mock + "/v21.0/dialog/oauth?" + url.Values{"client_id": {"mock-app"}, "redirect_uri": {redirect}, "state": {"st"}}.Encode())
	if err != nil || res.StatusCode != http.StatusOK {
		t.Fatalf("dialog: %v %v", err, res)
	}
	page, _ := io.ReadAll(res.Body)
	res.Body.Close()
	start := strings.Index(string(page), "code=")
	code := strings.SplitN(string(page[start+5:]), "&", 2)[0]
	var short, long struct {
		Token string `json:"access_token"`
	}
	exchange := url.Values{"client_id": {"mock-app"}, "client_secret": {string(testSecret)}, "redirect_uri": {redirect}, "code": {code}}
	if status := call(t, "GET", mock+"/v21.0/oauth/access_token?"+exchange.Encode(), "", nil, &short); status != 200 || short.Token == "" {
		t.Fatalf("code exchange: %d", status)
	}
	var reused graphErr
	if status := call(t, "GET", mock+"/v21.0/oauth/access_token?"+exchange.Encode(), "", nil, &reused); status != 400 || reused.Error.Code != 100 {
		t.Fatalf("reused code: %d %+v", status, reused)
	}
	upgrade := url.Values{"grant_type": {"fb_exchange_token"}, "client_id": {"mock-app"}, "client_secret": {string(testSecret)}, "fb_exchange_token": {short.Token}}
	if status := call(t, "GET", mock+"/v21.0/oauth/access_token?"+upgrade.Encode(), "", nil, &long); status != 200 || long.Token == "" {
		t.Fatalf("token upgrade: %d", status)
	}
	var pages struct {
		Data []map[string]string `json:"data"`
	}
	if status := call(t, "GET", mock+"/v21.0/me/accounts?fields=id,name,access_token&limit=100", long.Token, nil, &pages); status != 200 || len(pages.Data) != 2 || pages.Data[0]["access_token"] == "" {
		t.Fatalf("accounts: %d %+v", status, len(pages.Data))
	}
	id, token := pages.Data[0]["id"], pages.Data[0]["access_token"]
	if status := call(t, "POST", mock+"/v21.0/"+id+"/subscribed_apps", token, nil, nil); status != 200 {
		t.Fatalf("subscribe: %d", status)
	}
	return id, token
}

func TestFacebookOAuthSubscribeAndSend(t *testing.T) {
	mock, stub := setup(t, true)
	if status := call(t, "GET", mock.URL+"/v21.0/dialog/oauth?client_id=other&redirect_uri=http://x/cb", "", nil, nil); status != 400 {
		t.Fatalf("wrong client_id: %d", status)
	}
	pageID, token := connectPage(t, mock.URL)

	var sent struct {
		MessageID string `json:"message_id"`
	}
	body := map[string]any{"recipient": map[string]string{"id": "7001"}, "messaging_type": "RESPONSE", "message": map[string]string{"text": "Dạ shop còn hàng"}}
	if status := call(t, "POST", mock.URL+"/v21.0/"+pageID+"/messages", token, body, &sent); status != 200 || sent.MessageID == "" {
		t.Fatalf("send: %d %+v", status, sent)
	}
	var rejected graphErr
	if status := call(t, "POST", mock.URL+"/v21.0/"+pageID+"/messages", "wrong", body, &rejected); status != 400 || rejected.Error.Code != 190 {
		t.Fatalf("bad token: %d %+v", status, rejected)
	}
	// The echo is delivered asynchronously, signed like any Meta webhook.
	deadline := time.Now().Add(2 * time.Second)
	for len(stub.deliveries()) == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	got := stub.deliveries()
	if len(got) != 1 || got[0].path != "/webhooks/facebook" || got[0].headers.Get("X-Hub-Signature-256") != SignFacebook(testSecret, got[0].body) || !bytes.Contains(got[0].body, []byte(`"is_echo":true`)) {
		t.Fatalf("echo delivery: %+v", got)
	}
	var log struct{ Messages []Message }
	call(t, "GET", mock.URL+"/mock/messages?platform=facebook&customer=7001", "", nil, &log)
	if len(log.Messages) != 1 || log.Messages[0].Direction != "outbound" || log.Messages[0].Text != "Dạ shop còn hàng" {
		t.Fatalf("outbox: %+v", log.Messages)
	}
}

func TestFacebookCustomerWebhookIsSignedAndRequiresSubscription(t *testing.T) {
	mock, stub := setup(t, false)
	in := map[string]string{"account_id": PageID(1), "customer": "7002", "text": "Còn size M không?"}
	if status := call(t, "POST", mock.URL+"/mock/customers/facebook", "", in, nil); status != http.StatusConflict {
		t.Fatalf("unsubscribed page: %d", status)
	}
	connectPage(t, mock.URL)
	var out struct {
		Status int `json:"webhook_status"`
	}
	if status := call(t, "POST", mock.URL+"/mock/customers/facebook", "", in, &out); status != 200 || out.Status != 200 {
		t.Fatalf("customer: %d %+v", status, out)
	}
	got := stub.deliveries()
	if len(got) != 1 || got[0].headers.Get("X-Hub-Signature-256") != SignFacebook(testSecret, got[0].body) {
		t.Fatalf("delivery: %+v", got)
	}
	var batch struct {
		Object string `json:"object"`
		Entry  []struct {
			ID        string `json:"id"`
			Messaging []struct {
				Sender, Recipient struct{ ID string }
				Timestamp         int64
				Message           struct{ MID, Text string }
			} `json:"messaging"`
		} `json:"entry"`
	}
	if json.Unmarshal(got[0].body, &batch) != nil || batch.Object != "page" || batch.Entry[0].ID != PageID(1) {
		t.Fatalf("batch: %s", got[0].body)
	}
	m := batch.Entry[0].Messaging[0]
	if m.Sender.ID != "7002" || m.Recipient.ID != PageID(1) || m.Message.Text != "Còn size M không?" || m.Message.MID == "" || m.Timestamp <= 0 {
		t.Fatalf("message: %+v", m)
	}
}

func TestTelegramBotLifecycleAndCustomerUpdates(t *testing.T) {
	mock, stub := setup(t, false)
	var me struct {
		OK     bool `json:"ok"`
		Result struct {
			ID    int64 `json:"id"`
			IsBot bool  `json:"is_bot"`
		} `json:"result"`
	}
	if status := call(t, "POST", mock.URL+"/bot"+testToken+"/getMe", "", map[string]any{}, &me); status != 200 || !me.OK || me.Result.ID != 4242 || !me.Result.IsBot {
		t.Fatalf("getMe: %d %+v", status, me)
	}
	if status := call(t, "POST", mock.URL+"/botmalformed/getMe", "", map[string]any{}, nil); status != 404 {
		t.Fatalf("malformed token: %d", status)
	}
	in := map[string]string{"account_id": "4242", "customer": "9001", "text": "hello"}
	if status := call(t, "POST", mock.URL+"/mock/customers/telegram", "", in, nil); status != http.StatusConflict {
		t.Fatalf("no webhook: %d", status)
	}
	webhook := stubURL(t, stub) + "/webhooks/telegram/11111111-1111-1111-1111-111111111111"
	if status := call(t, "POST", mock.URL+"/bot"+testToken+"/setWebhook", "", map[string]any{"url": webhook, "secret_token": "fixture-secret", "allowed_updates": []string{"message"}}, nil); status != 200 {
		t.Fatalf("setWebhook: %d", status)
	}
	if status := call(t, "POST", mock.URL+"/mock/customers/telegram", "", in, nil); status != 200 {
		t.Fatalf("customer: %d", status)
	}
	photo := map[string]string{"account_id": "4242", "customer": "9001", "media": MediaImage}
	if status := call(t, "POST", mock.URL+"/mock/customers/telegram", "", photo, nil); status != 200 {
		t.Fatalf("photo: %d", status)
	}
	got := stub.deliveries()
	if len(got) != 2 || got[0].headers.Get(TelegramSecretHeader) != "fixture-secret" {
		t.Fatalf("deliveries: %+v", got)
	}
	var update struct {
		UpdateID int64 `json:"update_id"`
		Message  struct {
			MessageID int64 `json:"message_id"`
			Date      int64
			Chat      struct {
				ID   int64
				Type string
			}
			Text  *string
			Photo []any
		}
	}
	if json.Unmarshal(got[0].body, &update) != nil || update.UpdateID == 0 || update.Message.Chat.ID != 9001 || update.Message.Chat.Type != "private" || update.Message.Text == nil || *update.Message.Text != "hello" {
		t.Fatalf("text update: %s", got[0].body)
	}
	if json.Unmarshal(got[1].body, &update) != nil || len(update.Message.Photo) == 0 {
		t.Fatalf("photo update: %s", got[1].body)
	}

	var sent struct {
		OK bool `json:"ok"`
	}
	if status := call(t, "POST", mock.URL+"/bot"+testToken+"/sendMessage", "", map[string]string{"chat_id": "9001", "text": "Chào bạn"}, &sent); status != 200 || !sent.OK {
		t.Fatalf("sendMessage: %d", status)
	}
	if status := call(t, "POST", mock.URL+"/bot"+testToken+"/deleteWebhook", "", map[string]any{"drop_pending_updates": true}, nil); status != 200 {
		t.Fatalf("deleteWebhook: %d", status)
	}
	if status := call(t, "POST", mock.URL+"/mock/customers/telegram", "", in, nil); status != http.StatusConflict {
		t.Fatalf("after deleteWebhook: %d", status)
	}
}

func stubURL(t *testing.T, stub *connectorStub) string {
	t.Helper()
	s := httptest.NewServer(stub)
	t.Cleanup(s.Close)
	return s.URL
}

func TestFaultInjection(t *testing.T) {
	mock, _ := setup(t, false)
	if status := call(t, "PUT", mock.URL+"/mock/faults", "", map[string]any{"error_rate": 2}, nil); status != 400 {
		t.Fatalf("invalid rate: %d", status)
	}
	if status := call(t, "PUT", mock.URL+"/mock/faults", "", map[string]any{"error_rate": 1, "error_kind": "nope"}, nil); status != 400 {
		t.Fatalf("invalid kind: %d", status)
	}
	send := func() (int, map[string]any) {
		var out map[string]any
		status := call(t, "POST", mock.URL+"/bot"+testToken+"/sendMessage", "", map[string]string{"chat_id": "9001", "text": "x"}, &out)
		return status, out
	}
	cases := map[string]int{FaultBlocked: 403, FaultAuthExpired: 401, FaultRateLimit: 429, FaultChatNotFound: 400, FaultServerError: 500}
	for kind, want := range cases {
		call(t, "PUT", mock.URL+"/mock/faults", "", map[string]any{"error_rate": 1, "error_kind": kind}, nil)
		if status, out := send(); status != want || out["ok"] != false {
			t.Fatalf("%s: got %d %v", kind, status, out)
		}
	}
	pageID, token := connectPage(t, mock.URL)
	graph := map[string]int{FaultBlocked: 551, FaultAuthExpired: 190, FaultRateLimit: 4, FaultWindow24h: 10}
	for kind, code := range graph {
		call(t, "PUT", mock.URL+"/mock/faults", "", map[string]any{"error_rate": 1, "error_kind": kind}, nil)
		var e graphErr
		body := map[string]any{"recipient": map[string]string{"id": "7001"}, "message": map[string]string{"text": "x"}}
		if status := call(t, "POST", mock.URL+"/v21.0/"+pageID+"/messages", token, body, &e); status != 400 || e.Error.Code != code {
			t.Fatalf("%s: got %d %+v", kind, status, e)
		}
	}
	// Account latency applies to one bot only.
	call(t, "PUT", mock.URL+"/mock/faults", "", map[string]any{"account_latency_ms": map[string]int{"4242": 200}}, nil)
	started := time.Now()
	if status, _ := send(); status != 200 || time.Since(started) < 200*time.Millisecond {
		t.Fatalf("slow bot: %d after %s", status, time.Since(started))
	}
	started = time.Now()
	other := "5151:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	if status := call(t, "POST", mock.URL+"/bot"+other+"/sendMessage", "", map[string]string{"chat_id": "9001", "text": "x"}, nil); status != 200 || time.Since(started) > 150*time.Millisecond {
		t.Fatalf("other bot: %d after %s", status, time.Since(started))
	}
}

func TestPageTokensAreStableAcrossRestarts(t *testing.T) {
	a := New(Config{AppID: "mock-app", AppSecret: testSecret, Pages: 1})
	b := New(Config{AppID: "mock-app", AppSecret: testSecret, Pages: 1})
	if a.pages[0].token != b.pages[0].token || strings.Contains(a.pages[0].token, string(testSecret)) {
		t.Fatal("page token must be deterministic and must not reveal the app secret")
	}
}
