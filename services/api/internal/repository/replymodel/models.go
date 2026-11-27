// Package replymodel holds the GORM rows of tenant reply settings and bot reply proposals.
package replymodel

import "time"

type Settings struct {
	TenantID  string
	ReplyMode string
	Version   int64
	UpdatedBy string
	UpdatedAt time.Time
}

func (Settings) TableName() string { return "tenant_reply_settings" }

type Proposal struct {
	ID                string
	TenantID          string
	ThreadID          string
	TurnID            string
	TriggerMessageID  string
	ControlGeneration int64
	ContextRevision   int64
	Status            string
	Route             *string
	Source            *string
	ScriptVersionID   *string
	AIRunID           *string `gorm:"column:ai_run_id"`
	Text              *string
	Reason            *string
	RequestedBy       *string
	DecidedBy         *string
	SentMessageID     *string
	// CitedSourceIDs is a JSON string array (jsonb).
	CitedSourceIDs string `gorm:"column:cited_source_ids"`
	Attempts       int
	AvailableAt    time.Time
	Version        int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (Proposal) TableName() string { return "messaging_reply_proposals" }

// ProposalLabel joins a proposal with its script source label for the Desk.
type ProposalLabel struct {
	Proposal
	ScriptName    *string
	ScriptVersion *int64
}
