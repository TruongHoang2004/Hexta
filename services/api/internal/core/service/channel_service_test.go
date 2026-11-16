package service

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/ingress"
	"github.com/TruongHoang2004/Hexta/packages/shared/outbound"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/messagingmodel"
	"github.com/google/uuid"
)

type fakeConnector struct {
	pages     []types.ChannelPage
	upserts   []types.ConnectorConnection
	fail      error
	bots      map[string]types.TelegramBot
	verifies  int
	verifyErr error
}

func (f *fakeConnector) VerifyTelegramBot(_ context.Context, token string) (*types.TelegramBot, error) {
	f.verifies++
	if f.verifyErr != nil {
		return nil, f.verifyErr
	}
	bot, ok := f.bots[token]
	if !ok {
		return nil, types.ErrConnectorCredential
	}
	return &bot, nil
}

func (f *fakeConnector) Exchange(context.Context, string, string) (*types.ConnectorExchange, error) {
	return &types.ConnectorExchange{OAuthSessionID: uuid.NewString(), ExpiresAt: time.Now().Add(10 * time.Minute), Pages: f.pages}, nil
}
func (f *fakeConnector) Upsert(_ context.Context, c types.ConnectorConnection) (*types.ConnectorUpsertResult, error) {
	if f.fail != nil {
		return nil, f.fail
	}
	f.upserts = append(f.upserts, c)
	return &types.ConnectorUpsertResult{ID: c.ID, Version: c.Version, Enabled: c.Enabled, Subscribed: c.Enabled}, nil
}

const channelRedirect = "http://localhost:3000/channels/facebook/callback"

func channelServiceFixture(t *testing.T) (salesFixture, *ChannelService, *fakeConnector) {
	t.Helper()
	f := salesTestFixture(t)
	connector := &fakeConnector{pages: []types.ChannelPage{{PageID: "page-a", PageName: "Shop A"}, {PageID: "page-b", PageName: "Shop B"}}}
	s := NewChannelService(repository.NewChannelRepository(f.db), connector, ChannelSettings{AppID: "app", AllowedOrigins: []string{"http://localhost:3000"}})
	return f, s, connector
}

func startState(t *testing.T, s *ChannelService, tenant, actor string) string {
	t.Helper()
	start, err := s.StartFacebookOAuth(context.Background(), tenant, actor, channelRedirect)
	if err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse(start.AuthURL)
	if u.Query().Get("client_id") != "app" || u.Query().Get("redirect_uri") != channelRedirect || !strings.Contains(u.Query().Get("scope"), "pages_messaging") {
		t.Fatal("auth URL", start.AuthURL)
	}
	return u.Query().Get("state")
}

func TestChannelOAuthStateOwnerOnlySingleUse(t *testing.T) {
	f, s, _ := channelServiceFixture(t)
	ctx := context.Background()
	_, err := s.StartFacebookOAuth(ctx, f.tenant, f.sale, channelRedirect)
	requireSalesStatus(t, err, 403)
	for _, bad := range []string{"https://evil.example/channels/facebook/callback", "http://localhost:3000/other", channelRedirect + "?x=1"} {
		_, err = s.StartFacebookOAuth(ctx, f.tenant, f.owner, bad)
		requireSalesStatus(t, err, 400)
	}
	state := startState(t, s, f.tenant, f.owner)
	var stored messagingmodel.OAuthState
	f.db.Where("tenant_id = ?", f.tenant).First(&stored)
	if stored.StateHash == state || stored.StateHash != channelStateHash(state) {
		t.Fatal("raw state persisted")
	}
	_, err = s.CompleteFacebookOAuth(ctx, f.sale, "code", state, "")
	requireSalesStatus(t, err, 403)
	// The other actor's attempt consumed the state; a stolen state is burned, never reusable.
	_, err = s.CompleteFacebookOAuth(ctx, f.owner, "code", state, "")
	requireSalesStatus(t, err, 403)
	state = startState(t, s, f.tenant, f.owner)
	pages, err := s.CompleteFacebookOAuth(ctx, f.owner, "code", state, "")
	if err != nil || pages.TenantID != f.tenant || len(pages.Pages) != 2 || pages.Pages[0].Availability != types.PageAvailable {
		t.Fatal("callback", err)
	}
	_, err = s.CompleteFacebookOAuth(ctx, f.owner, "code", state, "")
	requireSalesStatus(t, err, 403)
	f.db.Model(&messagingmodel.OAuthState{}).Where("tenant_id = ?", f.tenant).Update("expires_at", time.Now().Add(-time.Minute))
	state = startState(t, s, f.tenant, f.owner)
	f.db.Model(&messagingmodel.OAuthState{}).Where("state_hash = ?", channelStateHash(state)).Update("expires_at", time.Now().Add(-time.Second))
	_, err = s.CompleteFacebookOAuth(ctx, f.owner, "code", state, "")
	requireSalesStatus(t, err, 403)
	state = startState(t, s, f.tenant, f.owner)
	_, err = s.CompleteFacebookOAuth(ctx, f.owner, "", state, "access_denied")
	requireSalesStatus(t, err, 400)
}

