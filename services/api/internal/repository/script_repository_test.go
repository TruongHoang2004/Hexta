package repository

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type scriptRepoFixture struct {
	db            *gorm.DB
	repo          *ScriptRepository
	tenant, other string
}

func newScriptRepoFixture(t *testing.T) scriptRepoFixture {
	t.Helper()
	db := testutil.Database(t)
	f := scriptRepoFixture{db: db, repo: NewScriptRepository(db), tenant: uuid.NewString(), other: uuid.NewString()}
	for _, id := range []string{f.tenant, f.other} {
		require.NoError(t, db.Create(&model.Tenant{ID: id, Name: id, Slug: id, OwnerID: "owner", Status: model.TenantStatusActive}).Error)
		require.NoError(t, db.Create(&model.TenantMember{TenantID: id, UserID: "owner", Role: model.RoleOwner}).Error)
	}
	require.NoError(t, db.Create(&model.TenantMember{TenantID: f.tenant, UserID: "member", Role: model.RoleMember}).Error)
	return f
}

func (f scriptRepoFixture) script(t *testing.T, tenant string) (types.Script, types.ScriptVersion) {
	t.Helper()
	ctx, now := context.Background(), time.Now().UTC()
	script := types.Script{ID: uuid.NewString(), TenantID: tenant, Name: "Price", Status: types.ScriptStatusInactive, Version: 1, CreateHash: "h", CreatedBy: "owner", CreatedAt: now, UpdatedAt: now}
	inserted, err := f.repo.InsertScript(ctx, script)
	require.Nil(t, err)
	require.True(t, inserted)
	draft := types.ScriptVersion{ID: uuid.NewString(), TenantID: tenant, ScriptID: script.ID, Number: 1, Content: types.ScriptContent{Triggers: []string{"price"}, Answer: "200k"}, CreatedBy: "owner", UpdatedBy: "owner", CreatedAt: now, UpdatedAt: now}
	require.Nil(t, f.repo.InsertVersion(ctx, draft))
	return script, draft
}

func TestScriptPublishedVersionContentIsImmutable(t *testing.T) {
	f := newScriptRepoFixture(t)
	ctx := context.Background()
	script, draft := f.script(t, f.tenant)
	draft.Content = types.ScriptContent{Triggers: []string{"price", "cost"}, Answer: "190k"}
	require.Nil(t, f.repo.UpdateDraft(ctx, draft))
	require.Nil(t, f.repo.PublishVersion(ctx, f.tenant, script.ID, draft.ID, "owner", time.Now().UTC()))

	draft.Content.Answer = "tampered"
	err := f.repo.UpdateDraft(ctx, draft)
	require.NotNil(t, err)
	require.Equal(t, http.StatusConflict, err.GetHttpStatus())
	stored, appErr := f.repo.Version(ctx, f.tenant, script.ID, draft.ID)
	require.Nil(t, appErr)
	require.Equal(t, "190k", stored.Content.Answer)
	require.Equal(t, []string{"price", "cost"}, stored.Content.Triggers)
	require.False(t, stored.Draft())

	// Re-publishing keeps the original publication stamp.
	first := *stored.PublishedAt
	require.Nil(t, f.repo.PublishVersion(ctx, f.tenant, script.ID, draft.ID, "other", time.Now().UTC().Add(time.Hour)))
	stored, _ = f.repo.Version(ctx, f.tenant, script.ID, draft.ID)
	require.WithinDuration(t, first, *stored.PublishedAt, time.Millisecond)
	require.Equal(t, "owner", stored.PublishedBy)
}

func TestScriptAllowsOnlyOneDraftAndUniqueNumbers(t *testing.T) {
	f := newScriptRepoFixture(t)
	ctx, now := context.Background(), time.Now().UTC()
	script, _ := f.script(t, f.tenant)
	second := types.ScriptVersion{ID: uuid.NewString(), TenantID: f.tenant, ScriptID: script.ID, Number: 2, Content: types.ScriptContent{Triggers: []string{"x"}, Answer: "x"}, CreatedBy: "owner", UpdatedBy: "owner", CreatedAt: now, UpdatedAt: now}
	err := f.repo.InsertVersion(ctx, second)
	require.NotNil(t, err, "a second unpublished draft must be rejected")
	require.Equal(t, http.StatusConflict, err.GetHttpStatus())

	published := second
	published.PublishedAt, published.PublishedBy = &now, "owner"
	published.Number = 1
	err = f.repo.InsertVersion(ctx, published)
	require.NotNil(t, err, "duplicate version numbers must be rejected")
	next, appErr := f.repo.NextNumber(ctx, f.tenant, script.ID)
	require.Nil(t, appErr)
	require.EqualValues(t, 2, next)
}

