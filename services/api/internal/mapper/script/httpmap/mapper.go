// Package httpmap converts script use-case values to HTTP DTOs.
package httpmap

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
)

func phrases(values []string) []string {
	if values == nil {
		return []string{}
	}
	return append([]string{}, values...)
}

func Version(value types.ScriptVersion) dto.ScriptVersionDTO {
	return dto.ScriptVersionDTO{ID: value.ID, Number: value.Number, Triggers: phrases(value.Content.Triggers), Answer: value.Content.Answer, Draft: value.Draft(), CreatedBy: value.CreatedBy, UpdatedBy: value.UpdatedBy, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt, PublishedAt: value.PublishedAt, PublishedBy: value.PublishedBy}
}

func Script(value types.ScriptDetail) dto.ScriptDTO {
	result := dto.ScriptDTO{ID: value.ID, Name: value.Name, Status: value.Status, ActiveVersionID: value.ActiveVersionID, Version: value.Version, CreatedBy: value.CreatedBy, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt, ArchivedAt: value.ArchivedAt, Versions: make([]dto.ScriptVersionDTO, len(value.Versions))}
	for i, v := range value.Versions {
		result.Versions[i] = Version(v)
	}
	return result
}

func Summary(value types.ScriptSummary) dto.ScriptSummaryDTO {
	return dto.ScriptSummaryDTO{ID: value.ID, Name: value.Name, Status: value.Status, ActiveVersionID: value.ActiveVersionID, ActiveVersionNumber: value.ActiveNumber, DraftVersionNumber: value.DraftNumber, Version: value.Version, CreatedBy: value.CreatedBy, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt, ArchivedAt: value.ArchivedAt}
}

func Scripts(values []types.ScriptSummary, total int64, page, size int) dto.ScriptsResponse {
	result := dto.ScriptsResponse{Items: make([]dto.ScriptSummaryDTO, len(values)), Total: total, Page: page, Size: size}
	for i, v := range values {
		result.Items[i] = Summary(v)
	}
	return result
}

// Member maps a permission row; the owner is reported with every permission because ownership implies them all.
func Member(value types.ScriptMemberPermissions) dto.MemberScriptPermissionsDTO {
	result := dto.MemberScriptPermissionsDTO{UserID: value.UserID, Email: value.Email, Role: value.Role, Owner: value.Owner, Permissions: value.Grants.Granted(), Version: value.Grants.Version}
	if value.Owner {
		result.Permissions = append([]string{}, types.ScriptPermissions...)
	}
	if !value.Grants.UpdatedAt.IsZero() {
		at := value.Grants.UpdatedAt
		result.UpdatedAt = &at
	}
	return result
}

func Members(values []types.ScriptMemberPermissions) dto.ScriptPermissionsResponse {
	result := dto.ScriptPermissionsResponse{Members: make([]dto.MemberScriptPermissionsDTO, len(values))}
	for i, v := range values {
		result.Members[i] = Member(v)
	}
	return result
}

func MyPermissions(value types.ScriptAccess) dto.MyScriptPermissionsResponse {
	return dto.MyScriptPermissionsResponse{Owner: value.Owner, Permissions: value.Effective()}
}
