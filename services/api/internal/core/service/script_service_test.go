package service

import (
	"context"
	"net/http"
	"sort"
	"testing"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
)

// fakeScriptStore is an in-memory ScriptStore that mirrors the repository's persistence guards (versioned saves,
// one draft per script, frozen published content) and rolls back on callback failure.
type fakeScriptStore struct {
	owner    map[string]string
	active   map[string]bool
	members  map[string]bool
	grants   map[string]types.ScriptPermissionSet
	scripts  map[string]types.Script
	versions map[string]types.ScriptVersion
	audits   []string
	// synced lists knowledge sync enqueues ("type:id") in commit order; rolled back with the transaction.
	synced []string
}

func newFakeScriptStore() *fakeScriptStore {
	return &fakeScriptStore{owner: map[string]string{}, active: map[string]bool{}, members: map[string]bool{}, grants: map[string]types.ScriptPermissionSet{}, scripts: map[string]types.Script{}, versions: map[string]types.ScriptVersion{}}
}

func memberKey(tenant, user string) string { return tenant + "|" + user }

func (f *fakeScriptStore) snapshot() *fakeScriptStore {
	c := newFakeScriptStore()
	for k, v := range f.owner {
		c.owner[k] = v
	}
	for k, v := range f.active {
		c.active[k] = v
	}
	for k, v := range f.members {
		c.members[k] = v
	}
	for k, v := range f.grants {
		c.grants[k] = v
	}
	for k, v := range f.scripts {
		c.scripts[k] = v
	}
	for k, v := range f.versions {
		c.versions[k] = v
	}
	c.audits = append([]string{}, f.audits...)
	c.synced = append([]string{}, f.synced...)
	return c
}

func (f *fakeScriptStore) Within(_ context.Context, fn func(repository.ScriptStore) *apperrors.Error) *apperrors.Error {
	saved := f.snapshot()
	if err := fn(f); err != nil {
		*f = *saved
		return err
	}
	return nil
}

func (f *fakeScriptStore) Access(_ context.Context, tenant, actor string) (types.ScriptAccess, *apperrors.Error) {
	return types.ScriptAccess{ActiveTenant: f.active[tenant], Member: f.members[memberKey(tenant, actor)], Owner: f.owner[tenant] == actor, Grants: f.grants[memberKey(tenant, actor)]}, nil
}

func (f *fakeScriptStore) Scripts(_ context.Context, tenant string, filter types.ScriptListFilter, limit, offset int) ([]types.ScriptSummary, int64, *apperrors.Error) {
	rows := []types.ScriptSummary{}
	for _, s := range f.scripts {
		if s.TenantID != tenant || (filter.Status != "" && s.Status != filter.Status) {
			continue
		}
		row := types.ScriptSummary{Script: s}
		for _, v := range f.versions {
			if v.ScriptID == s.ID && v.ID == s.ActiveVersionID {
				row.ActiveNumber = v.Number
			}
			if v.ScriptID == s.ID && v.Draft() {
				row.DraftNumber = v.Number
			}
		}
		rows = append(rows, row)
	}
	return rows, int64(len(rows)), nil
}

func (f *fakeScriptStore) Script(ctx context.Context, tenant, id string, _ bool) (*types.Script, *apperrors.Error) {
	s, ok := f.scripts[id]
	if !ok || s.TenantID != tenant {
		return nil, apperrors.ErrNotFound(ctx, "script", "not found")
	}
	return &s, nil
}

func (f *fakeScriptStore) Versions(_ context.Context, tenant, script string, limit int) ([]types.ScriptVersion, *apperrors.Error) {
	rows := []types.ScriptVersion{}
	for _, v := range f.versions {
		if v.TenantID == tenant && v.ScriptID == script {
			rows = append(rows, v)
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Number > rows[j].Number })
	if len(rows) > limit {
		rows = rows[:limit]
	}
	return rows, nil
}

