// Package persistencemap converts tenant reply settings and reply proposal rows to core values.
package persistencemap

import (
	"errors"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/replymodel"
)

// ErrStoredMode reports a reply mode column outside the known modes.
var ErrStoredMode = errors.New("stored reply mode is unknown")

func optional(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func text(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// Settings maps a stored row; an unknown mode fails instead of being read as a send permission.
func Settings(row replymodel.Settings) (types.ReplySettings, error) {
	if !types.ValidReplyMode(row.ReplyMode) {
		return types.ReplySettings{}, ErrStoredMode
	}
	return types.ReplySettings{TenantID: row.TenantID, Mode: row.ReplyMode, Configured: true, Version: row.Version, UpdatedBy: row.UpdatedBy, UpdatedAt: row.UpdatedAt}, nil
}

func SettingsRow(value types.ReplySettings) replymodel.Settings {
	return replymodel.Settings{TenantID: value.TenantID, ReplyMode: value.Mode, Version: value.Version, UpdatedBy: value.UpdatedBy, UpdatedAt: value.UpdatedAt}
}

func Proposal(row replymodel.Proposal) types.ReplyProposal {
	return types.ReplyProposal{
		ID: row.ID, TenantID: row.TenantID, ThreadID: row.ThreadID, TurnID: row.TurnID, TriggerMessageID: row.TriggerMessageID,
		ControlGeneration: row.ControlGeneration, ContextRevision: row.ContextRevision, Status: row.Status,
		Route: text(row.Route), Source: text(row.Source), ScriptVersionID: text(row.ScriptVersionID), AIRunID: text(row.AIRunID),
		Text: text(row.Text), Reason: text(row.Reason), RequestedBy: text(row.RequestedBy), DecidedBy: text(row.DecidedBy),
		SentMessageID: text(row.SentMessageID), Attempts: row.Attempts, AvailableAt: row.AvailableAt, Version: row.Version,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func ProposalRow(value types.ReplyProposal) replymodel.Proposal {
	return replymodel.Proposal{
		ID: value.ID, TenantID: value.TenantID, ThreadID: value.ThreadID, TurnID: value.TurnID, TriggerMessageID: value.TriggerMessageID,
		ControlGeneration: value.ControlGeneration, ContextRevision: value.ContextRevision, Status: value.Status,
		Route: optional(value.Route), Source: optional(value.Source), ScriptVersionID: optional(value.ScriptVersionID), AIRunID: optional(value.AIRunID),
		Text: optional(value.Text), Reason: optional(value.Reason), RequestedBy: optional(value.RequestedBy), DecidedBy: optional(value.DecidedBy),
		SentMessageID: optional(value.SentMessageID), Attempts: value.Attempts, AvailableAt: value.AvailableAt, Version: value.Version,
		CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt,
	}
}

// ProposalView maps a labelled row; staleness is a business derivation applied by the caller.
func ProposalView(row replymodel.ProposalLabel) types.ProposalView {
	view := types.ProposalView{Proposal: Proposal(row.Proposal), ScriptName: text(row.ScriptName)}
	if row.ScriptVersion != nil {
		view.ScriptVersion = *row.ScriptVersion
	}
	return view
}
