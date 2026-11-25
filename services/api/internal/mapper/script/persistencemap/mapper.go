// Package persistencemap converts script, script version and script permission rows to core values.
package persistencemap

import (
	"encoding/json"
	"errors"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/scriptmodel"
)

// ErrStoredTriggers reports a triggers column that is not a JSON array of strings.
var ErrStoredTriggers = errors.New("stored script triggers are not a string array")

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

func number(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func flag(value *bool) bool { return value != nil && *value }

func Script(row scriptmodel.Script) types.Script {
	return types.Script{ID: row.ID, TenantID: row.TenantID, Name: row.Name, Status: row.Status, ActiveVersionID: text(row.ActiveVersionID), Version: row.Version, CreateHash: row.CreateHash, CreatedBy: row.CreatedBy, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt, ArchivedAt: row.ArchivedAt}
}

func ScriptRow(value types.Script) scriptmodel.Script {
	return scriptmodel.Script{ID: value.ID, TenantID: value.TenantID, Name: value.Name, Status: value.Status, ActiveVersionID: optional(value.ActiveVersionID), Version: value.Version, CreateHash: value.CreateHash, CreatedBy: value.CreatedBy, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt, ArchivedAt: value.ArchivedAt}
}

func Summary(row scriptmodel.ScriptSummary) types.ScriptSummary {
	return types.ScriptSummary{Script: Script(row.Script), ActiveNumber: number(row.ActiveNumber), DraftNumber: number(row.DraftNumber)}
}

// Triggers encodes trigger phrases as a JSON array; nil becomes an empty array.
func Triggers(values []string) string {
	if values == nil {
		values = []string{}
	}
	data, _ := json.Marshal(values)
	return string(data)
}

func triggers(raw string) ([]string, error) {
	values := []string{}
	if err := json.Unmarshal([]byte(raw), &values); err != nil || values == nil {
		return nil, ErrStoredTriggers
	}
	return values, nil
}

func Version(row scriptmodel.Version) (types.ScriptVersion, error) {
	phrases, err := triggers(row.Triggers)
	if err != nil {
		return types.ScriptVersion{}, err
	}
	return types.ScriptVersion{ID: row.ID, TenantID: row.TenantID, ScriptID: row.ScriptID, Number: row.Number, Content: types.ScriptContent{Triggers: phrases, Answer: row.Answer}, CreatedBy: row.CreatedBy, UpdatedBy: row.UpdatedBy, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt, PublishedAt: row.PublishedAt, PublishedBy: text(row.PublishedBy)}, nil
}

func VersionRow(value types.ScriptVersion) scriptmodel.Version {
	return scriptmodel.Version{ID: value.ID, TenantID: value.TenantID, ScriptID: value.ScriptID, Number: value.Number, Triggers: Triggers(value.Content.Triggers), Answer: value.Content.Answer, CreatedBy: value.CreatedBy, UpdatedBy: value.UpdatedBy, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt, PublishedAt: value.PublishedAt, PublishedBy: optional(value.PublishedBy)}
}

func ActiveVersion(row scriptmodel.ActiveVersion) (types.ActiveScriptVersion, error) {
	phrases, err := triggers(row.Triggers)
	if err != nil {
		return types.ActiveScriptVersion{}, err
	}
	return types.ActiveScriptVersion{ScriptID: row.ScriptID, ScriptName: row.ScriptName, VersionID: row.VersionID, Number: row.Number, Content: types.ScriptContent{Triggers: phrases, Answer: row.Answer}, PublishedAt: row.PublishedAt}, nil
}

func PermissionSet(row scriptmodel.MemberPermissions) types.ScriptPermissionSet {
	return types.ScriptPermissionSet{TenantID: row.TenantID, UserID: row.UserID, Create: row.CanCreate, Edit: row.CanEdit, Activate: row.CanActivate, Archive: row.CanArchive, Version: row.Version, UpdatedBy: row.UpdatedBy, UpdatedAt: row.UpdatedAt}
}

func PermissionRow(value types.ScriptPermissionSet) scriptmodel.MemberPermissions {
	return scriptmodel.MemberPermissions{TenantID: value.TenantID, UserID: value.UserID, CanCreate: value.Create, CanEdit: value.Edit, CanActivate: value.Activate, CanArchive: value.Archive, Version: value.Version, UpdatedBy: value.UpdatedBy, UpdatedAt: value.UpdatedAt}
}

// Member maps a membership joined with its optional grant row; a missing row is the empty set at version 0.
func Member(tenant, owner string, row scriptmodel.MemberRow, email string) types.ScriptMemberPermissions {
	grants := types.ScriptPermissionSet{TenantID: tenant, UserID: row.UserID, Create: flag(row.CanCreate), Edit: flag(row.CanEdit), Activate: flag(row.CanActivate), Archive: flag(row.CanArchive), Version: number(row.Version), UpdatedBy: text(row.UpdatedBy)}
	if row.UpdatedAt != nil {
		grants.UpdatedAt = *row.UpdatedAt
	}
	return types.ScriptMemberPermissions{UserID: row.UserID, Email: email, Role: row.Role, Owner: row.UserID == owner, Grants: grants}
}
