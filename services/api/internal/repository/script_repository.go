package repository

import (
	"context"
	stderrors "errors"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/mapper/script/persistencemap"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/scriptmodel"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ScriptStore persists scripts, their versions and delegated script permissions. Every query is tenant-scoped;
// authorization decisions stay in the script service.
type ScriptStore interface {
	Within(context.Context, func(ScriptStore) *apperrors.Error) *apperrors.Error
	// Access reads the caller's tenant, membership and grants under share locks, so a concurrent revoke waits.
	Access(ctx context.Context, tenant, actor string) (types.ScriptAccess, *apperrors.Error)
	Scripts(ctx context.Context, tenant string, filter types.ScriptListFilter, limit, offset int) ([]types.ScriptSummary, int64, *apperrors.Error)
	Script(ctx context.Context, tenant, id string, lock bool) (*types.Script, *apperrors.Error)
	Versions(ctx context.Context, tenant, script string, limit int) ([]types.ScriptVersion, *apperrors.Error)
	Version(ctx context.Context, tenant, script, id string) (*types.ScriptVersion, *apperrors.Error)
	// Draft returns the script's single unpublished version, locked, or nil.
	Draft(ctx context.Context, tenant, script string) (*types.ScriptVersion, *apperrors.Error)
	NextNumber(ctx context.Context, tenant, script string) (int64, *apperrors.Error)
	InsertScript(ctx context.Context, script types.Script) (bool, *apperrors.Error)
	InsertVersion(ctx context.Context, version types.ScriptVersion) *apperrors.Error
	// UpdateDraft rewrites content only while the version is unpublished; published content is immutable.
	UpdateDraft(ctx context.Context, version types.ScriptVersion) *apperrors.Error
	// PublishVersion freezes a draft; a version that is already published is left unchanged.
	PublishVersion(ctx context.Context, tenant, script, id, actor string, at time.Time) *apperrors.Error
	SaveScript(ctx context.Context, script types.Script, expected int64) *apperrors.Error
	ActiveVersions(ctx context.Context, tenant string, limit int) ([]types.ActiveScriptVersion, *apperrors.Error)
	MemberPermissions(ctx context.Context, tenant, owner string, limit int) ([]types.ScriptMemberPermissions, *apperrors.Error)
	IsMember(ctx context.Context, tenant, user string) (bool, *apperrors.Error)
	// PermissionSet locks and returns a member's grants; no row is the empty set at version 0.
	PermissionSet(ctx context.Context, tenant, user string) (types.ScriptPermissionSet, *apperrors.Error)
	SavePermissionSet(ctx context.Context, set types.ScriptPermissionSet, expected int64) *apperrors.Error
	Audit(ctx context.Context, tenant, actor, action, kind, id string, state any) *apperrors.Error
	// EnqueueKnowledgeSync records a changed AI knowledge source in the caller's transaction.
	EnqueueKnowledgeSync(ctx context.Context, tenant, sourceType, sourceID string, at time.Time) *apperrors.Error
}

type ScriptRepository struct{ *SalesRepository }

func NewScriptRepository(db *gorm.DB) *ScriptRepository {
	return &ScriptRepository{NewSalesRepository(db)}
}

func (r *ScriptRepository) Within(ctx context.Context, fn func(ScriptStore) *apperrors.Error) *apperrors.Error {
	var failure *apperrors.Error
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if failure = fn(NewScriptRepository(tx)); failure != nil {
			return failure
		}
		return nil
	})
	if failure != nil {
		return failure
	}
	if err != nil {
		return scriptDBError(ctx, err)
	}
	return nil
}

// scriptDBError maps constraint races (second draft, duplicate number, removed member) to stable conflicts.
func scriptDBError(ctx context.Context, err error) *apperrors.Error {
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return apperrors.ErrNotFound(ctx, "script", "not found")
	}
	var pg *pgconn.PgError
	if stderrors.As(err, &pg) {
		switch {
		case pg.Code == "23505":
			return apperrors.ErrConflict(ctx, "script", "changed; refresh before retrying")
		case pg.Code == "23503" && pg.ConstraintName == "script_member_permissions_member_fkey":
			return apperrors.ErrConflict(ctx, "member", "is no longer a current workspace member")
		}
	}
	return apperrors.ErrSystemError(ctx, "script persistence operation failed")
}

