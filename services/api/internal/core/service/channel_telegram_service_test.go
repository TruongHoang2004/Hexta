package service

import (
	"context"
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

const (
	tgToken      = "777001:AAFixtureSecretPart_aaaaaaaaaaaaaaaaaaaa"
	tgRotated    = "777001:AAFixtureSecretPart_bbbbbbbbbbbbbbbbbbbb"
	tgOtherToken = "888002:AAFixtureSecretPart_cccccccccccccccccccc"
)

func deref(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func telegramFixture(t *testing.T) (salesFixture, *ChannelService, *fakeConnector) {
	f, s, connector := channelServiceFixture(t)
	connector.bots = map[string]types.TelegramBot{
		tgToken:      {ID: "777001", DisplayName: "Shop Bot", Username: "shop_bot"},
		tgRotated:    {ID: "777001", DisplayName: "Shop Bot 2", Username: "shop_bot"},
		tgOtherToken: {ID: "888002", DisplayName: "Other Bot", Username: "other_bot"},
	}
	return f, s, connector
}

func TestTelegramConnectOwnerOnlyAndOneBotOneTenant(t *testing.T) {
	f, s, connector := telegramFixture(t)
	ctx := context.Background()
	_, err := s.ConnectTelegramBot(ctx, f.tenant, f.sale, tgToken)
	requireSalesStatus(t, err, 403)
	if connector.verifies != 0 {
		t.Fatal("a non-owner must never trigger a Telegram call")
	}
	for _, bad := range []string{"", "not-a-token", "777001:short"} {
		_, err = s.ConnectTelegramBot(ctx, f.tenant, f.owner, bad)
		requireSalesStatus(t, err, 400)
	}
	_, err = s.ConnectTelegramBot(ctx, f.tenant, f.owner, "999999:AAFixtureSecretPart_unknownunknownunkno")
	requireSalesStatus(t, err, 400)
	connector.verifyErr = types.ErrConnectorProvider
	_, err = s.ConnectTelegramBot(ctx, f.tenant, f.owner, tgToken)
	requireSalesStatus(t, err, 502)
	connector.verifyErr = nil

	conn, err := s.ConnectTelegramBot(ctx, f.tenant, f.owner, tgToken)
	if err != nil || conn.Status != types.ChannelStatusActive || conn.AccountID != "777001" || conn.AccountUsername != "shop_bot" || conn.Version != 1 {
		t.Fatal("connect", err, conn)
	}
	last := connector.upserts[len(connector.upserts)-1]
	if last.BotToken != tgToken || last.AppID != types.ChannelAppTelegram || last.Platform != types.ChannelPlatformTelegram || !last.Enabled {
		t.Fatal("connector sync", last.AppID, last.Platform)
	}
	// The API stores no token anywhere it owns.
	var leaks, audits int64
	var rows []messagingmodel.Connection
	if err := f.db.Where("tenant_id = ?", f.tenant).Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if strings.Contains(strings.Join([]string{row.AccountID, row.AccountName, row.AccountUsername, deref(row.LastErrorCode), deref(row.LastErrorMessage)}, " "), "AAFixture") {
			leaks++
		}
	}
	var logs []model.AuditLog
	if err := f.db.Where("tenant_id = ?", f.tenant).Find(&logs).Error; err != nil {
		t.Fatal(err)
	}
	var connectAudits int64
	for _, log := range logs {
		if strings.Contains(log.AfterState+deref(log.BeforeState), "AAFixture") {
			audits++
		}
		if log.Action == "channel.telegram.connect" {
			connectAudits++
		}
	}
	if leaks != 0 || audits != 0 || connectAudits != 1 {
		t.Fatal("bot token persisted by the API or connect not audited", leaks, audits, connectAudits)
	}

	// 1 Bot = 1 tenant: another workspace cannot bind the same bot.
	_, err = s.ConnectTelegramBot(ctx, f.other, f.owner, tgRotated)
	requireSalesStatus(t, err, 409)
	// Reconnecting in the same tenant re-activates the binding in place.
	again, err := s.ConnectTelegramBot(ctx, f.tenant, f.owner, tgRotated)
	if err != nil || again.ID != conn.ID || again.Version != 2 || again.AccountName != "Shop Bot 2" {
		t.Fatal("reconnect", err, again)
	}
	list, err := s.ListTelegramConnections(ctx, f.tenant, f.sale)
	if err != nil || len(list) != 1 || list[0].Platform != types.ChannelPlatformTelegram {
		t.Fatal("member list", err)
	}
	if fb, _ := s.ListFacebookConnections(ctx, f.tenant, f.sale); len(fb) != 0 {
		t.Fatal("telegram bot listed as a Facebook Page")
	}
	// A Facebook action cannot address a Telegram connection.
	_, err = s.UpdateFacebookConnection(ctx, f.tenant, f.owner, conn.ID, "pause")
	requireSalesStatus(t, err, 404)
}

func TestTelegramRotatePauseDisconnect(t *testing.T) {
	f, s, connector := telegramFixture(t)
	ctx := context.Background()
	conn, err := s.ConnectTelegramBot(ctx, f.tenant, f.owner, tgToken)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.UpdateTelegramConnection(ctx, f.tenant, f.sale, conn.ID, "rotate_token", tgRotated)
	requireSalesStatus(t, err, 403)
	_, err = s.UpdateTelegramConnection(ctx, f.tenant, f.owner, conn.ID, "rotate_token", "")
	requireSalesStatus(t, err, 400)
	_, err = s.UpdateTelegramConnection(ctx, f.tenant, f.owner, conn.ID, "pause", tgRotated)
	requireSalesStatus(t, err, 400)
	_, err = s.UpdateTelegramConnection(ctx, f.tenant, f.owner, conn.ID, "rotate_token", tgOtherToken)
	requireSalesStatus(t, err, 409)

	// auth_expired from a send result, then rotation with the same bot recovers it.
	delivery := NewDeliveryService(repository.NewDeliveryRepository(f.db))
	health := outbound.ConnectionHealthEvent{SchemaVersion: 1, EventID: uuid.NewString(), TenantID: f.tenant, ConnectionID: conn.ID, ConnectionVersion: conn.Version, Status: outbound.ErrorAuthExpired, ErrorCode: "401", OccurredAt: time.Now()}
	if err := delivery.ConsumeHealth(ctx, health); err != nil {
		t.Fatal(err)
	}
	list, _ := s.ListTelegramConnections(ctx, f.tenant, f.owner)
	if list[0].Status != types.ChannelStatusAuthExpired || !strings.Contains(list[0].LastErrorMessage, "Telegram") {
		t.Fatal("auth_expired", list[0])
	}
	rotated, err := s.UpdateTelegramConnection(ctx, f.tenant, f.owner, conn.ID, "rotate_token", tgRotated)
	if err != nil || rotated.Status != types.ChannelStatusActive || !rotated.Enabled || rotated.Version != conn.Version+1 || rotated.LastErrorMessage != "" {
		t.Fatal("rotate", err, rotated)
	}
	if last := connector.upserts[len(connector.upserts)-1]; last.BotToken != tgRotated || last.Version != rotated.Version {
		t.Fatal("rotation must re-sync the token")
	}

	// Pause cancels pending intents and omits the token from the sync.
	intent := seedIntent(t, f, *rotated)
	paused, err := s.UpdateTelegramConnection(ctx, f.tenant, f.owner, conn.ID, "pause", "")
	if err != nil || paused.Status != types.ChannelStatusPaused || paused.Enabled {
		t.Fatal("pause", err)
	}
	var row messagingmodel.OutboundIntent
	f.db.Where("id = ?", intent.ID).First(&row)
	if row.Status != types.IntentCancelled {
		t.Fatal("pending intent not cancelled", row.Status)
	}
	if last := connector.upserts[len(connector.upserts)-1]; last.BotToken != "" || last.Enabled {
		t.Fatal("pause sync")
	}
	if _, err := s.UpdateTelegramConnection(ctx, f.tenant, f.owner, conn.ID, "disconnect", ""); err != nil || !connector.upserts[len(connector.upserts)-1].UnsubscribeOnDisable {
		t.Fatal("disconnect", err)
	}
	_, err = s.UpdateTelegramConnection(ctx, f.tenant, f.owner, conn.ID, "rotate_token", tgRotated)
	requireSalesStatus(t, err, 409)
	// After disconnect the bot is free for another workspace.
	if _, err := s.ConnectTelegramBot(ctx, f.other, f.owner, tgToken); err != nil {
		t.Fatal("handover after disconnect", err)
	}
}

// Telegram ingress creates a Telegram thread; sends use the bound chat and report Telegram failure categories.
func TestTelegramIngressAndSend(t *testing.T) {
	f, channels, _ := telegramFixture(t)
	ctx := context.Background()
	conn, err := channels.ConnectTelegramBot(ctx, f.tenant, f.owner, tgToken)
	if err != nil {
		t.Fatal(err)
	}
	in := NewIngressService(repository.NewIngressRepository(f.db))
	e := ingress.Event{SchemaVersion: 1, EventID: uuid.NewString(), CorrelationID: uuid.NewString(), Platform: "telegram", ConnectionID: conn.ID, ConnectionVersion: conn.Version, TenantID: f.tenant, AccountID: "777001", ConversationKey: "4242", MessageID: "7", SourceIdentity: "message:4242:7", Sender: "4242", Recipient: "777001", Direction: "inbound", Type: ingress.TypeMessage, Text: "Xin chào", SourceAt: time.Now(), ReceivedAt: time.Now(), Acquisition: "webhook", Purpose: "live"}
	if err := in.Consume(ctx, e); err != nil {
		t.Fatal(err)
	}
	voice := e
	voice.EventID, voice.MessageID, voice.SourceIdentity, voice.Type, voice.Text = uuid.NewString(), "8", "message:4242:8", ingress.TypeUnsupportedMedia, ""
	voice.Attachments = []ingress.Attachment{{Type: "audio"}}
	if err := in.Consume(ctx, voice); err != nil {
		t.Fatal(err)
	}
	m := NewMessagingService(repository.NewMessagingRepository(f.db))
	page, err := m.Conversations(ctx, f.tenant, f.sale, types.MessagingQuery{Limit: 20})
	if err != nil || len(page.Items) != 1 {
		t.Fatal(err)
	}
	thread := page.Items[0].Thread
	if types.ChannelThreadPlatform(thread.ExternalKey) != "telegram" || thread.DisplayName != "Telegram" || thread.RecipientID != "4242" || thread.ConnectionID != conn.ID {
		t.Fatal("telegram thread", thread)
	}
	detail, _ := m.Detail(ctx, f.tenant, f.sale, thread.ID, types.MessagingQuery{Limit: 20})
	if len(detail.Messages) != 2 || detail.Messages[0].ContentType != types.ContentUnsupportedMedia || detail.Messages[0].MediaKinds[0] != "audio" {
		t.Fatal("voice placeholder", detail.Messages)
	}
	// The simulator cannot write into a Telegram thread.
	_, err = m.Inbound(ctx, f.tenant, f.sale, types.MessagingInbound{ExternalKey: thread.ExternalKey, DisplayName: "x", Body: "spoof", RequestKey: uuid.NewString()})
	requireSalesStatus(t, err, 400)
	turn, err := m.Command(ctx, f.tenant, f.sale, page.Items[0].Turn.ID, "claim", types.MessagingCommand{ExpectedVersion: page.Items[0].Turn.Version})
	if err != nil {
		t.Fatal(err)
	}
	sent, err := m.SendChannel(ctx, f.tenant, f.sale, turn.ID, types.ChannelSendInput{ClientMessageID: uuid.NewString(), Text: "Cảm ơn", ExpectedVersion: turn.Version, ContextRevision: turn.ContextRevision})
	if err != nil || sent.DeliveryStatus != types.DeliverySending {
		t.Fatal("send", err)
	}
	var row messagingmodel.OutboundIntent
	f.db.Where("tenant_id = ? AND message_id = ?", f.tenant, sent.ID).First(&row)
	var attempt messagingmodel.OutboundAttempt
	f.db.Where("tenant_id = ? AND intent_id = ?", f.tenant, row.ID).First(&attempt)
	delivery := NewDeliveryService(repository.NewDeliveryRepository(f.db))
	result := outbound.ResultEvent{SchemaVersion: 1, IntentID: row.ID, AttemptID: attempt.ID, TenantID: f.tenant, ConnectionID: conn.ID, Status: outbound.StatusFailed, ErrorCategory: outbound.ErrorBotBlocked, ErrorCode: "403", OccurredAt: time.Now()}
	if err := delivery.ConsumeResult(ctx, result); err != nil {
		t.Fatal(err)
	}
	detail, _ = m.Detail(ctx, f.tenant, f.sale, thread.ID, types.MessagingQuery{Limit: 20})
	last := detail.Messages[0]
	if last.ID != sent.ID || last.DeliveryStatus != types.DeliveryFailed || last.DeliveryErrorReason != outbound.ErrorBotBlocked {
		t.Fatal("bot_blocked outcome", last)
	}
	// A blocked user does not expire the bot connection.
	if list, _ := channels.ListTelegramConnections(ctx, f.tenant, f.sale); list[0].Status != types.ChannelStatusActive {
		t.Fatal("bot_blocked must not change connection status")
	}
	// A paused bot locks sending.
	if _, err := channels.UpdateTelegramConnection(ctx, f.tenant, f.owner, conn.ID, "pause", ""); err != nil {
		t.Fatal(err)
	}
	turn, _ = repository.NewMessagingRepository(f.db).Turn(ctx, f.tenant, turn.ID)
	_, err = m.SendChannel(ctx, f.tenant, f.sale, turn.ID, types.ChannelSendInput{ClientMessageID: uuid.NewString(), Text: "again", ExpectedVersion: turn.Version, ContextRevision: turn.ContextRevision})
	requireSalesStatus(t, err, 409)
}
