// Package httpmap converts reply settings and reply proposals to HTTP DTOs.
package httpmap

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
)

func Settings(value types.ReplySettings) dto.ReplySettingsResponse {
	result := dto.ReplySettingsResponse{ReplyMode: value.Mode, Configured: value.Configured, Version: value.Version}
	if value.Configured {
		at := value.UpdatedAt
		result.UpdatedAt = &at
	}
	return result
}

// Proposal maps a proposal; routing jobs and terminal non-candidates carry no text.
func Proposal(value types.ProposalView) dto.ReplyProposal {
	p := value.Proposal
	result := dto.ReplyProposal{
		ID: p.ID, TurnID: p.TurnID, Status: p.Status, Route: p.Route, Source: p.Source, ScriptVersionID: p.ScriptVersionID,
		ScriptName: value.ScriptName, ScriptVersion: value.ScriptVersion, AIRunID: p.AIRunID, Text: p.Text, Reason: p.Reason,
		ControlGeneration: p.ControlGeneration, Stale: value.Stale, RequestedBy: p.RequestedBy, CreatedAt: p.CreatedAt,
	}
	for _, citation := range value.Citations {
		result.Citations = append(result.Citations, dto.ProposalCitation{SourceID: citation.SourceID, Kind: citation.Kind, Label: citation.Label})
	}
	return result
}

func Proposals(values []types.ProposalView) []dto.ReplyProposal {
	result := make([]dto.ReplyProposal, len(values))
	for i, value := range values {
		result[i] = Proposal(value)
	}
	return result
}