func storedContentError(ctx context.Context) *apperrors.Error {
	return apperrors.ErrSystemError(ctx, "stored script content is invalid")
}

func (r *ScriptRepository) Access(ctx context.Context, tenant, actor string) (types.ScriptAccess, *apperrors.Error) {
	workspace, err := r.Tenant(ctx, tenant)
	if err != nil {
		return types.ScriptAccess{}, err
	}
	member, err := r.Member(ctx, tenant, actor)
	if err != nil {
		return types.ScriptAccess{}, err
	}
	access := types.ScriptAccess{ActiveTenant: workspace.Status == model.TenantStatusActive, Member: member != nil, Owner: workspace.OwnerID == actor}
	var row scriptmodel.MemberPermissions
	q := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "SHARE"}).Where("tenant_id = ? AND user_id = ?", tenant, actor).Limit(1).Find(&row)
	if q.Error != nil {
		return types.ScriptAccess{}, scriptDBError(ctx, q.Error)
	}
	if q.RowsAffected == 1 {
		access.Grants = persistencemap.PermissionSet(row)
	}
	return access, nil
}

func (r *ScriptRepository) Scripts(ctx context.Context, tenant string, filter types.ScriptListFilter, limit, offset int) ([]types.ScriptSummary, int64, *apperrors.Error) {
	count := r.db.WithContext(ctx).Model(&scriptmodel.Script{}).Where("tenant_id = ?", tenant)
	if filter.Status != "" {
		count = count.Where("status = ?", filter.Status)
	}
	var total int64
	if err := count.Count(&total).Error; err != nil {
		return nil, 0, scriptDBError(ctx, err)
	}
	q := r.db.WithContext(ctx).Table("scripts AS s").
		Select("s.*, av.number AS active_number, d.number AS draft_number").
		Joins("LEFT JOIN script_versions AS av ON av.tenant_id = s.tenant_id AND av.script_id = s.id AND av.id = s.active_version_id").
		Joins("LEFT JOIN script_versions AS d ON d.tenant_id = s.tenant_id AND d.script_id = s.id AND d.published_at IS NULL").
		Where("s.tenant_id = ?", tenant)
	if filter.Status != "" {
		q = q.Where("s.status = ?", filter.Status)
	}
	rows := []scriptmodel.ScriptSummary{}
	if err := q.Order("s.updated_at DESC, s.id").Limit(limit).Offset(offset).Scan(&rows).Error; err != nil {
		return nil, 0, scriptDBError(ctx, err)
	}
	result := make([]types.ScriptSummary, len(rows))
	for i, row := range rows {
		result[i] = persistencemap.Summary(row)
	}
	return result, total, nil
}

func (r *ScriptRepository) Script(ctx context.Context, tenant, id string, lock bool) (*types.Script, *apperrors.Error) {
	var row scriptmodel.Script
	q := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenant, id)
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := q.First(&row).Error; err != nil {
		return nil, scriptDBError(ctx, err)
	}
	value := persistencemap.Script(row)
	return &value, nil
}

func (r *ScriptRepository) versions(ctx context.Context, rows []scriptmodel.Version) ([]types.ScriptVersion, *apperrors.Error) {
	result := make([]types.ScriptVersion, len(rows))
	for i, row := range rows {
		value, err := persistencemap.Version(row)
		if err != nil {
			return nil, storedContentError(ctx)
		}
		result[i] = value
	}
	return result, nil
}

// Versions returns the newest versions first, bounded by limit.
func (r *ScriptRepository) Versions(ctx context.Context, tenant, script string, limit int) ([]types.ScriptVersion, *apperrors.Error) {
	rows := []scriptmodel.Version{}
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND script_id = ?", tenant, script).Order("number DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, scriptDBError(ctx, err)
	}
	return r.versions(ctx, rows)
}

