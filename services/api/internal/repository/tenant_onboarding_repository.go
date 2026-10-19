package repository

import (
	"context"
	stderrors "errors"
	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TenantOnboardingRepository struct{ db *gorm.DB }

func NewTenantOnboardingRepository(db *gorm.DB) *TenantOnboardingRepository {
	return &TenantOnboardingRepository{db}
}
func (r *TenantOnboardingRepository) Transaction(ctx context.Context, fn func(*TenantOnboardingRepository) *apperrors.Error) *apperrors.Error {
	var failure *apperrors.Error
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		failure = fn(&TenantOnboardingRepository{tx})
		if failure != nil {
			return failure
		}
		return nil
	})
	if failure != nil {
		return failure
	}
	if err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}
func (r *TenantOnboardingRepository) Operator(ctx context.Context, user string) (bool, *apperrors.Error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.PlatformOperator{}).Where("user_id = ?", user).Count(&count).Error; err != nil {
		return false, salesDBError(ctx, err)
	}
	return count > 0, nil
}
func (r *TenantOnboardingRepository) Requests(ctx context.Context, applicant string) ([]model.TenantRegistration, *apperrors.Error) {
	rows := []model.TenantRegistration{}
	q := r.db.WithContext(ctx)
	if applicant != "" {
		q = q.Where("applicant_id = ?", applicant)
	}
	if err := q.Order("created_at DESC, id").Limit(100).Find(&rows).Error; err != nil {
		return nil, salesDBError(ctx, err)
	}
	return rows, nil
}
func (r *TenantOnboardingRepository) Request(ctx context.Context, id string) (*model.TenantRegistration, *apperrors.Error) {
	var row model.TenantRegistration
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&row).Error; err != nil {
		return nil, salesDBError(ctx, err)
	}
	return &row, nil
}
func (r *TenantOnboardingRepository) SaveRequest(ctx context.Context, row *model.TenantRegistration, create bool) *apperrors.Error {
	var err error
	if create {
		err = r.db.WithContext(ctx).Create(row).Error
	} else {
		err = r.db.WithContext(ctx).Save(row).Error
	}
	if err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}
func (r *TenantOnboardingRepository) Tenant(ctx context.Context, id string) (*model.Tenant, *apperrors.Error) {
	return NewSalesRepository(r.db).Tenant(ctx, id)
}
func (r *TenantOnboardingRepository) Owner(ctx context.Context, tenant, user string) (*model.Tenant, *apperrors.Error) {
	member, err := NewSalesRepository(r.db).Member(ctx, tenant, user)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, apperrors.ErrForbidden(ctx).SetDetail("tenant membership required")
	}
	var row model.Tenant
	dbErr := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", tenant).First(&row).Error
	if dbErr != nil {
		return nil, salesDBError(ctx, dbErr)
	}
	if row.OwnerID != user || row.Status != model.TenantStatusActive {
		return nil, apperrors.ErrForbidden(ctx).SetDetail("active tenant owner required")
	}
	return &row, nil
}
func (r *TenantOnboardingRepository) Provision(ctx context.Context, row *model.Tenant) *apperrors.Error {
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return salesDBError(ctx, err)
	}
	if err := r.db.WithContext(ctx).Create(&model.TenantMember{TenantID: row.ID, UserID: row.OwnerID, Role: model.RoleOwner}).Error; err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}
func (r *TenantOnboardingRepository) Invitation(ctx context.Context, id, hash string, lock bool) (*model.TenantInvitation, *apperrors.Error) {
	var row model.TenantInvitation
	q := r.db.WithContext(ctx)
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if id != "" {
		q = q.Where("id = ?", id)
	} else {
		q = q.Where("token_hash = ?", hash)
	}
	if err := q.First(&row).Error; err != nil {
		return nil, salesDBError(ctx, err)
	}
	return &row, nil
}
func (r *TenantOnboardingRepository) Invitations(ctx context.Context, tenant string) ([]model.TenantInvitation, *apperrors.Error) {
	rows := []model.TenantInvitation{}
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenant).Order("created_at DESC, id").Limit(100).Find(&rows).Error; err != nil {
		return nil, salesDBError(ctx, err)
	}
	return rows, nil
}
func (r *TenantOnboardingRepository) SaveInvitation(ctx context.Context, row *model.TenantInvitation, create bool) *apperrors.Error {
	var err error
	if create {
		err = r.db.WithContext(ctx).Create(row).Error
	} else {
		err = r.db.WithContext(ctx).Save(row).Error
	}
	if err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}
func (r *TenantOnboardingRepository) MatchesEmail(ctx context.Context, user, email string) (bool, *apperrors.Error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.AuthIdentities{}).Where("user_id = ? AND LOWER(identifier) = ? AND provider IN ?", user, email, []string{"local", "email", "google"}).Count(&count).Error; err != nil {
		return false, salesDBError(ctx, err)
	}
	return count > 0, nil
}
func (r *TenantOnboardingRepository) AddMember(ctx context.Context, tenant, user string) *apperrors.Error {
	member, err := NewSalesRepository(r.db).Member(ctx, tenant, user)
	if err != nil {
		return err
	}
	if member != nil {
		return nil
	}
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "tenant_id"}, {Name: "user_id"}}, DoNothing: true}).Create(&model.TenantMember{TenantID: tenant, UserID: user, Role: model.RoleMember}).Error; err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}
func (r *TenantOnboardingRepository) ReserveInvitation(ctx context.Context, tenant, email string) *apperrors.Error {
	// Serialize invitation creation on the tenant row; stale pending invitations
	// are revoked before issuing a new single-use link.
	var row model.Tenant
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", tenant).First(&row).Error; err != nil {
		return salesDBError(ctx, err)
	}
	if err := r.db.WithContext(ctx).Model(&model.TenantInvitation{}).Where("tenant_id = ? AND email = ? AND status = ?", tenant, email, "pending").Update("status", "revoked").Error; err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}
func (r *TenantOnboardingRepository) Audit(ctx context.Context, tenant, actor, action, id string, state any) *apperrors.Error {
	return NewSalesRepository(r.db).Audit(ctx, tenant, actor, action, "tenant", id, state)
}
func (r *TenantOnboardingRepository) SlugTaken(ctx context.Context, slug string) (bool, *apperrors.Error) {
	var row model.Tenant
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&row).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, salesDBError(ctx, err)
	}
	return true, nil
}
