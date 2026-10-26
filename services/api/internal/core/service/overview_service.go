package service

import (
	"context"
	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
	"unicode/utf8"
)

type OverviewService struct {
	repo *repository.OverviewRepository
}

func NewOverviewService(repo *repository.OverviewRepository) *OverviewService {
	return &OverviewService{repo: repo}
}
func (s *OverviewService) Read(ctx context.Context, tenant, actor string, query types.OverviewQuery) (*types.OverviewResult, *apperrors.Error) {
	if actor == "" {
		return nil, apperrors.ErrUnauthorized(ctx)
	}
	if _, err := uuid.Parse(tenant); err != nil {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("valid tenant ID required")
	}
	if query.Assignment == "" {
		query.Assignment = "all"
	}
	if err := validateOverviewQuery(ctx, query); err != nil {
		return nil, err
	}
	if query.From != nil {
		from, to := query.From.UTC(), query.To.UTC()
		query.From = &from
		query.To = &to
	}
	var result *types.OverviewResult
	failure := s.repo.WithinSnapshot(ctx, func(r *repository.OverviewRepository) *apperrors.Error {
		workspace, err := r.Tenant(ctx, tenant)
		if err != nil {
			return err
		}
		member, err := r.Member(ctx, tenant, actor)
		if err != nil {
			return err
		}
		if workspace.Status != model.TenantStatusActive || !member || workspace.OwnerID != actor {
			return apperrors.ErrForbidden(ctx).SetDetail("active tenant owner required")
		}
		if query.EmployeeID != nil {
			member, err = r.Member(ctx, tenant, *query.EmployeeID)
			if err != nil {
				return err
			}
			if !member {
				return apperrors.ErrBadRequest(ctx).SetDetail("employee must be a current tenant member")
			}
		}
		result, err = r.Read(ctx, tenant, query)
		return err
	})
	if failure != nil {
		return nil, failure
	}
	return result, nil
}

func validateOverviewQuery(ctx context.Context, query types.OverviewQuery) *apperrors.Error {
	invalid := func() *apperrors.Error { return apperrors.ErrBadRequest(ctx).SetDetail("invalid overview filters") }
	if query.Page < 1 || query.Page > 100000 || query.Size < 1 || query.Size > 100 {
		return invalid()
	}
	switch query.Assignment {
	case "all", "assigned", "unassigned":
	default:
		return invalid()
	}
	if query.EmployeeID != nil {
		if *query.EmployeeID == "" || utf8.RuneCountInString(*query.EmployeeID) > 255 || query.Assignment == "unassigned" {
			return invalid()
		}
	}
	if (query.From == nil) != (query.To == nil) {
		return invalid()
	}
	if query.From != nil && !query.From.Before(*query.To) {
		return invalid()
	}
	return nil
}