func (r *ScriptRepository) Version(ctx context.Context, tenant, script, id string) (*types.ScriptVersion, *apperrors.Error) {
	var row scriptmodel.Version
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND script_id = ? AND id = ?", tenant, script, id).First(&row).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound(ctx, "script version", "not found")
		}
		return nil, scriptDBError(ctx, err)
	}
	value, err := persistencemap.Version(row)
	if err != nil {
		return nil, storedContentError(ctx)
	}
	return &value, nil
}

func (r *ScriptRepository) Draft(ctx context.Context, tenant, script string) (*types.ScriptVersion, *apperrors.Error) {
	rows := []scriptmodel.Version{}
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("tenant_id = ? AND script_id = ? AND published_at IS NULL", tenant, script).Limit(1).Find(&rows).Error; err != nil {
		return nil, scriptDBError(ctx, err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	values, err := r.versions(ctx, rows)
	if err != nil {
		return nil, err
	}
	return &values[0], nil
}

// NextNumber must run while the script row is locked; the unique number constraint still rejects a race.
func (r *ScriptRepository) NextNumber(ctx context.Context, tenant, script string) (int64, *apperrors.Error) {
	var next int64
	if err := r.db.WithContext(ctx).Raw(`SELECT COALESCE(MAX(number), 0) + 1 FROM script_versions WHERE tenant_id = ? AND script_id = ?`, tenant, script).Scan(&next).Error; err != nil {
		return 0, scriptDBError(ctx, err)
	}
	return next, nil
}

// InsertScript reports false when the ID (idempotency key) already exists.
func (r *ScriptRepository) InsertScript(ctx context.Context, script types.Script) (bool, *apperrors.Error) {
	row := persistencemap.ScriptRow(script)
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&row)
	if result.Error != nil {
		return false, scriptDBError(ctx, result.Error)
	}
	return result.RowsAffected == 1, nil
}

func (r *ScriptRepository) InsertVersion(ctx context.Context, version types.ScriptVersion) *apperrors.Error {
	row := persistencemap.VersionRow(version)
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return scriptDBError(ctx, err)
	}
	return nil
}

func (r *ScriptRepository) UpdateDraft(ctx context.Context, version types.ScriptVersion) *apperrors.Error {
	result := r.db.WithContext(ctx).Model(&scriptmodel.Version{}).
		Where("tenant_id = ? AND script_id = ? AND id = ? AND published_at IS NULL", version.TenantID, version.ScriptID, version.ID).
		Updates(map[string]any{"triggers": gorm.Expr("?::jsonb", persistencemap.Triggers(version.Content.Triggers)), "answer": version.Content.Answer, "updated_by": version.UpdatedBy, "updated_at": version.UpdatedAt})
	if result.Error != nil {
		return scriptDBError(ctx, result.Error)
	}
	if result.RowsAffected != 1 {
		return apperrors.ErrConflict(ctx, "script version", "is published and cannot be changed")
	}
	return nil
}

func (r *ScriptRepository) PublishVersion(ctx context.Context, tenant, script, id, actor string, at time.Time) *apperrors.Error {
	err := r.db.WithContext(ctx).Model(&scriptmodel.Version{}).
		Where("tenant_id = ? AND script_id = ? AND id = ? AND published_at IS NULL", tenant, script, id).
		Updates(map[string]any{"published_at": at, "published_by": actor}).Error
	if err != nil {
		return scriptDBError(ctx, err)
	}
	return nil
}

// SaveScript applies a versioned change; the expected version predicate rejects concurrent updates.
func (r *ScriptRepository) SaveScript(ctx context.Context, script types.Script, expected int64) *apperrors.Error {
	row := persistencemap.ScriptRow(script)
	result := r.db.WithContext(ctx).Model(&scriptmodel.Script{}).Where("tenant_id = ? AND id = ? AND version = ?", script.TenantID, script.ID, expected).
		Select("name", "status", "active_version_id", "version", "updated_at", "archived_at").Updates(&row)
	if result.Error != nil {
		return scriptDBError(ctx, result.Error)
	}
	if result.RowsAffected != 1 {
		return apperrors.ErrConflict(ctx, "script", "changed; refresh before retrying")
	}
	return nil
}