func TestScriptActivePointerMustNameOwnVersion(t *testing.T) {
	f := newScriptRepoFixture(t)
	ctx := context.Background()
	script, _ := f.script(t, f.tenant)
	_, foreign := f.script(t, f.tenant)
	script.Status, script.ActiveVersionID, script.Version = types.ScriptStatusActive, foreign.ID, 2
	require.NotNil(t, f.repo.SaveScript(ctx, script, 1), "pointer to another script's version must fail the composite FK")

	// Status and pointer must agree.
	require.Error(t, f.db.Exec(`UPDATE scripts SET status='active' WHERE id=?`, script.ID).Error)
}

func TestScriptSaveRejectsStaleVersionAndCrossTenantAccess(t *testing.T) {
	f := newScriptRepoFixture(t)
	ctx := context.Background()
	script, draft := f.script(t, f.tenant)
	script.Name, script.Version = "Renamed", 2
	require.Nil(t, f.repo.SaveScript(ctx, script, 1))
	script.Version = 3
	err := f.repo.SaveScript(ctx, script, 1)
	require.NotNil(t, err)
	require.Equal(t, http.StatusConflict, err.GetHttpStatus())

	_, err = f.repo.Script(ctx, f.other, script.ID, false)
	require.NotNil(t, err)
	require.Equal(t, http.StatusNotFound, err.GetHttpStatus())
	_, err = f.repo.Version(ctx, f.other, script.ID, draft.ID)
	require.NotNil(t, err)
	foreign := draft
	foreign.TenantID = f.other
	require.NotNil(t, f.repo.UpdateDraft(ctx, foreign), "an update scoped to another tenant must not touch the row")

	rows, total, err := f.repo.Scripts(ctx, f.tenant, types.ScriptListFilter{}, 20, 0)
	require.Nil(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, "Renamed", rows[0].Name)
	require.EqualValues(t, 1, rows[0].DraftNumber)
	rows, _, _ = f.repo.Scripts(ctx, f.other, types.ScriptListFilter{}, 20, 0)
	require.Empty(t, rows)
}

func TestScriptActiveVersionsProjection(t *testing.T) {
	f := newScriptRepoFixture(t)
	ctx, now := context.Background(), time.Now().UTC()
	script, draft := f.script(t, f.tenant)
	f.script(t, f.tenant) // inactive draft is never offered
	require.Nil(t, f.repo.PublishVersion(ctx, f.tenant, script.ID, draft.ID, "owner", now))
	script.Status, script.ActiveVersionID, script.Version = types.ScriptStatusActive, draft.ID, 2
	require.Nil(t, f.repo.SaveScript(ctx, script, 1))

	rows, err := f.repo.ActiveVersions(ctx, f.tenant, 500)
	require.Nil(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, draft.ID, rows[0].VersionID)
	require.Equal(t, []string{"price"}, rows[0].Content.Triggers)

	require.NoError(t, f.db.Exec(`UPDATE tenants SET status=? WHERE id=?`, model.TenantStatusSuspended, f.tenant).Error)
	rows, err = f.repo.ActiveVersions(ctx, f.tenant, 500)
	require.Nil(t, err)
	require.Empty(t, rows, "a suspended tenant must not auto-answer")
}

func TestScriptPermissionsFollowMembership(t *testing.T) {
	f := newScriptRepoFixture(t)
	ctx, now := context.Background(), time.Now().UTC()
	set, err := f.repo.PermissionSet(ctx, f.tenant, "member")
	require.Nil(t, err)
	require.Zero(t, set.Version)
	set = set.With(types.ScriptPermissionEdit, true)
	set.Version, set.UpdatedBy, set.UpdatedAt = 1, "owner", now
	require.Nil(t, f.repo.SavePermissionSet(ctx, set, 0))
	err = f.repo.SavePermissionSet(ctx, set, 0)
	require.NotNil(t, err, "a concurrent first grant must conflict")
	require.Equal(t, http.StatusConflict, err.GetHttpStatus())

	access, err := f.repo.Access(ctx, f.tenant, "member")
	require.Nil(t, err)
	require.True(t, access.Allows(types.ScriptPermissionEdit))
	require.False(t, access.Allows(types.ScriptPermissionActivate))
	other, err := f.repo.Access(ctx, f.other, "member")
	require.Nil(t, err)
	require.False(t, other.Allows(types.ScriptPermissionEdit), "grants are tenant scoped")

	members, err := f.repo.MemberPermissions(ctx, f.tenant, "owner", 500)
	require.Nil(t, err)
	require.Len(t, members, 2)

	// A non-member cannot hold grants, and removing a membership removes its grants.
	outsider := types.ScriptPermissionSet{TenantID: f.tenant, UserID: "outsider", Edit: true, Version: 1, UpdatedBy: "owner", UpdatedAt: now}
	require.NotNil(t, f.repo.SavePermissionSet(ctx, outsider, 0))
	require.NoError(t, f.db.Exec(`DELETE FROM tenant_members WHERE tenant_id=? AND user_id='member'`, f.tenant).Error)
	set, err = f.repo.PermissionSet(ctx, f.tenant, "member")
	require.Nil(t, err)
	require.Zero(t, set.Version)
}
