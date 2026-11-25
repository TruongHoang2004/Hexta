package dto

import "time"

// ReplySettingsResponse is the tenant's bot reply mode. off (the default) keeps the bot from running: new
// conversations wait for staff. configured=false means the owner never saved a mode; reply_mode then reads as off.
type ReplySettingsResponse struct {
	ReplyMode  string     `json:"reply_mode" enums:"off,suggest,auto_send"`
	Configured bool       `json:"configured"`
	Version    int64      `json:"version"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
}

// UpdateReplySettingsRequest changes the reply mode; expected_version is the version last read (0 before the
// first save).
type UpdateReplySettingsRequest struct {
	ReplyMode       string `json:"reply_mode" validate:"required,oneof=off suggest auto_send" enums:"off,suggest,auto_send"`
	ExpectedVersion int64  `json:"expected_version" validate:"min=0"`
}

// ReplyProposal is a bot reply proposal shown on the Desk. stale=true means conversation control changed after it
// was created: it can never be sent. status routing means the bot is still preparing it (no text yet).
type ReplyProposal struct {
	ID                string `json:"id"`
	TurnID            string `json:"turn_id"`
	Status            string `json:"status" enums:"routing,pending,sent,auto_sent,discarded,superseded,handed_off,no_candidate"`
	Route             string `json:"route,omitempty" enums:"script,answer,clarify,handoff"`
	Source            string `json:"source,omitempty" enums:"script,ai"`
	ScriptVersionID   string `json:"script_version_id,omitempty"`
	ScriptName        string `json:"script_name,omitempty"`
	ScriptVersion     int64  `json:"script_version,omitempty"`
	AIRunID           string `json:"ai_run_id,omitempty"`
	Text              string `json:"text,omitempty"`
	Reason            string `json:"reason,omitempty"`
	ControlGeneration int64  `json:"control_generation"`
	Stale             bool   `json:"stale"`
	// RequestedBy is set when a person asked for this proposal (regenerate).
	RequestedBy string    `json:"requested_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// SendProposalRequest sends a pending proposal: omit text to send it as-is, or pass the edited text. Retrying with
// the same client_message_id replays the original message.
type SendProposalRequest struct {
	ClientMessageID string `json:"client_message_id" validate:"required,uuid"`
	Text            string `json:"text" validate:"omitempty,max=2000"`
}
