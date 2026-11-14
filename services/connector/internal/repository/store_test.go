package repository

import (
	"context"
	"encoding/json"
	"github.com/TruongHoang2004/Hexta/packages/shared/ingress"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func fixtureStore(t *testing.T) (*Store, service.Connection) {
	t.Helper()
	dsn := os.Getenv("CONNECTOR_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("migrated isolated CONNECTOR_TEST_DATABASE_URL required")
	}
	if !strings.Contains(dsn, "/connector_fixture?") {
		t.Fatal("test fixture database required")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatal("test pool")
	}
	t.Cleanup(pool.Close)
	s := &Store{pool}
	c := service.Connection{ID: uuid.NewString(), TenantID: uuid.NewString(), Version: 1, Platform: "facebook", AccountID: uuid.NewString(), AppID: uuid.NewString(), Enabled: true}
	if s.PutConnection(context.Background(), c) != nil {
		t.Fatal("provision")
	}
	return s, c
}
func TestDurableAcceptanceDuplicateQuarantineAndFencing(t *testing.T) {
	s, c := fixtureStore(t)
	ctx := context.Background()
	now := time.Now().UTC()
	e := ingress.Event{Platform: "facebook", AccountID: c.AccountID, ConversationKey: "customer", MessageID: uuid.NewString(), Sender: "customer", Recipient: c.AccountID, Direction: "inbound", Type: "message", Text: "fixture", SourceAt: now, ReceivedAt: now, Acquisition: "webhook", Purpose: "live"}
	e.SourceIdentity = "message:" + e.MessageID
	items := []plugin.Item{{AccountID: c.AccountID, Event: e}, {AccountID: c.AccountID, Reason: "malformed_item"}, {AccountID: "unknown", Reason: "unsupported_event"}}
	receipt := uuid.NewString()
	for i := 0; i < 2; i++ {
		if err := s.Accept(ctx, c.AppID, c.Platform, receipt, items, now); err != nil {
			t.Fatal(err)
		}
	}
	var count int
	if err := s.Pool.QueryRow(ctx, "SELECT count(*) FROM connector.events WHERE receipt_id=$1", receipt).Scan(&count); err != nil || count != 3 {
		t.Fatal("batch/dedup", count, err)
	}
	// Crash after publish and before marking causes redelivery with stable event identity.
	var id string
	var payload []byte
	if s.Pool.QueryRow(ctx, "SELECT id,payload FROM connector.events WHERE receipt_id=$1 AND state='pending'", receipt).Scan(&id, &payload) != nil {
		t.Fatal("durability")
	}
	var decoded ingress.Event
	if json.Unmarshal(payload, &decoded) != nil || decoded.TenantID != c.TenantID {
		t.Fatal("trusted mapping")
	}
	// Target this event without depending on global queue order in the shared test fixture.
	if _, err := s.Pool.Exec(ctx, "UPDATE connector.events SET state='leased',fence=1,available_at=now()-interval '1 second' WHERE id=$1", id); err != nil {
		t.Fatal(err)
	}
	if err := s.Finish(ctx, Leased{ID: id, Fence: 1}, true); err == nil {
		t.Fatal("expired lease published")
	}
	s.Pool.Exec(ctx, "UPDATE connector.events SET fence=2,available_at=now()+interval '30 seconds' WHERE id=$1", id)
	if s.Finish(ctx, Leased{ID: id, Fence: 1}, true) == nil {
		t.Fatal("stale fence")
	}
	if s.Finish(ctx, Leased{ID: id, Fence: 2}, false) != nil {
		t.Fatal("retry")
	}
	// Webhook/poll overlap uses provider message identity, not acquisition or receipt identity.
	items[0].Event.Acquisition = "poll"
	items[0].Event.Purpose = "history"
	if s.Accept(ctx, c.AppID, c.Platform, uuid.NewString(), items[:1], now) != nil {
		t.Fatal("overlap")
	}
	if s.Pool.QueryRow(ctx, "SELECT count(*) FROM connector.events WHERE event_key=$1", decoded.Identity()).Scan(&count) != nil || count != 1 {
		t.Fatal("overlap duplicated")
	}
	changed := c
	changed.TenantID = uuid.NewString()
	changed.Version++
	if s.PutConnection(ctx, changed) == nil {
		t.Fatal("tenant reassignment")
	}
	c.Enabled = false
	c.Version++
	if s.PutConnection(ctx, c) != nil {
		t.Fatal("disable")
	}
	r := uuid.NewString()
	if s.Accept(ctx, c.AppID, c.Platform, r, items[:1], now) != nil {
		t.Fatal("disabled receipt")
	}
	var reason string
	s.Pool.QueryRow(ctx, "SELECT reason FROM connector.events WHERE receipt_id=$1", r).Scan(&reason)
	if reason != "disabled_connection" {
		t.Fatal("disabled mapping", reason)
	}
}

type fetchFixture struct{ plugin.BasePlugin }

func (fetchFixture) Platform() string                  { return "fixture" }
func (fetchFixture) Capabilities() plugin.Capabilities { return plugin.Capabilities{Fetch: true} }
func (fetchFixture) Fetch(_ context.Context, r plugin.FetchRequest) (plugin.Page, error) {
	return plugin.Page{NextCursor: "next", Items: []plugin.Item{{AccountID: r.AccountID, Reason: "unsupported_fixture"}}}, nil
}
func TestSyncCheckpointAndCancellation(t *testing.T) {
	s, c := fixtureStore(t)
	ctx := context.Background()
	id := uuid.NewString()
	_, err := s.Pool.Exec(ctx, "INSERT INTO connector.sync_jobs(id,app_id,platform,account_id,mode,page_limit,state,available_at) VALUES($1,$2,'fixture',$3,'history',10,'pending',now())", id, c.AppID, c.AccountID)
	if err != nil {
		t.Fatal(err)
	}
	j, err := s.ClaimSync(ctx)
	if err != nil {
		t.Fatal(err)
	}
	page := plugin.Page{Items: []plugin.Item{{AccountID: c.AccountID, Reason: "fixture"}}, NextCursor: "next"}
	if err = s.CommitPage(ctx, j, page); err != nil {
		t.Fatal(err)
	}
	var cursor string
	s.Pool.QueryRow(ctx, "SELECT cursor FROM connector.sync_jobs WHERE id=$1", id).Scan(&cursor)
	if cursor != "next" {
		t.Fatal("checkpoint")
	}
	if s.CommitPage(ctx, j, page) == nil {
		t.Fatal("stale checkpoint")
	}
	s.Pool.Exec(ctx, "UPDATE connector.sync_jobs SET state='cancelled',fence=fence+1 WHERE id=$1", id)
	if s.CommitPage(ctx, j, page) == nil {
		t.Fatal("cancelled checkpoint")
	}
}

func fixtureEvent(c service.Connection, now time.Time) plugin.Item {
	e := ingress.Event{Platform: "facebook", AccountID: c.AccountID, ConversationKey: "customer", MessageID: uuid.NewString(), Sender: "customer", Recipient: c.AccountID, Direction: "inbound", Type: "message", Text: "fixture", SourceAt: now, ReceivedAt: now, Acquisition: "webhook", Purpose: "live"}
	e.SourceIdentity = "message:" + e.MessageID
	return plugin.Item{AccountID: c.AccountID, Event: e}
}

func TestAcceptDoesNotSerializeUnrelatedReceipts(t *testing.T) {
	s, blocked := fixtureStore(t)
	ctx := context.Background()
	now := time.Now().UTC()
	other := service.Connection{ID: uuid.NewString(), TenantID: uuid.NewString(), Version: 1, Platform: "facebook", AccountID: uuid.NewString(), AppID: blocked.AppID, Enabled: true}
	if s.PutConnection(ctx, other) != nil {
		t.Fatal("provision")
	}
	// Hold one connection row so an in-flight Accept stops inside its transaction.
	lock, err := s.Pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Rollback(ctx)
	if _, err = lock.Exec(ctx, "SELECT 1 FROM connector.connections WHERE id=$1 FOR UPDATE", blocked.ID); err != nil {
		t.Fatal(err)
	}
	inflight := make(chan error, 1)
	go func() {
		inflight <- s.Accept(ctx, blocked.AppID, blocked.Platform, uuid.NewString(), []plugin.Item{fixtureEvent(blocked, now)}, now)
	}()
	deadline := time.Now().Add(5 * time.Second)
	for {
		var waiting bool
		if err = s.Pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE wait_event_type='Lock' AND query LIKE '%connector.connections%FOR SHARE%')").Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("in-flight accept never reached the connection lookup")
		}
		time.Sleep(10 * time.Millisecond)
	}
	// Another receipt for a different Page must commit while the first transaction is still open.
	unrelated, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err = s.Accept(unrelated, other.AppID, other.Platform, uuid.NewString(), []plugin.Item{fixtureEvent(other, now)}, now); err != nil {
		t.Fatal("unrelated receipt was serialized behind an in-flight accept:", err)
	}
	if err = lock.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	if err = <-inflight; err != nil {
		t.Fatal("in-flight accept", err)
	}
}