func (f *fakeScriptStore) Version(ctx context.Context, tenant, script, id string) (*types.ScriptVersion, *apperrors.Error) {
	v, ok := f.versions[id]
	if !ok || v.TenantID != tenant || v.ScriptID != script {
		return nil, apperrors.ErrNotFound(ctx, "script version", "not found")
	}
	return &v, nil
}

func (f *fakeScriptStore) Draft(_ context.Context, tenant, script string) (*types.ScriptVersion, *apperrors.Error) {
	for _, v := range f.versions {
		if v.TenantID == tenant && v.ScriptID == script && v.Draft() {
			return &v, nil
		}
	}
	return nil, nil
}

func (f *fakeScriptStore) NextNumber(_ context.Context, tenant, script string) (int64, *apperrors.Error) {
	var max int64
	for _, v := range f.versions {
		if v.TenantID == tenant && v.ScriptID == script && v.Number > max {
			max = v.Number
		}
	}
	return max + 1, nil
}

func (f *fakeScriptStore) InsertScript(_ context.Context, s types.Script) (bool, *apperrors.Error) {
	if _, exists := f.scripts[s.ID]; exists {
		return false, nil
	}
	f.scripts[s.ID] = s
	return true, nil
}

func (f *fakeScriptStore) InsertVersion(ctx context.Context, v types.ScriptVersion) *apperrors.Error {
	for _, existing := range f.versions {
		if existing.ScriptID == v.ScriptID && (existing.Number == v.Number || (existing.Draft() && v.Draft())) {
			return apperrors.ErrConflict(ctx, "script", "changed; refresh before retrying")
		}
	}
	f.versions[v.ID] = v
	return nil
}

func (f *fakeScriptStore) UpdateDraft(ctx context.Context, v types.ScriptVersion) *apperrors.Error {
	current, ok := f.versions[v.ID]
	if !ok || !current.Draft() {
		return apperrors.ErrConflict(ctx, "script version", "is published and cannot be changed")
	}
	current.Content, current.UpdatedBy, current.UpdatedAt = v.Content, v.UpdatedBy, v.UpdatedAt
	f.versions[v.ID] = current
	return nil
}

func (f *fakeScriptStore) PublishVersion(_ context.Context, _, _, id, actor string, at time.Time) *apperrors.Error {
	if v, ok := f.versions[id]; ok && v.Draft() {
		v.PublishedAt, v.PublishedBy = &at, actor
		f.versions[id] = v
	}
	return nil
}

func (f *fakeScriptStore) SaveScript(ctx context.Context, s types.Script, expected int64) *apperrors.Error {
	current, ok := f.scripts[s.ID]
	if !ok || current.Version != expected {
		return apperrors.ErrConflict(ctx, "script", "changed; refresh before retrying")
	}
	f.scripts[s.ID] = s
	return nil
}

func (f *fakeScriptStore) ActiveVersions(_ context.Context, tenant string, _ int) ([]types.ActiveScriptVersion, *apperrors.Error) {
	rows := []types.ActiveScriptVersion{}
	for _, s := range f.scripts {
		if s.TenantID == tenant && s.Status == types.ScriptStatusActive && f.active[tenant] {
			v := f.versions[s.ActiveVersionID]
			rows = append(rows, types.ActiveScriptVersion{ScriptID: s.ID, ScriptName: s.Name, VersionID: v.ID, Number: v.Number, Content: v.Content, PublishedAt: *v.PublishedAt})
		}
	}
	return rows, nil
}

func (f *fakeScriptStore) MemberPermissions(_ context.Context, tenant, owner string, _ int) ([]types.ScriptMemberPermissions, *apperrors.Error) {
	rows := []types.ScriptMemberPermissions{}
	for key := range f.members {
		if t, user := key[:len(tenant)], key[len(tenant)+1:]; t == tenant {
			rows = append(rows, types.ScriptMemberPermissions{UserID: user, Owner: user == owner, Grants: f.grants[key]})
		}
	}
	return rows, nil
}

