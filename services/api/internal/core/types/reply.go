package types

import (
	"errors"
	"time"
)

// Reply modes (MVP 5 FR-5.1, owner decision 2026-11-26). The owner chooses how much the bot may do alone while a
// turn is under bot handling.
const (
	// ReplyModeOff is the default for every tenant: new turns wait for staff and the bot never runs. A customer
	// message on a turn still under bot handling hands it to staff (HandoffReasonBotOff).
	ReplyModeOff = "off"
	// ReplyModeSuggest: the bot only proposes; nothing reaches the customer until a sale sends it.
	ReplyModeSuggest = "suggest"
	// ReplyModeAutoSend: the bot sends script answers and clarifying questions itself while the turn is BOT_ACTIVE.
	// AI-written answers still wait for staff review.
	ReplyModeAutoSend = "auto_send"

	// HandoffReasonBotOff is the handoff reason when a customer writes on a bot-handled turn while the bot is off.
	HandoffReasonBotOff = "bot_off"

	// BotActorID is the actor recorded for bot-authored messages, routing jobs and router transitions.
	BotActorID = "system:bot"

	// DeliveryCancelledReason marks an automatic send cancelled before dispatch because control changed.
	DeliveryCancelledReason = "cancelled"
)

// ReplySettings is a tenant's bot reply configuration. Configured is false while the owner has never saved it;
// the mode then reads as off.
type ReplySettings struct {
	TenantID, Mode, UpdatedBy string
	Configured                bool
	Version                   int64
	UpdatedAt                 time.Time
}

// BotOn reports whether the bot may handle turns (suggest or auto_send).
func (s ReplySettings) BotOn() bool { return s.Mode == ReplyModeSuggest || s.Mode == ReplyModeAutoSend }

// ValidReplyMode reports whether mode is one of the reply modes.
func ValidReplyMode(mode string) bool {
	return mode == ReplyModeOff || mode == ReplyModeSuggest || mode == ReplyModeAutoSend
}

// ReplySettingsUpdate is the owner's change with its optimistic version (0 when never configured).
type ReplySettingsUpdate struct {
	Mode            string
	ExpectedVersion int64
}

// Proposal statuses. routing is a durable router job; pending is a proposal waiting for a sale. The rest are terminal.
const (
	ProposalRouting     = "routing"
	ProposalPending     = "pending"
	ProposalSent        = "sent"
	ProposalAutoSent    = "auto_sent"
	ProposalDiscarded   = "discarded"
	ProposalSuperseded  = "superseded"
	ProposalHandedOff   = "handed_off"
	ProposalNoCandidate = "no_candidate"
)

// Router routes and candidate sources.
const (
	RouteScript  = "script"
	RouteAnswer  = "answer"
	RouteClarify = "clarify"
	RouteHandoff = "handoff"

	SourceScript = "script"
	SourceAI     = "ai"
)

// ReplyProposal is one router job and, once routed, its candidate. ControlGeneration and ContextRevision are the
// turn basis it was created under; ScriptVersionID or AIRunID names its source.
type ReplyProposal struct {
	ID, TenantID, ThreadID, TurnID, TriggerMessageID string
	ControlGeneration, ContextRevision               int64
	Status, Route, Source, ScriptVersionID, AIRunID  string
	Text, Reason, RequestedBy, DecidedBy             string
	SentMessageID                                    string
	Attempts                                         int
	AvailableAt                                      time.Time
	Version                                          int64
	CreatedAt, UpdatedAt                             time.Time
}

// Stale reports whether the turn's control changed (or it closed) since the proposal was created. A stale proposal
// can never be sent; this is the control-generation fence for bot output.
func (p ReplyProposal) Stale(turn MessagingTurn) bool {
	return turn.ID != p.TurnID || turn.Lifecycle != TurnOpen || turn.ControlGeneration != p.ControlGeneration
}

// ProposalRef names one due routing job.
type ProposalRef struct{ TenantID, ThreadID, ProposalID string }

// ProposalView is a proposal with its derived staleness for the Desk.
type ProposalView struct {
	Proposal ReplyProposal
	Stale    bool
	// ScriptName and ScriptVersion label a script source; empty when the source is AI or the script is unknown.
	ScriptName    string
	ScriptVersion int64
}

// ProposalSendInput sends a pending proposal as-is (empty Text) or edited.
type ProposalSendInput struct {
	ClientMessageID, Text string
}

// ReplyCandidate is the router decision for one customer message. It is untrusted until Go applies its gates.
type ReplyCandidate struct {
	Route, Source, ScriptVersionID, AIRunID, Text string
	// HandoffReason is a messaging handoff reason code when Route is handoff.
	HandoffReason string
	// Reason is a safe decision category kept on the proposal for review.
	Reason string
}

// AI run contract (transport-neutral mirror of the internal gRPC run).
const (
	AIRunQueued    = "queued"
	AIRunRunning   = "running"
	AIRunSucceeded = "succeeded"
	AIRunFailed    = "failed"
	AIRunCancelled = "cancelled"
)

// AIScript is an authorized active script snapshot item; ID is the script version ID.
type AIScript struct{ ID, Trigger, Text string }

// AIHistoryItem is one prior message: Role is user (customer) or assistant (bot/staff).
type AIHistoryItem struct{ Role, Text string }

// AIRunRequest asks the AI service for a candidate. RunID is the stable idempotency identity.
type AIRunRequest struct {
	TenantID, RunID, TurnID            string
	ContextRevision, ControlGeneration int64
	Message                            string
	Scripts                            []AIScript
	History                            []AIHistoryItem
	AllowGeneration                    bool
}

// AIRun is the AI service's run state. Every field is untrusted candidate data.
type AIRun struct {
	RunID, Status, Route, Text, ScriptID, Reason, ErrorCode string
	ContextRevision, ControlGeneration                      int64
}

// ErrAIUnavailable means no AI service is configured; the router then hands the conversation to staff.
var ErrAIUnavailable = errors.New("AI service is not configured")

// ErrAIRunNotFound means the AI service holds no run with this ID: StartRun never reached it (or was lost before
// commit). StartRun is idempotent on the run ID, so the router may issue it again.
var ErrAIRunNotFound = errors.New("AI run not found")

// ErrAIRunConflict means the AI service already holds this run ID with different input. The router must not read
// that run as its answer; it hands the turn to staff.
var ErrAIRunConflict = errors.New("AI run ID was accepted with different input")
