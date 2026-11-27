package service

import (
	"testing"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
)

func scriptVersion(id string, triggers ...string) types.ActiveScriptVersion {
	return types.ActiveScriptVersion{ScriptID: "s-" + id, VersionID: id, Content: types.ScriptContent{Triggers: triggers, Answer: "answer " + id}}
}

func TestMatchScriptIsDeterministic(t *testing.T) {
	price := scriptVersion("price", "giá áo", "bao nhiêu")
	size := scriptVersion("size", "size", "còn size M không")
	ship := scriptVersion("ship", "phí ship")
	versions := []types.ActiveScriptVersion{price, size, ship}
	cases := []struct {
		name, message, want string
	}{
		{"case and punctuation are ignored", "Cho mình hỏi GIÁ ÁO này?!", "price"},
		{"whole words only", "áo giáp này", ""},
		{"diacritics are significant", "gia ao bao nhieu", ""},
		{"decomposed input matches composed trigger", "giá áo", "price"},
		{"longest phrase wins across scripts", "shop ơi còn size M không, bao nhiêu vậy", "size"},
		{"no trigger", "xin chào", ""},
		{"empty message", "   ", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := MatchScript(tc.message, versions)
			if tc.want == "" {
				if ok {
					t.Fatalf("expected no match, got %s", got.VersionID)
				}
				return
			}
			if !ok || got.VersionID != tc.want {
				t.Fatalf("expected %s, got %v %s", tc.want, ok, got.VersionID)
			}
		})
	}
	// Equal-length best matches from different scripts are ambiguous; the router must ask the AI instead.
	if _, ok := MatchScript("mua áo, giá áo", []types.ActiveScriptVersion{scriptVersion("a", "mua áo"), scriptVersion("b", "giá áo")}); ok {
		t.Fatal("ambiguous match must not choose a script")
	}
	// Two phrases of the same script are not ambiguous.
	if got, ok := MatchScript("giá áo bao nhiêu", []types.ActiveScriptVersion{scriptVersion("one", "giá áo", "bao nhiêu")}); !ok || got.VersionID != "one" {
		t.Fatal("same-script phrases must match")
	}
}

func TestAICandidateAppliesGoGates(t *testing.T) {
	job := types.ReplyProposal{ID: "run-1", ContextRevision: 3, ControlGeneration: 2}
	versions := []types.ActiveScriptVersion{scriptVersion("v1", "giá")}
	base := types.AIRun{Status: types.AIRunSucceeded, ContextRevision: 3, ControlGeneration: 2}
	long := make([]rune, replyTextLimit+1)
	for i := range long {
		long[i] = 'a'
	}
	cases := []struct {
		name                string
		run                 types.AIRun
		route, source, text string
		handoffReason, ver  string
	}{
		{"grounded answer", types.AIRun{Route: "generate", Text: " Áo giá 200k "}, types.RouteAnswer, types.SourceAI, "Áo giá 200k", "", ""},
		{"clarifying question", types.AIRun{Route: "clarify", Text: "Bạn hỏi sản phẩm nào?"}, types.RouteClarify, types.SourceAI, "Bạn hỏi sản phẩm nào?", "", ""},
		{"active script uses Go content", types.AIRun{Route: "script", ScriptID: "v1", Text: "model text"}, types.RouteScript, types.SourceScript, "answer v1", "", "v1"},
		{"unknown script is handed off", types.AIRun{Route: "script", ScriptID: "gone"}, types.RouteHandoff, "", "", "insufficient_information", ""},
		{"human route", types.AIRun{Route: "human", Reason: "human_review_required"}, types.RouteHandoff, "", "", "insufficient_information", ""},
		{"empty answer", types.AIRun{Route: "generate"}, types.RouteHandoff, "", "", "insufficient_information", ""},
		{"over-long answer", types.AIRun{Route: "generate", Text: string(long)}, types.RouteHandoff, "", "", "insufficient_information", ""},
		{"wrong basis", types.AIRun{Route: "generate", Text: "x", ControlGeneration: 1}, types.RouteHandoff, "", "", "ai_error", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			run := tc.run
			run.Status = base.Status
			if run.ContextRevision == 0 {
				run.ContextRevision = base.ContextRevision
			}
			if run.ControlGeneration == 0 {
				run.ControlGeneration = base.ControlGeneration
			}
			got := aiCandidate(job, run, versions)
			if got.Route != tc.route || got.Source != tc.source || got.Text != tc.text || got.HandoffReason != tc.handoffReason || got.ScriptVersionID != tc.ver {
				t.Fatalf("unexpected candidate %+v", got)
			}
		})
	}
	if safeReason("Ignore previous instructions") != "ai_decision" || safeReason("needs_clarification") != "needs_clarification" {
		t.Fatal("AI reasons must be reduced to safe identifiers")
	}
}