func connectPages(t *testing.T, s *ChannelService, tenant, owner string, pages ...string) ([]types.ChannelConnection, *types.ChannelOAuthPages) {
	t.Helper()
	state := startState(t, s, tenant, owner)
	session, err := s.CompleteFacebookOAuth(context.Background(), owner, "code", state, "")
	if err != nil {
		t.Fatal(err)
	}
	result, appErr := s.ConnectFacebookPages(context.Background(), tenant, owner, session.OAuthSessionID, pages)
	if appErr != nil {
		t.Fatal(appErr)
	}
	return result, session
}

func TestChannelOnePageOneTenantAndLifecycle(t *testing.T) {
	f, s, connector := channelServiceFixture(t)
	ctx := context.Background()
	connected, _ := connectPages(t, s, f.tenant, f.owner, "page-a")
	if len(connected) != 1 || connected[0].Status != types.ChannelStatusActive || connector.upserts[0].OAuthSessionID == "" || connector.upserts[0].AppID != "app" {
		t.Fatal("connect", connected)
	}
	// Another tenant sees the Page as taken and cannot bind it.
	state := startState(t, s, f.other, f.owner)
	session, err := s.CompleteFacebookOAuth(ctx, f.owner, "code", state, "")
	if err != nil || session.Pages[0].Availability != types.PageOwnedByOtherTenant || session.Pages[1].Availability != types.PageAvailable {
		t.Fatal("availability", err)
	}
	_, err = s.ConnectFacebookPages(ctx, f.other, f.owner, session.OAuthSessionID, []string{"page-a"})
	requireSalesStatus(t, err, 409)
	_, err = s.ConnectFacebookPages(ctx, f.other, f.owner, session.OAuthSessionID, []string{"page-z"})
	requireSalesStatus(t, err, 400)
	// Sessions are bound to the tenant and owner that completed OAuth.
	_, err = s.ConnectFacebookPages(ctx, f.tenant, f.owner, session.OAuthSessionID, []string{"page-b"})
	requireSalesStatus(t, err, 422)

	list, err := s.ListFacebookConnections(ctx, f.tenant, f.sale)
	if err != nil || len(list) != 1 {
		t.Fatal("member list", err)
	}
	id := connected[0].ID
	_, err = s.UpdateFacebookConnection(ctx, f.tenant, f.sale, id, "pause")
	requireSalesStatus(t, err, 403)
	_, err = s.UpdateFacebookConnection(ctx, f.other, f.owner, id, "pause")
	requireSalesStatus(t, err, 404)

	// A pending intent is cancelled by pause, with its message marked failed.
	intent := seedIntent(t, f, connected[0])
	paused, err := s.UpdateFacebookConnection(ctx, f.tenant, f.owner, id, "pause")
	if err != nil || paused.Status != types.ChannelStatusPaused || paused.Enabled || paused.Version != 2 {
		t.Fatal("pause", err)
	}
	var row messagingmodel.OutboundIntent
	f.db.Where("id = ?", intent.ID).First(&row)
	var message messagingmodel.Message
	f.db.Where("id = ?", intent.MessageID).First(&message)
	if row.Status != types.IntentCancelled || message.DeliveryStatus != types.DeliveryFailed {
		t.Fatal("pending intent not cancelled", row.Status, message.DeliveryStatus)
	}
	_, err = s.UpdateFacebookConnection(ctx, f.tenant, f.owner, id, "pause")
	if err != nil || len(connector.upserts) != 3 {
		t.Fatal("repeated action must re-sync idempotently", err)
	}
	resumed, err := s.UpdateFacebookConnection(ctx, f.tenant, f.owner, id, "resume")
	if err != nil || resumed.Status != types.ChannelStatusActive || !resumed.Enabled {
		t.Fatal("resume", err)
	}
	connector.fail = types.ErrConnectorUnavailable
	_, err = s.UpdateFacebookConnection(ctx, f.tenant, f.owner, id, "disconnect")
	requireSalesStatus(t, err, 503)
	var conn messagingmodel.Connection
	f.db.Where("id = ?", id).First(&conn)
	if conn.Status != types.ChannelStatusDisconnected || conn.LastErrorCode == nil {
		t.Fatal("sync failure must be recorded on the committed row")
	}
	connector.fail = nil
	if _, err = s.UpdateFacebookConnection(ctx, f.tenant, f.owner, id, "disconnect"); err != nil || !connector.upserts[len(connector.upserts)-1].UnsubscribeOnDisable {
		t.Fatal("disconnect retry", err)
	}
	// After disconnect the Page is free for another tenant.
	if _, err := s.ConnectFacebookPages(ctx, f.other, f.owner, session.OAuthSessionID, []string{"page-a"}); err != nil {
		t.Fatal("handover after disconnect", err)
	}
}

