// Package scriptmodel holds the GORM rows of the script aggregate and delegated script permissions.
package scriptmodel

import "time"

type Script struct {
	ID              string
	TenantID        string
	Name            string
	Status          string
	ActiveVersionID *string
	Version         int64
	CreateHash      string
	CreatedBy       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ArchivedAt      *time.Time
}

func (Script) TableName() string { return "scripts" }

// ScriptSummary is a list projection joining the active and draft version numbers.
type ScriptSummary struct {
	Script
	ActiveNumber *int64
	DraftNumber  *int64
}

type Version struct {
	ID          string
	TenantID    string
	ScriptID    string
	Number      int64
	Triggers    string
	Answer      string
	CreatedBy   string
	UpdatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt *time.Time
	PublishedBy *string
}

func (Version) TableName() string { return "script_versions" }

// ActiveVersion is the reply-router projection of an active script's current version.
type ActiveVersion struct {
	ScriptID    string
	ScriptName  string
	VersionID   string
	Number      int64
	Triggers    string
	Answer      string
	PublishedAt time.Time
}

type MemberPermissions struct {
	TenantID    string
	UserID      string
	CanCreate   bool
	CanEdit     bool
	CanActivate bool
	CanArchive  bool
	Version     int64
	UpdatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (MemberPermissions) TableName() string { return "script_member_permissions" }

// MemberRow is a current member joined with its optional grant row.
type MemberRow struct {
	UserID      string
	Role        string
	CanCreate   *bool
	CanEdit     *bool
	CanActivate *bool
	CanArchive  *bool
	Version     *int64
	UpdatedBy   *string
	UpdatedAt   *time.Time
}