func (r *ScriptRepository) ActiveVersions(ctx context.Context, tenant string, limit int) ([]types.ActiveScriptVersion, *apperrors.Error) {
	rows := []scriptmodel.ActiveVersion{}
	err := r.db.WithContext(ctx).Table("scripts AS s").
		Select("s.id AS script_id, s.name AS script_name, v.id AS version_id, v.number, v.triggers, v.answer, v.published_at").
		Joins("JOIN tenants AS t ON t.id = s.tenant_id AND t.status = ?", model.TenantStatusActive).
		Joins("JOIN script_versions AS v ON v.tenant_id = s.tenant_id AND v.script_id = s.id AND v.id = s.active_version_id").
		Where("s.tenant_id = ? AND s.status = ?", tenant, types.ScriptStatusActive).
		Order("s.name, s.id").Limit(limit).Scan(&rows).Error
	if err != nil {
		return nil, scriptDBError(ctx, err)
	}
	result := make([]types.ActiveScriptVersion, len(rows))
	for i, row := range rows {
		value, err := persistencemap.ActiveVersion(row)
		if err != nil {
			return nil, storedContentError(ctx)
		}
		result[i] = value
	}
	return result, nil
}

// MemberPermissions lists current members with their grants; pending invitations are not members.
func (r *ScriptRepository) MemberPermissions(ctx context.Context, tenant, owner string, limit int) ([]types.ScriptMemberPermissions, *apperrors.Error) {
	rows := []scriptmodel.MemberRow{}
	err := r.db.WithContext(ctx).Table("tenant_members AS m").
		Select("m.user_id, m.role, p.can_create, p.can_edit, p.can_activate, p.can_archive, p.version, p.updated_by, p.updated_at").
		Joins("LEFT JOIN script_member_permissions AS p ON p.tenant_id = m.tenant_id AND p.user_id = m.user_id").
		Where("m.tenant_id = ?", tenant).Order("m.created_at, m.id").Limit(limit).Scan(&rows).Error
	if err != nil {
		return nil, scriptDBError(ctx, err)
	}
	result := make([]types.ScriptMemberPermissions, len(rows))
	for i, row := range rows {
		result[i] = persistencemap.Member(tenant, owner, row, "")
	}
	return result, nil
}

func (r *ScriptRepository) IsMember(ctx context.Context, tenant, user string) (bool, *apperrors.Error) {
	member, err := r.Member(ctx, tenant, user)
	if err != nil {
		return false, err
	}
	return member != nil, nil
}

func (r *ScriptRepository) PermissionSet(ctx context.Context, tenant, user string) (types.ScriptPermissionSet, *apperrors.Error) {
	var row scriptmodel.MemberPermissions
	q := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("tenant_id = ? AND user_id = ?", tenant, user).Limit(1).Find(&row)
	if q.Error != nil {
		return types.ScriptPermissionSet{}, scriptDBError(ctx, q.Error)
	}
	if q.RowsAffected == 0 {
		return types.ScriptPermissionSet{TenantID: tenant, UserID: user}, nil
	}
	return persistencemap.PermissionSet(row), nil
}

// SavePermissionSet inserts the first grant row (expected 0) or applies a versioned update.
func (r *ScriptRepository) SavePermissionSet(ctx context.Context, set types.ScriptPermissionSet, expected int64) *apperrors.Error {
	row := persistencemap.PermissionRow(set)
	var result *gorm.DB
	if expected == 0 {
		row.CreatedAt = set.UpdatedAt
		result = r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&row)
	} else {
		result = r.db.WithContext(ctx).Model(&scriptmodel.MemberPermissions{}).Where("tenant_id = ? AND user_id = ? AND version = ?", set.TenantID, set.UserID, expected).
			Select("can_create", "can_edit", "can_activate", "can_archive", "version", "updated_by", "updated_at").Updates(&row)
	}
	if result.Error != nil {
		return scriptDBError(ctx, result.Error)
	}
	if result.RowsAffected != 1 {
		return apperrors.ErrConflict(ctx, "script permissions", "were updated elsewhere; reload to see them")
	}
	return nil
}

var _ ScriptStore = (*ScriptRepository)(nil)