func (f *fakeScriptStore) IsMember(_ context.Context, tenant, user string) (bool, *apperrors.Error) {
	return f.members[memberKey(tenant, user)], nil
}

func (f *fakeScriptStore) PermissionSet(_ context.Context, tenant, user string) (types.ScriptPermissionSet, *apperrors.Error) {
	set, ok := f.grants[memberKey(tenant, user)]
	if !ok {
		return types.ScriptPermissionSet{TenantID: tenant, UserID: user}, nil
	}
	return set, nil
}

func (f *fakeScriptStore) SavePermissionSet(ctx context.Context, set types.ScriptPermissionSet, expected int64) *apperrors.Error {
	if f.grants[memberKey(set.TenantID, set.UserID)].Version != expected {
		return apperrors.ErrConflict(ctx, "script permissions", "were updated elsewhere; reload to see them")
	}
	f.grants[memberKey(set.TenantID, set.UserID)] = set
	return nil
}

func (f *fakeScriptStore) Audit(_ context.Context, _, _, action, _, _ string, _ any) *apperrors.Error {
	f.audits = append(f.audits, action)
	return nil
}

func (f *fakeScriptStore) EnqueueKnowledgeSync(_ context.Context, _, sourceType, sourceID string, _ time.Time) *apperrors.Error {
	f.synced = append(f.synced, sourceType+":"+sourceID)
	return nil
}

type scriptFixture struct {
	store                 *fakeScriptStore
	service               *ScriptService
	tenant, owner, member string
	other, outsider, peer string
}

func newScriptFixture(t *testing.T) scriptFixture {
	t.Helper()
	f := scriptFixture{store: newFakeScriptStore(), tenant: uuid.NewString(), other: uuid.NewString(), owner: "owner", member: "member", outsider: "outsider", peer: "member2"}
	for _, tenant := range []string{f.tenant, f.other} {
		f.store.owner[tenant], f.store.active[tenant] = f.owner, true
		f.store.members[memberKey(tenant, f.owner)] = true
	}
	f.store.members[memberKey(f.tenant, f.member)] = true
	f.store.members[memberKey(f.tenant, f.peer)] = true
	f.service = NewScriptService(f.store, nil)
	return f
}

// grantOnly gives the member exactly one permission through the owner-only grant command.
func (f scriptFixture) grantOnly(t *testing.T, user, permission string) {
	t.Helper()
	current, _ := f.store.PermissionSet(context.Background(), f.tenant, user)
	if _, err := f.service.Grant(context.Background(), f.tenant, f.owner, user, permission, current.Version); err != nil {
		t.Fatal(err)
	}
}

func scriptInput(answer string) types.ScriptCreateInput {
	return types.ScriptCreateInput{Name: "Shirt price", Content: types.ScriptContent{Triggers: []string{"shirt price", "how much"}, Answer: answer}}
}

func (f scriptFixture) create(t *testing.T) *types.ScriptDetail {
	t.Helper()
	script, err := f.service.Create(context.Background(), f.tenant, f.owner, uuid.NewString(), scriptInput("It costs 200k."))
	if err != nil {
		t.Fatal(err)
	}
	return script
}

func requireScriptStatus(t *testing.T, err *apperrors.Error, status int) {
	t.Helper()
	if err == nil || err.GetHttpStatus() != status {
		t.Fatalf("expected status %d, got %v", status, err)
	}
}

func draftOf(t *testing.T, script *types.ScriptDetail) types.ScriptVersion {
	t.Helper()
	for _, v := range script.Versions {
		if v.Draft() {
			return v
		}
	}
	t.Fatal("script has no draft")
	return types.ScriptVersion{}
}

