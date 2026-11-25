package types

import "time"

const (
	ScriptStatusInactive = "inactive"
	ScriptStatusActive   = "active"
	ScriptStatusArchived = "archived"

	// The four delegated script permissions. Each grants only its own action; the tenant owner holds all four.
	ScriptPermissionCreate   = "script.create"
	ScriptPermissionEdit     = "script.edit"
	ScriptPermissionActivate = "script.activate"
	ScriptPermissionArchive  = "script.archive"

	ScriptMaxTriggers      = 20
	ScriptMaxTriggerLength = 120
	ScriptMaxAnswerLength  = 2000
	ScriptMaxNameLength    = 120
)

// ScriptPermissions lists every delegable permission in display order.
var ScriptPermissions = []string{ScriptPermissionCreate, ScriptPermissionEdit, ScriptPermissionActivate, ScriptPermissionArchive}

// ScriptPermissionSet is one member's delegated grants. Version 0 means no grant has ever been recorded.
type ScriptPermissionSet struct {
	TenantID, UserID                string
	Create, Edit, Activate, Archive bool
	Version                         int64
	UpdatedBy                       string
	UpdatedAt                       time.Time
}

// Has reports whether the set grants the given permission.
func (s ScriptPermissionSet) Has(permission string) bool {
	switch permission {
	case ScriptPermissionCreate:
		return s.Create
	case ScriptPermissionEdit:
		return s.Edit
	case ScriptPermissionActivate:
		return s.Activate
	case ScriptPermissionArchive:
		return s.Archive
	}
	return false
}

// With returns a copy with the permission set to granted.
func (s ScriptPermissionSet) With(permission string, granted bool) ScriptPermissionSet {
	switch permission {
	case ScriptPermissionCreate:
		s.Create = granted
	case ScriptPermissionEdit:
		s.Edit = granted
	case ScriptPermissionActivate:
		s.Activate = granted
	case ScriptPermissionArchive:
		s.Archive = granted
	}
	return s
}

// Granted lists the granted permissions in display order.
func (s ScriptPermissionSet) Granted() []string {
	result := []string{}
	for _, p := range ScriptPermissions {
		if s.Has(p) {
			result = append(result, p)
		}
	}
	return result
}

// ScriptAccess is the caller's current-state authority, read inside the command transaction.
type ScriptAccess struct {
	ActiveTenant, Member, Owner bool
	Grants                      ScriptPermissionSet
}

// Allows reports whether the caller may perform the action guarded by permission.
func (a ScriptAccess) Allows(permission string) bool {
	return a.ActiveTenant && a.Member && (a.Owner || a.Grants.Has(permission))
}

// Effective lists the permissions the caller holds now.
func (a ScriptAccess) Effective() []string {
	result := []string{}
	for _, p := range ScriptPermissions {
		if a.Allows(p) {
			result = append(result, p)
		}
	}
	return result
}

// ScriptContent is the versioned part of a script: the customer phrases it answers and the reply text.
type ScriptContent struct {
	Triggers []string
	Answer   string
}

// ScriptVersion is one numbered version. PublishedAt is set by its first activation; from then on its content is
// immutable. An unpublished version is the script's single open draft.
type ScriptVersion struct {
	ID, TenantID, ScriptID string
	Number                 int64
	Content                ScriptContent
	CreatedBy, UpdatedBy   string
	CreatedAt, UpdatedAt   time.Time
	PublishedAt            *time.Time
	PublishedBy            string
}

func (v ScriptVersion) Draft() bool { return v.PublishedAt == nil }

// Script is the aggregate root. ActiveVersionID is non-empty exactly when Status is active.
type Script struct {
	ID, TenantID, Name, Status, ActiveVersionID string
	Version                                     int64
	CreateHash, CreatedBy                       string
	CreatedAt, UpdatedAt                        time.Time
	ArchivedAt                                  *time.Time
}

// ScriptSummary is a list row with the active and draft version numbers (0 when absent).
type ScriptSummary struct {
	Script
	ActiveNumber, DraftNumber int64
}

// ScriptDetail is a script with its version history, newest first.
type ScriptDetail struct {
	Script
	Versions []ScriptVersion
}

type ScriptListFilter struct {
	Status string
}

type ScriptCreateInput struct {
	Name    string
	Content ScriptContent
}

// ScriptDraftInput replaces the open draft's content, or opens a new draft when none exists.
type ScriptDraftInput struct {
	Name            string
	Content         ScriptContent
	ExpectedVersion int64
}

// ScriptMemberPermissions is one row of the owner's permission screen.
type ScriptMemberPermissions struct {
	UserID, Email, Role string
	Owner               bool
	Grants              ScriptPermissionSet
}

// ActiveScriptVersion is the published content a reply router may use. The version ID is stable and its content
// never changes, so a proposal can record it as its source.
type ActiveScriptVersion struct {
	ScriptID, ScriptName, VersionID string
	Number                          int64
	Content                         ScriptContent
	PublishedAt                     time.Time
}