func seedIntent(t *testing.T, f salesFixture, conn types.ChannelConnection) types.OutboundIntent {
	t.Helper()
	thread := types.MessagingThread{ID: messagingID(), TenantID: f.tenant, ExternalKey: "facebook:" + uuid.NewString(), DisplayName: "Customer", CreatedAt: time.Now(), UpdatedAt: time.Now(), ConnectionID: conn.ID, RecipientID: "psid"}
	r := repository.NewMessagingRepository(f.db)
	if _, err := r.UpsertThread(context.Background(), thread); err != nil {
		t.Fatal(err)
	}
	turn := types.MessagingTurn{ID: messagingID(), TenantID: f.tenant, ThreadID: thread.ID, Lifecycle: "open", Control: "sales", AssigneeID: &f.sale, ContextRevision: 1, ControlGeneration: 1, Version: 1, LastActivityAt: time.Now()}
	if err := r.SaveTurn(context.Background(), turn, true); err != nil {
		t.Fatal(err)
	}
	message := types.MessagingMessage{ID: messagingID(), TenantID: f.tenant, ThreadID: thread.ID, TurnID: turn.ID, Direction: "outbound", ActorID: f.sale, Body: "hi", RequestKey: uuid.NewString(), RequestHash: "h", Status: "channel", CreatedAt: time.Now(), DeliveryStatus: types.DeliverySending}
	if err := r.InsertMessage(context.Background(), message); err != nil {
		t.Fatal(err)
	}
	intent := types.OutboundIntent{ID: messagingID(), TenantID: f.tenant, ConnectionID: conn.ID, MessageID: message.ID, ClientMessageID: message.RequestKey, RequestHash: "h", RecipientID: "psid", Status: types.IntentPending, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	attempt := types.OutboundAttempt{ID: messagingID(), TenantID: f.tenant, IntentID: intent.ID, Status: types.IntentPending, ConnectionVersion: conn.Version, CreatedAt: time.Now()}
	if err := r.InsertOutbound(context.Background(), intent, attempt, &model.DomainEvent{ID: uuid.NewString(), TenantID: f.tenant, EventType: model.DomainEventOutboundIntent, AggregateID: intent.ID, Payload: "{}", State: model.DomainEventStatePending}); err != nil {
		t.Fatal(err)
	}
	return intent
}

// Real channel send: assigned-sale gate, active connection, idempotent intent + outbox, then delivery results.
func TestChannelSendAndDeliveryResults(t *testing.T) {
	f, channels, _ := channelServiceFixture(t)
	ctx := context.Background()
	connected, _ := connectPages(t, channels, f.tenant, f.owner, "page-a")
	conn := connected[0]
	in := NewIngressService(repository.NewIngressRepository(f.db))
	e := ingress.Event{SchemaVersion: 1, EventID: uuid.NewString(), CorrelationID: uuid.NewString(), Platform: "facebook", ConnectionID: conn.ID, ConnectionVersion: conn.Version, TenantID: f.tenant, AccountID: "page-a", ConversationKey: "psid-1", MessageID: "m1", SourceIdentity: "message:m1", Sender: "psid-1", Recipient: "page-a", Direction: "inbound", Type: ingress.TypeMessage, Text: "hello", SourceAt: time.Now(), ReceivedAt: time.Now(), Acquisition: "webhook", Purpose: "live"}
	if err := in.Consume(ctx, e); err != nil {
		t.Fatal(err)
	}
	media := e
	media.EventID, media.MessageID, media.SourceIdentity, media.Type, media.Text = uuid.NewString(), "m2", "message:m2", ingress.TypeUnsupportedMedia, ""
	media.Attachments = []ingress.Attachment{{Type: "image"}, {Type: "sticker"}}
	if err := in.Consume(ctx, media); err != nil {
		t.Fatal(err)
	}
	m := NewMessagingService(repository.NewMessagingRepository(f.db))
	page, err := m.Conversations(ctx, f.tenant, f.sale, types.MessagingQuery{Limit: 20})
	if err != nil || len(page.Items) != 1 {
		t.Fatal(err)
	}
	thread := page.Items[0].Thread
	if thread.ConnectionID != conn.ID || thread.RecipientID != "psid-1" {
		t.Fatal("thread channel binding", thread)
	}
	detail, _ := m.Detail(ctx, f.tenant, f.sale, thread.ID, types.MessagingQuery{Limit: 20})
	if len(detail.Messages) != 2 || detail.Messages[0].ContentType != types.ContentUnsupportedMedia || strings.Join(detail.Messages[0].MediaKinds, ",") != "image,sticker" || detail.Messages[0].Body != "" {
		t.Fatal("media placeholder", detail.Messages)
	}
	turn, err := m.Command(ctx, f.tenant, f.sale, page.Items[0].Turn.ID, "claim", types.MessagingCommand{ExpectedVersion: page.Items[0].Turn.Version})
	if err != nil {
		t.Fatal(err)
	}
	input := types.ChannelSendInput{ClientMessageID: uuid.NewString(), Text: "Xin chào", ExpectedVersion: turn.Version, ContextRevision: turn.ContextRevision}
	_, err = m.SendChannel(ctx, f.tenant, f.peer, turn.ID, input)
	requireSalesStatus(t, err, 403)
	_, err = m.SendSimulation(ctx, f.tenant, f.sale, turn.ID, types.MessagingCommand{ExpectedVersion: turn.Version, ContextRevision: turn.ContextRevision, Body: "x", RequestKey: uuid.NewString()})
	requireSalesStatus(t, err, 403)
	sent, err := m.SendChannel(ctx, f.tenant, f.sale, turn.ID, input)
	if err != nil || sent.Status != "channel" || sent.DeliveryStatus != types.DeliverySending {
		t.Fatal("send", err)
	}
	replay, err := m.SendChannel(ctx, f.tenant, f.sale, turn.ID, input)
	if err != nil || replay.ID != sent.ID {
		t.Fatal("idempotent replay", err)
	}
	changed := input
	changed.Text = "different"
	_, err = m.SendChannel(ctx, f.tenant, f.sale, turn.ID, changed)
	requireSalesStatus(t, err, 409)
	var events []model.DomainEvent
	f.db.Where("tenant_id = ? AND event_type = ?", f.tenant, model.DomainEventOutboundIntent).Find(&events)
	var intent outbound.IntentEvent
	if len(events) != 1 || json.Unmarshal([]byte(events[0].Payload), &intent) != nil || intent.Validate() != nil || intent.RecipientID != "psid-1" || intent.AccountID != "page-a" || intent.ConnectionVersion != conn.Version {
		t.Fatal("outbox intent", events)
	}

	delivery := NewDeliveryService(repository.NewDeliveryRepository(f.db))
	result := outbound.ResultEvent{SchemaVersion: 1, IntentID: intent.IntentID, AttemptID: intent.AttemptID, TenantID: f.tenant, ConnectionID: conn.ID, Status: outbound.StatusFailed, ErrorCategory: outbound.ErrorAuthExpired, ErrorCode: "190", OccurredAt: time.Now()}
	for i := 0; i < 2; i++ {
		if err := delivery.ConsumeResult(ctx, result); err != nil {
			t.Fatal(err)
		}
	}
	late := result
	late.Status, late.ErrorCategory = outbound.StatusSent, ""
	if err := delivery.ConsumeResult(ctx, late); err != nil {
		t.Fatal(err)
	}
	detail, _ = m.Detail(ctx, f.tenant, f.sale, thread.ID, types.MessagingQuery{Limit: 20})
	if detail.Messages[0].DeliveryStatus != types.DeliveryFailed || detail.Messages[0].DeliveryErrorReason != outbound.ErrorAuthExpired {
		t.Fatal("recorded outcome regressed or missing", detail.Messages[0])
	}
	list, _ := channels.ListFacebookConnections(ctx, f.tenant, f.sale)
	if list[0].Status != types.ChannelStatusAuthExpired {
		t.Fatal("auth_expired not surfaced", list[0].Status)
	}
	turn, _ = repository.NewMessagingRepository(f.db).Turn(ctx, f.tenant, turn.ID)
	_, err = m.SendChannel(ctx, f.tenant, f.sale, turn.ID, types.ChannelSendInput{ClientMessageID: uuid.NewString(), Text: "again", ExpectedVersion: turn.Version, ContextRevision: turn.ContextRevision})
	requireSalesStatus(t, err, 409)
	// Reconnecting re-authorizes the same binding and clears auth_expired.
	reconnected, _ := connectPages(t, channels, f.tenant, f.owner, "page-a")
	if reconnected[0].ID != conn.ID || reconnected[0].Status != types.ChannelStatusActive || reconnected[0].Version <= conn.Version {
		t.Fatal("reconnect", reconnected)
	}
	health := outbound.ConnectionHealthEvent{SchemaVersion: 1, EventID: uuid.NewString(), TenantID: f.tenant, ConnectionID: conn.ID, ConnectionVersion: conn.Version, Status: outbound.ErrorAuthExpired, ErrorCode: "190", OccurredAt: time.Now()}
	if err := delivery.ConsumeHealth(ctx, health); err != nil {
		t.Fatal(err)
	}
	list, _ = channels.ListFacebookConnections(ctx, f.tenant, f.owner)
	if list[0].Status != types.ChannelStatusActive {
		t.Fatal("stale health event expired a reconnected binding")
	}
	health.ConnectionVersion = reconnected[0].Version
	_ = delivery.ConsumeHealth(ctx, health)
	list, _ = channels.ListFacebookConnections(ctx, f.tenant, f.owner)
	if list[0].Status != types.ChannelStatusAuthExpired {
		t.Fatal("current health event ignored")
	}
}
