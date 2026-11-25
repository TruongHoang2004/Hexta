package dto

import "time"

// CreateScriptRequest bounds mirror the service rules: 1–20 trigger phrases of at most 120 characters and an
// answer of at most 2000 characters. The service trims and de-duplicates phrases.
type CreateScriptRequest struct {
	Name     string   `json:"name" validate:"required,max=120"`
	Triggers []string `json:"triggers" validate:"required,min=1,max=20,dive,max=120"`
	Answer   string   `json:"answer" validate:"required,max=2000"`
}

type SaveScriptDraftRequest struct {
	Name            string   `json:"name" validate:"required,max=120"`
	Triggers        []string `json:"triggers" validate:"required,min=1,max=20,dive,max=120"`
	Answer          string   `json:"answer" validate:"required,max=2000"`
	ExpectedVersion int64    `json:"expected_version" validate:"min=1"`
}

type ActivateScriptRequest struct {
	VersionID       string `json:"version_id" validate:"required,uuid"`
	ExpectedVersion int64  `json:"expected_version" validate:"min=1"`
}

// ScriptTransitionRequest carries the expected script version for deactivate and archive.
type ScriptTransitionRequest struct {
	ExpectedVersion int64 `json:"expected_version" validate:"min=1"`
}

type ScriptListQuery struct {
	Status string `form:"status" validate:"omitempty,oneof=inactive active archived"`
}

type ScriptVersionDTO struct {
	ID          string     `json:"id"`
	Number      int64      `json:"number"`
	Triggers    []string   `json:"triggers"`
	Answer      string     `json:"answer"`
	Draft       bool       `json:"draft"`
	CreatedBy   string     `json:"created_by"`
	UpdatedBy   string     `json:"updated_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	PublishedBy string     `json:"published_by,omitempty"`
}

type ScriptSummaryDTO struct {
	ID                  string     `json:"id"`
	Name                string     `json:"name"`
	Status              string     `json:"status"`
	ActiveVersionID     string     `json:"active_version_id,omitempty"`
	ActiveVersionNumber int64      `json:"active_version_number,omitempty"`
	DraftVersionNumber  int64      `json:"draft_version_number,omitempty"`
	Version             int64      `json:"version"`
	CreatedBy           string     `json:"created_by"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	ArchivedAt          *time.Time `json:"archived_at,omitempty"`
}

// ScriptDTO is a script with its version history, newest first (at most 100 versions).
type ScriptDTO struct {
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	Status          string             `json:"status"`
	ActiveVersionID string             `json:"active_version_id,omitempty"`
	Version         int64              `json:"version"`
	CreatedBy       string             `json:"created_by"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	ArchivedAt      *time.Time         `json:"archived_at,omitempty"`
	Versions        []ScriptVersionDTO `json:"versions"`
}

type ScriptsResponse struct {
	Items []ScriptSummaryDTO `json:"items"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Size  int                `json:"size"`
}

type ScriptPermissionChangeRequest struct {
	Permission      string `json:"permission" validate:"required,oneof=script.create script.edit script.activate script.archive"`
	ExpectedVersion int64  `json:"expected_version" validate:"min=0"`
}

// MemberScriptPermissionsDTO: owner rows always list all four permissions and cannot be changed.
type MemberScriptPermissionsDTO struct {
	UserID      string     `json:"user_id"`
	Email       string     `json:"email,omitempty"`
	Role        string     `json:"role,omitempty"`
	Owner       bool       `json:"owner"`
	Permissions []string   `json:"permissions"`
	Version     int64      `json:"version"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type ScriptPermissionsResponse struct {
	Members []MemberScriptPermissionsDTO `json:"members"`
}

// MyScriptPermissionsResponse lists the caller's effective permissions, for showing or hiding actions only.
type MyScriptPermissionsResponse struct {
	Owner       bool     `json:"owner"`
	Permissions []string `json:"permissions"`
}
