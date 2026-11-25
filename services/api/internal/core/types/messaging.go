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
	// CloseCause is TurnCloseExpired or TurnCloseManual on a closed turn; ClosedBy names the manual closer.
	CloseCause, ClosedBy string
	// Outcome (one of SessionOutcomes) and Note describe the session; both are empty until recorded.
	Outcome, Note string
	// LastHandlerID is the last staff member assigned to the turn; it survives return-to-bot and unassignment.
	LastHandlerID *string
	// ReviewPending marks a BOT_ACTIVE turn whose live bot proposal waits for staff review (owner decision D3,
	// 2026-11-26): the turn counts as waiting for staff and does not expire. It is cleared when the proposal is sent,
	// discarded or superseded and on every control change, so it is only ever set on a BOT_ACTIVE open turn.
	ReviewPending bool
}

// A conversation session is a turn: it ends by the fixed 24-hour inactivity expiry or a manual close.
const (
	TurnCloseExpired = "expired"
	TurnCloseManual  = "closed_manually"
	// MaxSessionNoteRunes bounds a session note.
	MaxSessionNoteRunes = 1000
)

// SessionOutcomes is the proposed outcome set for a conversation session (owner decision pending on the exact set).
var SessionOutcomes = map[string]bool{
	"order_placed": true, "consulted": true, "price_inquiry": true, "complaint": true, "no_response": true, "other": true,
}

// TurnClosure describes how a turn closes. Expiry and manual close share one closing write (CloseTurn).
type TurnClosure struct {
	Cause, ClosedBy, Outcome, Note string
}

// SessionInput closes a session manually or edits a closed session's outcome and note at the expected version.
type SessionInput struct {
	ExpectedVersion int64
	Outcome, Note   string
}

// SessionOrder is an order created from Desk during a session.
type SessionOrder struct {
	ID, OrderNumber, Status string
}

// MessagingSession is one turn of a thread with its message count, first message and linked orders.
type MessagingSession struct {
	Turn           MessagingTurn
	OpenedAt       time.Time
	MessageCount   int64
	FirstMessageID string
	Orders         []SessionOrder
}

// MessagingSessionPage lists a thread's sessions newest first.
type MessagingSessionPage struct {
	Items      []MessagingSession
	NextCursor string
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
