package service

import (
	"context"
	"net/http"
	"testing"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
)

// TestScriptLifecycleAgainstPostgres runs the full lifecycle and the acceptance scenario
// ("a member with only script.edit cannot activate") through the real repository and Atlas schema.
func TestScriptLifecycleAgainstPostgres(t *testing.T) {
	f := salesTestFixture(t)
	ctx := context.Background()
	s := NewScriptService(repository.NewScriptRepository(f.db), repository.NewTenantDBRepository(f.db))

	if _, err := s.Grant(ctx, f.tenant, f.owner, f.sale, types.ScriptPermissionEdit, 0); err != nil {
		t.Fatal(err)
	}
	script, err := s.Create(ctx, f.tenant, f.owner, uuid.NewString(), types.ScriptCreateInput{Name: "Shirt price", Content: types.ScriptContent{Triggers: []string{"shirt price"}, Answer: "200k"}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.Create(ctx, f.tenant, f.sale, uuid.NewString(), types.ScriptCreateInput{Name: "x", Content: types.ScriptContent{Triggers: []string{"x"}, Answer: "x"}})
	requireSalesStatus(t, err, http.StatusForbidden)
	v1 := script.Versions[0]
	_, err = s.Activate(ctx, f.tenant, f.sale, script.ID, v1.ID, script.Version)
	requireSalesStatus(t, err, http.StatusForbidden)

	active, err := s.Activate(ctx, f.tenant, f.owner, script.ID, v1.ID, script.Version)
	if err != nil {
		t.Fatal(err)
	}
	edited, err := s.SaveDraft(ctx, f.tenant, f.sale, script.ID, types.ScriptDraftInput{Name: "Shirt price", Content: types.ScriptContent{Triggers: []string{"shirt price"}, Answer: "180k"}, ExpectedVersion: active.Version})
	if err != nil {
		t.Fatal(err)
	}
	if len(edited.Versions) != 2 || edited.Versions[0].Number != 2 || !edited.Versions[0].Draft() || edited.ActiveVersionID != v1.ID {
		t.Fatalf("editing an active script must open v2: %+v", edited)
	}
	rows, err := s.ActiveScriptVersions(ctx, f.tenant)
	if err != nil || len(rows) != 1 || rows[0].Content.Answer != "200k" {
		t.Fatalf("the reply router must still see published v1: %+v %v", rows, err)
	}
	_, err = s.SaveDraft(ctx, f.tenant, f.sale, script.ID, types.ScriptDraftInput{Name: "Shirt price", Content: types.ScriptContent{Triggers: []string{"shirt price"}, Answer: "stale"}, ExpectedVersion: active.Version})
	requireSalesStatus(t, err, http.StatusConflict)

	archived, err := s.Archive(ctx, f.tenant, f.owner, script.ID, edited.Version)
	if err != nil {
		t.Fatal(err)
	}
	if archived.Status != types.ScriptStatusArchived || len(archived.Versions) != 2 {
		t.Fatalf("archive must retain history: %+v", archived)
	}
	if rows, _ := s.ActiveScriptVersions(ctx, f.tenant); len(rows) != 0 {
		t.Fatal("archived scripts must not be offered")
	}

	members, err := s.MemberPermissions(ctx, f.tenant, f.owner)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range members {
		if m.UserID == f.sale {
			found = m.Grants.Has(types.ScriptPermissionEdit) && !m.Grants.Has(types.ScriptPermissionActivate)
		}
	}
	if !found {
		t.Fatalf("permission list must show the sale's single grant: %+v", members)
	}
	var audits int64
	if err := f.db.Raw(`SELECT count(*) FROM audit_logs WHERE tenant_id = ? AND action LIKE 'script.%'`, f.tenant).Scan(&audits).Error; err != nil || audits < 5 {
		t.Fatalf("script commands must be audited: %d %v", audits, err)
	}
}
