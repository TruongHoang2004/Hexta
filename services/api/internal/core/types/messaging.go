package types

import "time"

type MessagingScope struct{ Active, Member, Owner bool }
type MessagingThread struct {
	ID, TenantID, ExternalKey, DisplayName string
	CreatedAt, UpdatedAt                   time.Time
	// ConnectionID/RecipientID bind a real-channel thread for outbound delivery; empty for simulator threads.
	ConnectionID, RecipientID string
}
type MessagingTurn struct {
	ID, TenantID, ThreadID, Lifecycle, Control string
	AssigneeID                                 *string
	NeedsHuman                                 bool
	// NeedsHumanReason is the handoff reason that raised NeedsHuman; empty when it was raised without one.
	NeedsHumanReason                            string
	ContextRevision, ControlGeneration, Version int64
	LastActivityAt                              time.Time
	// ClosedAt is set once the turn closes; a closed turn is never reopened.
	ClosedAt *time.Time
	// ReviewPending marks a BOT_ACTIVE turn whose live bot proposal waits for staff review (owner decision D3,
	// 2026-11-26): the turn counts as waiting for staff and does not expire. It is cleared when the proposal is sent,
	// discarded or superseded and on every control change, so it is only ever set on a BOT_ACTIVE open turn.
	ReviewPending bool
}

const (
	TurnOpen   = "open"
	TurnClosed = "closed"
	// TurnInactivityWindow is the single MVP inactivity period after which an open turn closes.
	TurnInactivityWindow = 24 * time.Hour
)

// ExpiryDue reports whether an open turn has been inactive for the whole window. needs_human and a bot proposal
// waiting for review always block expiry.
func (t MessagingTurn) ExpiryDue(now time.Time) bool {
	return t.Lifecycle == TurnOpen && !t.NeedsHuman && !t.ReviewPending && !t.LastActivityAt.After(now.Add(-TurnInactivityWindow))
}

// Inbound placements decide where a customer message lands relative to the thread's latest turn.
const (
	InboundAppend  = "append"  // continue the open turn
	InboundNewTurn = "new"     // open a new unassigned turn
	InboundHistory = "history" // record on the closed turn only; neither reopens nor extends it
)

// InboundPlacement places a customer message by its source event time. A message that is not newer than the
// closed turn's recorded activity, or that is already older than a whole inactivity window, is delayed or
// replayed history: it must not manufacture a new turn.
func InboundPlacement(latest *MessagingTurn, sourceAt, now time.Time) string {
	switch {
	case latest == nil:
		return InboundNewTurn
	case latest.Lifecycle == TurnOpen:
		return InboundAppend
	case !sourceAt.After(latest.LastActivityAt) || !sourceAt.After(now.Add(-TurnInactivityWindow)):
		return InboundHistory
	default:
		return InboundNewTurn
	}
}

// MessagingSummary is the tenant-wide Desk notification state.
type MessagingSummary struct {
	WaitingStaff int64
}

// TurnRef names one turn for background processing.
type TurnRef struct{ TenantID, ThreadID, TurnID string }
type MessagingMessage struct {
	ID, TenantID, ThreadID, TurnID, Direction, ActorID, Body, RequestKey, RequestHash, Status string
	CreatedAt                                                                                 time.Time
	ContentType, DeliveryStatus, DeliveryErrorReason                                          string
	// MediaKinds is set for unsupported_media placeholders (e.g. image, sticker); no media is stored.
	MediaKinds []string
}
type MessagingConversation struct {
	Thread MessagingThread
	Turn   *MessagingTurn
}
type MessagingDetail struct {
	Conversation MessagingConversation
	Messages     []MessagingMessage
	NextCursor   string
	// Proposals are the latest turn's live bot reply proposals (routing or pending) with derived staleness.
	Proposals []ProposalView
}
type MessagingPage struct {
	Items      []MessagingConversation
	NextCursor string
}
type MessagingQuery struct {
	Cursor, Filter string
	Limit          int
}
type MessagingInbound struct{ ExternalKey, DisplayName, Body, RequestKey string }
type MessagingCommand struct {
	ExpectedVersion, ContextRevision int64
	AssigneeID                       *string
	Body, RequestKey, Reason         string
}

// HandlingState projects existing durable control fields without a second state store.
// needs_human is an escalation flag, not a controller: it survives a claim until the assigned sale resolves it,
// and only ever blocks the bot. Unknown combinations fail closed; consumers must not treat them as bot permission.
func (t MessagingTurn) HandlingState() string {
	if t.Control == "system" && t.AssigneeID == nil && !t.NeedsHuman {
		return "BOT_ACTIVE"
	}
	if t.Control == "sales" && t.AssigneeID == nil {
		return "WAITING_STAFF"
	}
	if t.Control == "sales" && t.AssigneeID != nil {
		return "STAFF_ACTIVE"
	}
	return "UNKNOWN"
}
