package service

import (
	"context"
	"github.com/TruongHoang2004/Hexta/packages/shared/ingress"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/messagingmodel"
	"github.com/google/uuid"
	"testing"
	"time"
)

func TestIngressAuthorityReplayAndNoAISideEffects(t *testing.T) {
	f := salesTestFixture(t)
	ctx := context.Background()
	id := uuid.NewString()
	if err := f.db.Exec("INSERT INTO messaging_connections(id,tenant_id,version,platform,account_id,enabled) VALUES(?,?,1,'facebook','page',true)", id, f.tenant).Error; err != nil {
		t.Fatal(err)
	}
	s := NewIngressService(repository.NewIngressRepository(f.db))
	e := ingress.Event{SchemaVersion: 1, EventID: uuid.NewString(), CorrelationID: uuid.NewString(), Platform: "facebook", ConnectionID: id, ConnectionVersion: 1, TenantID: f.tenant, AccountID: "page", ConversationKey: "psid", MessageID: "m1", SourceIdentity: "message:m1", Sender: "psid", Recipient: "page", Direction: "inbound", Type: "message", Text: "real text", SourceAt: time.Now(), ReceivedAt: time.Now(), Acquisition: "webhook", Purpose: "live"}
	if err := s.Consume(ctx, e); err != nil {
		t.Fatal(err)
	}
	e.EventID = uuid.NewString()
	e.Acquisition = "poll"
	if err := s.Consume(ctx, e); err != nil {
		t.Fatal(err)
	}
	for i, mode := range []string{"history", "echo", "disabled", "stale", "wrong-account"} {
		v := e
		v.MessageID = mode
		v.SourceIdentity = "message:" + mode
		v.EventID = uuid.NewString()
		switch i {
		case 0:
			v.Purpose = "history"
		case 1:
			v.Type = "echo"
			v.Direction = "outbound"
		case 2:
			f.db.Exec("UPDATE messaging_connections SET enabled=false WHERE id=?", id)
		case 3:
			f.db.Exec("UPDATE messaging_connections SET enabled=true,version=2 WHERE id=?", id)
		case 4:
			v.ConnectionVersion = 2
			v.AccountID = "different"
		}
		if err := s.Consume(ctx, v); err != nil {
			t.Fatal(err)
		}
	}
	var count int64
	f.db.Model(&messagingmodel.Message{}).Where("tenant_id=?", f.tenant).Count(&count)
	if count != 1 {
		t.Fatal("duplicate/unsafe event", count)
	}
	f.db.Model(&messagingmodel.Turn{}).Where("tenant_id=?", f.tenant).Count(&count)
	if count != 1 {
		t.Fatal("history opened turn")
	}
	v := e
	v.TenantID = uuid.NewString()
	if s.Consume(ctx, v) != nil {
		t.Fatal("cross tenant quarantine failed")
	}
	f.db.Table("messaging_ingress_quarantine").Count(&count)
	if count != 1 {
		t.Fatal("cross tenant was not quarantined")
	}
	var msg messagingmodel.Message
	f.db.Where("tenant_id=?", f.tenant).First(&msg)
	m := NewMessagingService(repository.NewMessagingRepository(f.db))
	if _, err := m.Inbound(ctx, f.tenant, f.sale, types.MessagingInbound{ExternalKey: "facebook:" + e.Key(), DisplayName: "Spoof", Body: "fake", RequestKey: uuid.NewString()}); err == nil {
		t.Fatal("simulator mutated real thread")
	}
}
