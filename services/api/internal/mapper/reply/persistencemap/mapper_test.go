package persistencemap

import (
	"errors"
	"reflect"
	"testing"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/replymodel"
)

func TestSettingsRejectUnknownStoredMode(t *testing.T) {
	if _, err := Settings(replymodel.Settings{ReplyMode: "always"}); !errors.Is(err, ErrStoredMode) {
		t.Fatal("an unknown stored mode must never be read as a send permission")
	}
	value, err := Settings(replymodel.Settings{TenantID: "t", ReplyMode: types.ReplyModeAutoSend, Version: 3})
	if err != nil || !value.Configured || value.Mode != types.ReplyModeAutoSend || value.Version != 3 {
		t.Fatal(value, err)
	}
}

func TestProposalRoundTripKeepsMissingSourcesNull(t *testing.T) {
	routing := types.ReplyProposal{ID: "p", Status: types.ProposalRouting, ControlGeneration: 2, ContextRevision: 5, Version: 1}
	row := ProposalRow(routing)
	if row.Source != nil || row.ScriptVersionID != nil || row.AIRunID != nil || row.Text != nil || row.SentMessageID != nil {
		t.Fatal("empty values must be stored as NULL so the source checks hold", row)
	}
	if row.CitedSourceIDs != "[]" {
		t.Fatalf("no citations must be stored as an empty JSON array: %q", row.CitedSourceIDs)
	}
	if back := Proposal(row); !reflect.DeepEqual(back, routing) {
		t.Fatalf("round trip changed the proposal: %+v", back)
	}
	version := int64(3)
	name := "Báo giá áo"
	view := ProposalView(replymodel.ProposalLabel{Proposal: ProposalRow(types.ReplyProposal{ID: "p", Source: types.SourceScript, ScriptVersionID: "v"}), ScriptName: &name, ScriptVersion: &version})
	if view.ScriptName != name || view.ScriptVersion != 3 || view.Stale {
		t.Fatal(view)
	}
}

func TestProposalCitationsRoundTripAndTolerateBadStoredValues(t *testing.T) {
	cited := types.ReplyProposal{ID: "p", CitedSourceIDs: []string{"product:a", "document:b"}}
	row := ProposalRow(cited)
	if row.CitedSourceIDs != `["product:a","document:b"]` {
		t.Fatal(row.CitedSourceIDs)
	}
	if back := Proposal(row); !reflect.DeepEqual(back.CitedSourceIDs, cited.CitedSourceIDs) {
		t.Fatal(back.CitedSourceIDs)
	}
	for _, stored := range []string{"", "{}", "not json", "[]"} {
		if got := Proposal(replymodel.Proposal{CitedSourceIDs: stored}).CitedSourceIDs; got != nil {
			t.Fatalf("%q read as %v", stored, got)
		}
	}
}