func TestConcurrentAcceptDeduplicatesWithoutErrors(t *testing.T) {
	s, c := fixtureStore(t)
	ctx := context.Background()
	now := time.Now().UTC()
	first, second := fixtureEvent(c, now), fixtureEvent(c, now)
	receipt := uuid.NewString()
	run := func(n int, accept func(i int) error) {
		t.Helper()
		var wg sync.WaitGroup
		errs := make(chan error, n)
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				errs <- accept(i)
			}(i)
		}
		wg.Wait()
		close(errs)
		for err := range errs {
			if err != nil {
				t.Fatal("concurrent accept", err)
			}
		}
	}
	// Provider retries of one receipt race on the receipt row instead of failing on its primary key.
	run(8, func(int) error {
		return s.Accept(ctx, c.AppID, c.Platform, receipt, []plugin.Item{first}, now)
	})
	// Different receipts with the same messages in opposite order neither deadlock nor duplicate events.
	run(8, func(i int) error {
		items := []plugin.Item{first, second}
		if i%2 == 1 {
			items = []plugin.Item{second, first}
		}
		return s.Accept(ctx, c.AppID, c.Platform, uuid.NewString(), items, now)
	})
	for _, item := range []plugin.Item{first, second} {
		var count int
		if err := s.Pool.QueryRow(ctx, "SELECT count(*) FROM connector.events WHERE event_key=$1", service.Envelope(item.Event, c).Identity()).Scan(&count); err != nil || count != 1 {
			t.Fatal("event identity stored", count, err)
		}
	}
}