func TestScriptEachPermissionGrantsOnlyItsOwnAction(t *testing.T) {
	ctx := context.Background()
	for _, permission := range types.ScriptPermissions {
		t.Run(permission, func(t *testing.T) {
			f := newScriptFixture(t)
			f.grantOnly(t, f.member, permission)
			allowed := map[string]bool{permission: true}
			attempts := map[string]func() *apperrors.Error{
				types.ScriptPermissionCreate: func() *apperrors.Error {
					_, err := f.service.Create(ctx, f.tenant, f.member, uuid.NewString(), scriptInput("Hello"))
					return err
				},
				types.ScriptPermissionEdit: func() *apperrors.Error {
					script := f.create(t)
					_, err := f.service.SaveDraft(ctx, f.tenant, f.member, script.ID, types.ScriptDraftInput{Name: "Renamed", Content: types.ScriptContent{Triggers: []string{"price"}, Answer: "New"}, ExpectedVersion: script.Version})
					return err
				},
				types.ScriptPermissionActivate: func() *apperrors.Error {
					script := f.create(t)
					_, err := f.service.Activate(ctx, f.tenant, f.member, script.ID, draftOf(t, script).ID, script.Version)
					return err
				},
				types.ScriptPermissionArchive: func() *apperrors.Error {
					script := f.create(t)
					_, err := f.service.Archive(ctx, f.tenant, f.member, script.ID, script.Version)
					return err
				},
			}
			for action, attempt := range attempts {
				err := attempt()
				if allowed[action] && err != nil {
					t.Fatalf("%s should allow %s: %v", permission, action, err)
				}
				if !allowed[action] {
					requireScriptStatus(t, err, http.StatusForbidden)
				}
			}
		})
	}
}

func TestScriptMemberWithOnlyEditCannotActivateOrDeactivate(t *testing.T) {
	f := newScriptFixture(t)
	ctx := context.Background()
	f.grantOnly(t, f.member, types.ScriptPermissionEdit)
	script := f.create(t)
	_, err := f.service.Activate(ctx, f.tenant, f.member, script.ID, draftOf(t, script).ID, script.Version)
	requireScriptStatus(t, err, http.StatusForbidden)
	active, err := f.service.Activate(ctx, f.tenant, f.owner, script.ID, draftOf(t, script).ID, script.Version)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.service.Deactivate(ctx, f.tenant, f.member, script.ID, active.Version)
	requireScriptStatus(t, err, http.StatusForbidden)
	if got := f.store.scripts[script.ID]; got.Status != types.ScriptStatusActive || got.Version != active.Version {
		t.Fatalf("refused command changed the script: %+v", got)
	}
}

func TestScriptReadsNeedMembershipAndManagementNeedsOwner(t *testing.T) {
	f := newScriptFixture(t)
	ctx := context.Background()
	script := f.create(t)
	if _, err := f.service.Get(ctx, f.tenant, f.member, script.ID); err != nil {
		t.Fatalf("members without grants can read scripts: %v", err)
	}
	_, _, err := f.service.List(ctx, f.tenant, f.outsider, types.ScriptListFilter{}, 1, 20)
	requireScriptStatus(t, err, http.StatusForbidden)
	_, err = f.service.Get(ctx, f.other, f.owner, script.ID)
	requireScriptStatus(t, err, http.StatusNotFound)
	_, err = f.service.MemberPermissions(ctx, f.tenant, f.member)
	requireScriptStatus(t, err, http.StatusForbidden)
	_, err = f.service.Create(ctx, f.tenant, "", uuid.NewString(), scriptInput("x"))
	requireScriptStatus(t, err, http.StatusUnauthorized)

	// A suspended workspace refuses even its owner.
	f.store.active[f.tenant] = false
	_, err = f.service.Create(ctx, f.tenant, f.owner, uuid.NewString(), scriptInput("x"))
	requireScriptStatus(t, err, http.StatusForbidden)
}

