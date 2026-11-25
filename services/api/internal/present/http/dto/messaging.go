package dto

import (
	"strings"
	"time"
)

type MessagingQuery struct {
	Cursor string `form:"cursor" validate:"omitempty,uuid"`
	Filter string `form:"filter" validate:"omitempty,oneof=all mine unassigned needs_human waiting bot"`
	Limit  int    `form:"limit" validate:"min=1,max=100"`
}
type MessagingInboundRequest struct {
	ExternalKey string `json:"external_key" validate:"required,max=120"`
	DisplayName string `json:"display_name" validate:"required,max=120"`
	Body        string `json:"body" validate:"required,max=4000"`
	RequestKey  string `json:"request_key" validate:"required,uuid"`
}
type MessagingCommandRequest struct {
	ExpectedVersion int64   `json:"expected_version" validate:"min=1"`
	AssigneeID      *string `json:"assignee_id" validate:"omitempty,min=1,max=255"`
}
type MessagingHandoffRequest struct {
	ExpectedVersion int64  `json:"expected_version" validate:"min=1"`
	Reason          string `json:"reason" validate:"required,oneof=customer_requested_staff insufficient_information ai_error business_review staff_handoff"`
}
type MessagingSendRequest struct {
	ExpectedVersion int64  `json:"expected_version" validate:"min=1"`
	ContextRevision int64  `json:"context_revision" validate:"min=1"`
	Body            string `json:"body" validate:"required,max=4000"`
	RequestKey      string `json:"request_key" validate:"required,uuid"`
}
type MessagingThread struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	ExternalKey string `json:"external_key"`
	// ConnectionID names the bound channel connection of a real conversation; empty for simulator threads.
	ConnectionID string    `json:"connection_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
type MessagingTurn struct {
	ID            string  `json:"id"`
	ThreadID      string  `json:"thread_id"`
	Lifecycle     string  `json:"lifecycle"`
	Control       string  `json:"control"`
	HandlingState string  `json:"handling_state"`
	AssigneeID    *string `json:"assignee_id"`
	NeedsHuman    bool    `json:"needs_human"`
	// NeedsHumanReason is the handoff reason code behind needs_human; omitted when none was given.
	NeedsHumanReason  string    `json:"needs_human_reason,omitempty"`
	ContextRevision   int64     `json:"context_revision"`
	ControlGeneration int64     `json:"control_generation"`
	Version           int64     `json:"version"`
	LastActivityAt    time.Time `json:"last_activity_at"`
	// ClosedAt is set on a closed turn (24-hour inactivity expiry or manual close); a new customer message opens a new turn.
	ClosedAt *time.Time `json:"closed_at,omitempty"`
	// CloseCause is expired or closed_manually on a closed turn.
	CloseCause string `json:"close_cause,omitempty" enums:"expired,closed_manually"`
	// ClosedBy is the member who closed the session manually.
	ClosedBy string `json:"closed_by,omitempty"`
	// Outcome and Note describe the session; expired sessions have no outcome until one is recorded.
	Outcome string `json:"outcome,omitempty" enums:"order_placed,consulted,price_inquiry,complaint,no_response,other"`
	Note    string `json:"note,omitempty"`
	// LastHandlerID is the last staff member assigned to the session (kept after return to the bot).
	LastHandlerID *string `json:"last_handler_id,omitempty"`
	// ReviewPending is true while the bot handles the turn and its proposal waits for staff review: the
	// conversation counts as waiting for staff and does not expire until the proposal is sent, discarded or
	// superseded.
	ReviewPending bool `json:"review_pending"`
}

// MessagingSessionRequest closes a session manually or edits a closed session's outcome and note.
type MessagingSessionRequest struct {
	ExpectedVersion int64  `json:"expected_version" validate:"min=1"`
	Outcome         string `json:"outcome" validate:"required,oneof=order_placed consulted price_inquiry complaint no_response other" enums:"order_placed,consulted,price_inquiry,complaint,no_response,other"`
	Note            string `json:"note" validate:"max=1000"`
}

// Normalize trims the note before validation, matching the service, so its length bound applies to the stored text.
func (r *MessagingSessionRequest) Normalize() {
	r.Note = strings.TrimSpace(r.Note)
}

// MessagingSessionQuery pages a conversation's sessions.
type MessagingSessionQuery struct {
	Cursor string `form:"cursor" validate:"omitempty,uuid"`
	Limit  int    `form:"limit" validate:"min=1,max=100"`
}

// MessagingSessionOrder is an order created from Desk during the session.
type MessagingSessionOrder struct {
	ID          string `json:"id"`
	OrderNumber string `json:"order_number"`
	Status      string `json:"status"`
}

// MessagingSession is one conversation session (turn).
type MessagingSession struct {
	Turn MessagingTurn `json:"turn"`
	// OpenedAt is the time of the session's earliest message.
	OpenedAt     time.Time `json:"opened_at"`
	MessageCount int64     `json:"message_count"`
	// FirstMessageID lets the thread jump to where the session starts.
	FirstMessageID string                  `json:"first_message_id"`
	Orders         []MessagingSessionOrder `json:"orders"`
}

// MessagingSessionPage lists a conversation's sessions newest first.
type MessagingSessionPage struct {
	Items      []MessagingSession `json:"items"`
	NextCursor string             `json:"next_cursor"`
}

// MessagingSummary is the tenant-wide Desk notification state.
type MessagingSummary struct {
	// WaitingStaffCount is the number of open conversations waiting for staff: unassigned under staff control, or
	// under bot handling with a proposal waiting for review (review_pending).
	WaitingStaffCount int64 `json:"waiting_staff_count"`
}

// SendMessageRequest sends a real-channel reply. Version/context guard against replying to stale context.
type SendMessageRequest struct {
	ClientMessageID string `json:"client_message_id" validate:"required,uuid"`
	Text            string `json:"text" validate:"required,max=2000"`
	ExpectedVersion int64  `json:"expected_version" validate:"min=1"`
	ContextRevision int64  `json:"context_revision" validate:"min=1"`
}
type PlaceholderMetaDTO struct {
	MediaKinds []string `json:"media_kinds"`
}
type MessagingMessage struct {
	ID                  string              `json:"id"`
	ThreadID            string              `json:"thread_id"`
	TurnID              string              `json:"turn_id"`
	Direction           string              `json:"direction"`
	ActorID             string              `json:"actor_id"`
	Body                string              `json:"body"`
	Status              string              `json:"status"`
	CreatedAt           time.Time           `json:"created_at"`
	ContentType         string              `json:"content_type"`
	PlaceholderMeta     *PlaceholderMetaDTO `json:"placeholder_meta"`
	DeliveryStatus      string              `json:"delivery_status"`
	DeliveryErrorReason string              `json:"delivery_error_reason"`
}

type MessagingConversation struct {
	Thread MessagingThread `json:"thread"`
	Turn   *MessagingTurn  `json:"turn"`
	Source string          `json:"source"`
}
type MessagingPage struct {
	Items      []MessagingConversation `json:"items"`
	NextCursor string                  `json:"next_cursor"`
}
type MessagingDetail struct {
	Conversation MessagingConversation `json:"conversation"`
	Messages     []MessagingMessage    `json:"messages"`
	NextCursor   string                `json:"next_cursor"`
	// Proposals are the latest turn's live bot reply proposals (at most one routing or pending proposal).
	Proposals []ReplyProposal `json:"proposals"`
}
