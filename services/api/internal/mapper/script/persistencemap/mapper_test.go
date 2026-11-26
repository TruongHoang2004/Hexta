package persistencemap

import (
	"errors"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/scriptmodel"
)

func TestVersionRejectsUnsupportedStoredTriggers(t *testing.T) {
	for _, raw := range []string{`{"a":1}`, `null`, `[1,2]`, `not json`} {
		if _, err := Version(scriptmodel.Version{Triggers: raw}); !errors.Is(err, ErrStoredTriggers) {
			t.Fatalf("%s: expected ErrStoredTriggers, got %v", raw, err)
		}
	}
}

func TestVersionRoundTripKeepsDraftStateAndEmptyTriggers(t *testing.T) {
	at := time.Date(2026, 11, 26, 10, 0, 0, 123456000, time.UTC)
	published := types.ScriptVersion{ID: "v", Number: 3, Content: types.ScriptContent{Triggers: []string{"giá áo"}, Answer: "a"}, PublishedAt: &at, PublishedBy: "owner"}
	back, err := Version(VersionRow(published))
	if err != nil || back.Draft() || !back.PublishedAt.Equal(at) || back.PublishedBy != "owner" || back.Content.Triggers[0] != "giá áo" {
		t.Fatalf("round trip lost published state: %+v %v", back, err)
	}
	draft, err := Version(VersionRow(types.ScriptVersion{ID: "d"}))
	if err != nil || !draft.Draft() || draft.Content.Triggers == nil || len(draft.Content.Triggers) != 0 {
		t.Fatalf("nil triggers must persist as an empty array: %+v %v", draft, err)
	}
}

func TestMemberWithoutGrantRowIsEmptySetAtVersionZero(t *testing.T) {
	m := Member("t", "owner", scriptmodel.MemberRow{UserID: "u", Role: "member"}, "u@example.test")
	if m.Owner || m.Grants.Version != 0 || len(m.Grants.Granted()) != 0 || m.Email != "u@example.test" {
		t.Fatalf("unexpected member mapping: %+v", m)
	}
}