func TestScriptEditingActiveScriptCreatesNewDraftAndKeepsPublishedContent(t *testing.T) {
	f := newScriptFixture(t)
	ctx := context.Background()
	script := f.create(t)
	if script.Status != types.ScriptStatusInactive || len(script.Versions) != 1 || !script.Versions[0].Draft() || script.Versions[0].Number != 1 {
		t.Fatalf("new script must be inactive with draft v1: %+v", script)
	}
	v1 := script.Versions[0]
	active, err := f.service.Activate(ctx, f.tenant, f.owner, script.ID, v1.ID, script.Version)
	if err != nil {
		t.Fatal(err)
	}
	if active.Status != types.ScriptStatusActive || active.ActiveVersionID != v1.ID || active.Versions[0].Draft() {
		t.Fatalf("activation must publish v1 and point at it: %+v", active)
	}

	edited, err := f.service.SaveDraft(ctx, f.tenant, f.owner, script.ID, types.ScriptDraftInput{Name: "Shirt price", Content: types.ScriptContent{Triggers: []string{"shirt price"}, Answer: "Now 180k."}, ExpectedVersion: active.Version})
	if err != nil {
		t.Fatal(err)
	}
	if len(edited.Versions) != 2 || edited.Versions[0].Number != 2 || !edited.Versions[0].Draft() || edited.ActiveVersionID != v1.ID || edited.Status != types.ScriptStatusActive {
		t.Fatalf("editing an active script must open draft v2 and keep v1 active: %+v", edited)
	}
	if got := f.store.versions[v1.ID]; got.Content.Answer != "It costs 200k." {
		t.Fatalf("published v1 content changed: %q", got.Content.Answer)
	}

	// Saving again rewrites the same open draft instead of adding versions.
	again, err := f.service.SaveDraft(ctx, f.tenant, f.owner, script.ID, types.ScriptDraftInput{Name: "Shirt price", Content: types.ScriptContent{Triggers: []string{"shirt price"}, Answer: "Now 170k."}, ExpectedVersion: edited.Version})
	if err != nil {
		t.Fatal(err)
	}
	if len(again.Versions) != 2 || again.Versions[0].ID != edited.Versions[0].ID || again.Versions[0].Content.Answer != "Now 170k." {
		t.Fatalf("second save must update draft v2 in place: %+v", again.Versions)
	}

	v2, err := f.service.Activate(ctx, f.tenant, f.owner, script.ID, again.Versions[0].ID, again.Version)
	if err != nil {
		t.Fatal(err)
	}
	if v2.ActiveVersionID != again.Versions[0].ID {
		t.Fatal("activation must switch the pointer to v2")
	}
	// Re-activating an older published version is a pointer switch; it does not copy or edit content.
	back, err := f.service.Activate(ctx, f.tenant, f.owner, script.ID, v1.ID, v2.Version)
	if err != nil || back.ActiveVersionID != v1.ID || len(back.Versions) != 2 {
		t.Fatalf("rollback to v1 failed: %+v %v", back, err)
	}
	_, err = f.service.Activate(ctx, f.tenant, f.owner, script.ID, v1.ID, back.Version)
	requireScriptStatus(t, err, http.StatusConflict)

	// The store guard rejects rewriting published content even if a caller bypasses the service.
	frozen := f.store.versions[v1.ID]
	frozen.Content.Answer = "tampered"
	requireScriptStatus(t, f.store.UpdateDraft(ctx, frozen), http.StatusConflict)
}

