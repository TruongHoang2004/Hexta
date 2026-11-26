package types

import (
	"testing"
	"time"
)

func TestMessagingUnknownControlDoesNotAuthorizeBot(t *testing.T) {
	actor := "sale"
	for _, turn := range []MessagingTurn{
		{Control: "system", AssigneeID: &actor},
		{Control: "system", NeedsHuman: true},
		{Control: "unexpected"},
	} {
		if turn.HandlingState() != "UNKNOWN" {
			t.Fatalf("unsafe control combination projected as %s", turn.HandlingState())
		}
	}
}

// needs_human is an escalation flag that survives a claim; it never changes who controls the turn.
func TestMessagingNeedsHumanKeepsStaffControl(t *testing.T) {
	actor := "sale"
	cases := map[string]MessagingTurn{
		"WAITING_STAFF": {Control: "sales", NeedsHuman: true},
		"STAFF_ACTIVE":  {Control: "sales", AssigneeID: &actor, NeedsHuman: true},
		"BOT_ACTIVE":    {Control: "system"},
	}
	for want, turn := range cases {
		if got := turn.HandlingState(); got != want {
			t.Fatalf("%+v projected as %s, want %s", turn, got, want)
		}
	}
}

// A closed turn keeps its last control fields, so clients are told CLOSED rather than a stale handler.
func TestMessagingClosedTurnReportsClosedHandlingState(t *testing.T) {
	actor := "sale"
	closed := MessagingTurn{Lifecycle: TurnClosed, Control: "sales", AssigneeID: &actor}
	if closed.HandlingState() != "STAFF_ACTIVE" || closed.ReportedHandlingState() != HandlingClosed {
		t.Fatalf("closed turn reported as %s (control projection %s)", closed.ReportedHandlingState(), closed.HandlingState())
	}
	if open := (MessagingTurn{Lifecycle: TurnOpen, Control: "system"}); open.ReportedHandlingState() != "BOT_ACTIVE" {
		t.Fatalf("open turn reported as %s", open.ReportedHandlingState())
	}
}

func TestMessagingTurnExpiryEligibility(t *testing.T) {
	now := time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)
	due := now.Add(-TurnInactivityWindow)
	cases := []struct {
		name string
		turn MessagingTurn
		want bool
	}{
		{"exactly 24h inactive", MessagingTurn{Lifecycle: TurnOpen, LastActivityAt: due}, true},
		{"older than 24h", MessagingTurn{Lifecycle: TurnOpen, LastActivityAt: due.Add(-time.Hour)}, true},
		{"recent activity", MessagingTurn{Lifecycle: TurnOpen, LastActivityAt: due.Add(time.Second)}, false},
		{"needs human blocks expiry", MessagingTurn{Lifecycle: TurnOpen, NeedsHuman: true, LastActivityAt: due.Add(-48 * time.Hour)}, false},
		{"already closed", MessagingTurn{Lifecycle: TurnClosed, LastActivityAt: due.Add(-time.Hour)}, false},
	}
	for _, c := range cases {
		if got := c.turn.ExpiryDue(now); got != c.want {
			t.Fatalf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestMessagingInboundPlacementIgnoresReplayedHistory(t *testing.T) {
	now := time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)
	lastActivity := now.Add(-30 * time.Hour)
	open := &MessagingTurn{Lifecycle: TurnOpen, LastActivityAt: lastActivity}
	closed := &MessagingTurn{Lifecycle: TurnClosed, LastActivityAt: lastActivity}
	cases := []struct {
		name     string
		latest   *MessagingTurn
		sourceAt time.Time
		want     string
	}{
		{"first message", nil, now, InboundNewTurn},
		{"open turn continues", open, now, InboundAppend},
		{"old message on open turn still appends", open, lastActivity.Add(-time.Hour), InboundAppend},
		{"genuinely new after expiry", closed, now.Add(-time.Minute), InboundNewTurn},
		{"replay of closed turn activity", closed, lastActivity, InboundHistory},
		{"delayed message from before the last activity", closed, lastActivity.Add(-time.Minute), InboundHistory},
		{"delayed beyond a whole window", closed, lastActivity.Add(time.Hour), InboundHistory},
	}
	for _, c := range cases {
		if got := InboundPlacement(c.latest, c.sourceAt, now); got != c.want {
			t.Fatalf("%s: got %s, want %s", c.name, got, c.want)
		}
	}
}