func TestAcceptRejectsBatchBeyondPendingCapacity(t *testing.T) {
	s, c := fixtureStore(t)
	ctx := context.Background()
	now := time.Now().UTC()
	var backlog int
	if err := s.Pool.QueryRow(ctx, "SELECT count(*) FROM connector.events WHERE state IN ('pending','leased')").Scan(&backlog); err != nil {
		t.Fatal(err)
	}
	previous := maxPendingEvents
	maxPendingEvents = backlog + 1
	t.Cleanup(func() { maxPendingEvents = previous })

	receipt := uuid.NewString()
	if s.Accept(ctx, c.AppID, c.Platform, receipt, []plugin.Item{fixtureEvent(c, now), fixtureEvent(c, now)}, now) == nil {
		t.Fatal("batch above capacity accepted")
	}
	var claimed bool
	if err := s.Pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM connector.receipts WHERE id=$1)", receipt).Scan(&claimed); err != nil || claimed {
		t.Fatal("rejected receipt was retained", claimed, err)
	}
	// The same receipt is admitted once it fits, so provider retries succeed after the backlog drains.
	if err := s.Accept(ctx, c.AppID, c.Platform, receipt, []plugin.Item{fixtureEvent(c, now)}, now); err != nil {
		t.Fatal("batch at capacity", err)
	}
}