func TestScriptDeactivateAndArchiveKeepHistory(t *testing.T) {
	f := newScriptFixture(t)
	ctx := context.Background()
	script := f.create(t)
	_, err := f.service.Deactivate(ctx, f.tenant, f.owner, script.ID, script.Version)
	requireScriptStatus(t, err, http.StatusConflict)
	active, err := f.service.Activate(ctx, f.tenant, f.owner, script.ID, draftOf(t, script).ID, script.Version)
	if err != nil {
		t.Fatal(err)
	}
	inactive, err := f.service.Deactivate(ctx, f.tenant, f.owner, script.ID, active.Version)
	if err != nil || inactive.Status != types.ScriptStatusInactive || inactive.ActiveVersionID != "" || len(inactive.Versions) != 1 {
		t.Fatalf("deactivate: %+v %v", inactive, err)
	}
	again, err := f.service.Activate(ctx, f.tenant, f.owner, script.ID, inactive.Versions[0].ID, inactive.Version)
	if err != nil {
		t.Fatal(err)
	}
	archived, err := f.service.Archive(ctx, f.tenant, f.owner, script.ID, again.Version)
	if err != nil {
		t.Fatal(err)
	}
	if archived.Status != types.ScriptStatusArchived || archived.ActiveVersionID != "" || archived.ArchivedAt == nil || len(archived.Versions) != 1 {
		t.Fatalf("archive must keep history and clear the pointer: %+v", archived)
	}
	if active, _ := f.service.ActiveScriptVersions(ctx, f.tenant); len(active) != 0 {
		t.Fatalf("archived script must not be offered to the reply router: %+v", active)
	}
	_, err = f.service.SaveDraft(ctx, f.tenant, f.owner, script.ID, types.ScriptDraftInput{Name: "x", Content: types.ScriptContent{Triggers: []string{"x"}, Answer: "x"}, ExpectedVersion: archived.Version})
	requireScriptStatus(t, err, http.StatusConflict)
	_, err = f.service.Activate(ctx, f.tenant, f.owner, script.ID, archived.Versions[0].ID, archived.Version)
	requireScriptStatus(t, err, http.StatusConflict)
	_, err = f.service.Archive(ctx, f.tenant, f.owner, script.ID, archived.Version)
	requireScriptStatus(t, err, http.StatusConflict)
}

func TestScriptLifecycleEnqueuesKnowledgeSync(t *testing.T) {
	f := newScriptFixture(t)
	ctx := context.Background()
	script := f.create(t)
	key := types.KnowledgeSourceScript + ":" + script.ID
	if len(f.store.synced) != 0 {
		t.Fatalf("an inactive script is not bot knowledge: %v", f.store.synced)
	}
	// An inactive script's draft edit is not knowledge either.
	edited, err := f.service.SaveDraft(ctx, f.tenant, f.owner, script.ID, types.ScriptDraftInput{Name: "Shirt", Content: types.ScriptContent{Triggers: []string{"shirt"}, Answer: "200k"}, ExpectedVersion: script.Version})
	if err != nil || len(f.store.synced) != 0 {
		t.Fatalf("draft of an inactive script enqueued sync: %v %v", f.store.synced, err)
	}
	active, err := f.service.Activate(ctx, f.tenant, f.owner, script.ID, draftOf(t, edited).ID, edited.Version)
	if err != nil || len(f.store.synced) != 1 || f.store.synced[0] != key {
		t.Fatalf("activation must enqueue the script: %v %v", f.store.synced, err)
	}
	renamed, err := f.service.SaveDraft(ctx, f.tenant, f.owner, script.ID, types.ScriptDraftInput{Name: "Shirt v2", Content: types.ScriptContent{Triggers: []string{"shirt"}, Answer: "190k"}, ExpectedVersion: active.Version})
	if err != nil || len(f.store.synced) != 2 {
		t.Fatalf("an active script's rename must refresh its title: %v %v", f.store.synced, err)
	}
	inactive, err := f.service.Deactivate(ctx, f.tenant, f.owner, script.ID, renamed.Version)
	if err != nil || len(f.store.synced) != 3 {
		t.Fatalf("deactivation must enqueue removal: %v %v", f.store.synced, err)
	}
	// A rejected transition rolls the enqueue back with the transaction.
	if _, err := f.service.Deactivate(ctx, f.tenant, f.owner, script.ID, inactive.Version); err == nil || len(f.store.synced) != 3 {
		t.Fatalf("failed transition must not enqueue: %v %v", f.store.synced, err)
	}
	if _, err := f.service.Archive(ctx, f.tenant, f.owner, script.ID, inactive.Version); err != nil || len(f.store.synced) != 4 {
		t.Fatalf("archive must enqueue removal: %v %v", f.store.synced, err)
	}
}

func TestScriptStaleExpectedVersionConflicts(t *testing.T) {
	f := newScriptFixture(t)
	ctx := context.Background()
	script := f.create(t)
	edit := types.ScriptDraftInput{Name: "Shirt price", Content: types.ScriptContent{Triggers: []string{"price"}, Answer: "A"}, ExpectedVersion: script.Version}
	saved, err := f.service.SaveDraft(ctx, f.tenant, f.owner, script.ID, edit)
	if err != nil {
		t.Fatal(err)
	}
	edit.Content.Answer = "B"
	_, err = f.service.SaveDraft(ctx, f.tenant, f.owner, script.ID, edit)
	requireScriptStatus(t, err, http.StatusConflict)
	_, err = f.service.Activate(ctx, f.tenant, f.owner, script.ID, draftOf(t, saved).ID, script.Version)
	requireScriptStatus(t, err, http.StatusConflict)
	if got := draftOf(t, saved); f.store.versions[got.ID].Content.Answer != "A" {
		t.Fatal("stale edit must not change the draft")
	}
}

func TestScriptCreateIsIdempotentPerKey(t *testing.T) {
	f := newScriptFixture(t)
	ctx := context.Background()
	key := uuid.NewString()
	first, err := f.service.Create(ctx, f.tenant, f.owner, key, scriptInput("Hi"))
	if err != nil {
		t.Fatal(err)
	}
	replay, err := f.service.Create(ctx, f.tenant, f.owner, key, scriptInput("Hi"))
	if err != nil || replay.ID != first.ID || len(f.store.versions) != 1 {
		t.Fatalf("replay must return the original script without a new version: %+v %v", replay, err)
	}
	_, err = f.service.Create(ctx, f.tenant, f.owner, key, scriptInput("Different"))
	requireScriptStatus(t, err, http.StatusConflict)
	_, err = f.service.Create(ctx, f.other, f.owner, key, scriptInput("Hi"))
	requireScriptStatus(t, err, http.StatusConflict)
	_, err = f.service.Create(ctx, f.tenant, f.owner, "not-a-uuid", scriptInput("Hi"))
	requireScriptStatus(t, err, http.StatusBadRequest)
}

func TestScriptContentValidation(t *testing.T) {
	f := newScriptFixture(t)
	ctx := context.Background()
	long := make([]byte, types.ScriptMaxAnswerLength+1)
	for i := range long {
		long[i] = 'a'
	}
	for name, input := range map[string]types.ScriptCreateInput{
		"no triggers":  {Name: "n", Content: types.ScriptContent{Triggers: []string{"  "}, Answer: "a"}},
		"blank answer": {Name: "n", Content: types.ScriptContent{Triggers: []string{"t"}, Answer: "   "}},
		"long answer":  {Name: "n", Content: types.ScriptContent{Triggers: []string{"t"}, Answer: string(long)}},
		"blank name":   {Name: " ", Content: types.ScriptContent{Triggers: []string{"t"}, Answer: "a"}},
	} {
		_, err := f.service.Create(ctx, f.tenant, f.owner, uuid.NewString(), input)
		if err == nil || err.GetHttpStatus() != http.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %v", name, err)
		}
	}
	script, err := f.service.Create(ctx, f.tenant, f.owner, uuid.NewString(), types.ScriptCreateInput{Name: " Ship ", Content: types.ScriptContent{Triggers: []string{"ship  fee", "Ship fee", " delivery "}, Answer: " Free. "}})
	if err != nil {
		t.Fatal(err)
	}
	content := script.Versions[0].Content
	if script.Name != "Ship" || content.Answer != "Free." || len(content.Triggers) != 2 || content.Triggers[0] != "ship fee" || content.Triggers[1] != "delivery" {
		t.Fatalf("content must be trimmed and de-duplicated: %+v %+v", script.Name, content)
	}
}

func TestScriptPermissionGrantRevokeRules(t *testing.T) {
	f := newScriptFixture(t)
	ctx := context.Background()
	_, err := f.service.Grant(ctx, f.tenant, f.member, f.peer, types.ScriptPermissionEdit, 0)
	requireScriptStatus(t, err, http.StatusForbidden)
	f.grantOnly(t, f.member, types.ScriptPermissionActivate)
	// Holding a script permission never lets a member grant others.
	_, err = f.service.Grant(ctx, f.tenant, f.member, f.peer, types.ScriptPermissionActivate, 0)
	requireScriptStatus(t, err, http.StatusForbidden)
	_, err = f.service.Grant(ctx, f.tenant, f.owner, f.owner, types.ScriptPermissionEdit, 0)
	requireScriptStatus(t, err, http.StatusBadRequest)
	_, err = f.service.Grant(ctx, f.tenant, f.owner, f.outsider, types.ScriptPermissionEdit, 0)
	requireScriptStatus(t, err, http.StatusBadRequest)
	_, err = f.service.Grant(ctx, f.tenant, f.owner, f.member, "script.delete", 1)
	requireScriptStatus(t, err, http.StatusBadRequest)

	// The member's set is at version 1 after the first grant; version 0 is now stale.
	_, err = f.service.Grant(ctx, f.tenant, f.owner, f.member, types.ScriptPermissionEdit, 0)
	requireScriptStatus(t, err, http.StatusConflict)
	same, err := f.service.Grant(ctx, f.tenant, f.owner, f.member, types.ScriptPermissionActivate, 1)
	if err != nil || same.Grants.Version != 1 {
		t.Fatalf("granting a held permission must change nothing: %+v %v", same, err)
	}

	script := f.create(t)
	active, err := f.service.Activate(ctx, f.tenant, f.member, script.ID, draftOf(t, script).ID, script.Version)
	if err != nil {
		t.Fatal(err)
	}
	revoked, err := f.service.Revoke(ctx, f.tenant, f.owner, f.member, types.ScriptPermissionActivate, 1)
	if err != nil || revoked.Grants.Version != 2 || len(revoked.Grants.Granted()) != 0 {
		t.Fatalf("revoke: %+v %v", revoked, err)
	}
	// Revocation applies to the member's very next action.
	_, err = f.service.Deactivate(ctx, f.tenant, f.member, script.ID, active.Version)
	requireScriptStatus(t, err, http.StatusForbidden)

	// Removing the membership removes authority even if a grant row lingered.
	f.grantOnly(t, f.member, types.ScriptPermissionArchive)
	delete(f.store.members, memberKey(f.tenant, f.member))
	_, err = f.service.Archive(ctx, f.tenant, f.member, script.ID, active.Version)
	requireScriptStatus(t, err, http.StatusForbidden)

	me, err := f.service.MyPermissions(ctx, f.tenant, f.owner)
	if err != nil || len(me.Effective()) != 4 {
		t.Fatalf("owner implicitly holds every permission: %+v %v", me, err)
	}
}

func TestActiveScriptVersionsReturnsOnlyPublishedActiveContent(t *testing.T) {
	f := newScriptFixture(t)
	ctx := context.Background()
	script := f.create(t)
	f.create(t) // stays inactive
	active, err := f.service.Activate(ctx, f.tenant, f.owner, script.ID, draftOf(t, script).ID, script.Version)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.service.SaveDraft(ctx, f.tenant, f.owner, script.ID, types.ScriptDraftInput{Name: "Shirt price", Content: types.ScriptContent{Triggers: []string{"x"}, Answer: "unpublished draft"}, ExpectedVersion: active.Version}); err != nil {
		t.Fatal(err)
	}
	rows, err := f.service.ActiveScriptVersions(ctx, f.tenant)
	if err != nil || len(rows) != 1 || rows[0].VersionID != active.ActiveVersionID || rows[0].Content.Answer != "It costs 200k." {
		t.Fatalf("only the active published version may be returned: %+v %v", rows, err)
	}
	if rows, _ := f.service.ActiveScriptVersions(ctx, f.other); len(rows) != 0 {
		t.Fatal("active versions must be tenant scoped")
	}
	_, err = f.service.ActiveScriptVersions(ctx, "bad")
	requireScriptStatus(t, err, http.StatusBadRequest)
}
