package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"unicode/utf8"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
)

type WorkflowService struct {
	repo *repository.WorkflowRepository
}

func NewWorkflowService(repo *repository.WorkflowRepository) *WorkflowService {
	return &WorkflowService{repo: repo}
}

func (s *WorkflowService) authorized(ctx context.Context, tenant, actor string, fn func(*repository.WorkflowRepository, bool) *apperrors.Error) *apperrors.Error {
	if actor == "" {
		return apperrors.ErrUnauthorized(ctx)
	}
	if _, err := uuid.Parse(tenant); err != nil {
		return apperrors.ErrBadRequest(ctx).SetDetail("valid tenant ID required")
	}
	return s.repo.Within(ctx, func(r *repository.WorkflowRepository) *apperrors.Error {
		workspace, err := r.Tenant(ctx, tenant)
		if err != nil {
			return err
		}
		if workspace.Status != model.TenantStatusActive {
			return apperrors.ErrForbidden(ctx).SetDetail("tenant is not active")
		}
		member, err := r.Member(ctx, tenant, actor)
		if err != nil {
			return err
		}
		if member == nil {
			return apperrors.ErrForbidden(ctx).SetDetail("tenant membership required")
		}
		return fn(r, workspace.OwnerID == actor)
	})
}
func workflowHash(input any) string {
	data, _ := json.Marshal(input)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
func workflowOwner(ctx context.Context, owner bool) *apperrors.Error {
	if !owner {
		return apperrors.ErrForbidden(ctx).SetDetail("tenant owner required")
	}
	return nil
}
func workflowMember(ctx context.Context, r *repository.WorkflowRepository, tenant, user string) *apperrors.Error {
	member, err := r.Member(ctx, tenant, user)
	if err != nil {
		return err
	}
	if member == nil {
		return apperrors.ErrBadRequest(ctx).SetDetail("recipient must be a current tenant member")
	}
	return nil
}
func workflowInput(ctx context.Context, input *types.WorkflowInput) *apperrors.Error {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || utf8.RuneCountInString(input.Name) > 120 || len(input.Steps) < 1 || len(input.Steps) > 30 {
		return apperrors.ErrBadRequest(ctx).SetDetail("name and 1–30 workflow steps required")
	}
	for i := range input.Steps {
		step := &input.Steps[i]
		step.Name = strings.TrimSpace(step.Name)
		if step.Name == "" || utf8.RuneCountInString(step.Name) > 120 || (step.HandoffMode != "direct" && step.HandoffMode != "acceptance") {
			return apperrors.ErrBadRequest(ctx).SetDetail("each step needs a name and direct or acceptance handoff mode")
		}
	}
	return nil
}
func makeWorkflowVersion(tenant, id string, number int64, input types.WorkflowInput) *repository.WorkflowVersion {
	row := &repository.WorkflowVersion{ID: uuid.NewString(), TenantID: tenant, DefinitionID: id, Number: number, Name: input.Name, Steps: []repository.WorkflowStep{}}
	for i, step := range input.Steps {
		row.Steps = append(row.Steps, repository.WorkflowStep{ID: uuid.NewString(), TenantID: tenant, WorkflowVersionID: row.ID, Ordinal: i, Name: step.Name, HandoffMode: step.HandoffMode})
	}
	return row
}
func (s *WorkflowService) Definitions(ctx context.Context, tenant, actor string, page, size int) ([]repository.WorkflowDefinition, int64, *apperrors.Error) {
	var rows []repository.WorkflowDefinition
	var total int64
	err := s.authorized(ctx, tenant, actor, func(r *repository.WorkflowRepository, _ bool) *apperrors.Error {
		limit, offset := pageBounds(page, size)
		var err *apperrors.Error
		rows, total, err = r.Definitions(ctx, tenant, limit, offset)
		return err
	})
	return rows, total, err
}
func (s *WorkflowService) Publish(ctx context.Context, tenant, actor, id, key string, input types.WorkflowInput) (*repository.WorkflowDefinition, *apperrors.Error) {
	if err := workflowInput(ctx, &input); err != nil {
		return nil, err
	}
	create := id == ""
	if create {
		parsed, err := uuid.Parse(key)
		if err != nil {
			return nil, apperrors.ErrBadRequest(ctx).SetDetail("Idempotency-Key must be a UUID")
		}
		id = parsed.String()
		if input.ExpectedVersion != 0 {
			return nil, apperrors.ErrBadRequest(ctx).SetDetail("new workflow expected_version must be zero")
		}
	}
	var row *repository.WorkflowDefinition
	err := s.authorized(ctx, tenant, actor, func(r *repository.WorkflowRepository, owner bool) *apperrors.Error {
		if err := workflowOwner(ctx, owner); err != nil {
			return err
		}
		if create {
			row = &repository.WorkflowDefinition{ID: id, TenantID: tenant, Name: input.Name, Version: 0, CreateHash: workflowHash(input)}
			inserted, err := r.InsertDefinition(ctx, row)
			if err != nil {
				return err
			}
			if !inserted {
				row, err = r.Definition(ctx, tenant, id)
				if err != nil {
					return err
				}
				if row.CreateHash != workflowHash(input) {
					return apperrors.ErrConflict(ctx, "idempotency key", "already used for different workflow data")
				}
				return nil
			}
		} else {
			var err *apperrors.Error
			row, err = r.Definition(ctx, tenant, id)
			if err != nil {
				return err
			}
		}
		if row.Version != input.ExpectedVersion {
			return apperrors.ErrConflict(ctx, "workflow", "changed; reload before publishing")
		}
		row.Name = input.Name
		row.Version++
		version := makeWorkflowVersion(tenant, id, row.Version, input)
		if err := r.Publish(ctx, row, version, input.ExpectedVersion); err != nil {
			return err
		}
		row.Versions = []repository.WorkflowVersion{*version}
		return r.Audit(ctx, tenant, actor, "workflow.published", "workflow", id, version)
	})
	return row, err
}
func (s *WorkflowService) Items(ctx context.Context, tenant, actor string, page, size int) ([]repository.WorkItem, int64, *apperrors.Error) {
	var rows []repository.WorkItem
	var total int64
	err := s.authorized(ctx, tenant, actor, func(r *repository.WorkflowRepository, _ bool) *apperrors.Error {
		limit, offset := pageBounds(page, size)
		var err *apperrors.Error
		rows, total, err = r.WorkItems(ctx, tenant, limit, offset)
		return err
	})
	return rows, total, err
}
func (s *WorkflowService) Item(ctx context.Context, tenant, actor, id string) (*repository.WorkItem, *apperrors.Error) {
	var row *repository.WorkItem
	err := s.authorized(ctx, tenant, actor, func(r *repository.WorkflowRepository, _ bool) *apperrors.Error {
		var err *apperrors.Error
		row, err = r.WorkItem(ctx, tenant, id)
		return err
	})
	return row, err
}
func (s *WorkflowService) Create(ctx context.Context, tenant, actor, key string, input types.WorkInput) (*repository.WorkItem, *apperrors.Error) {
	parsed, err := uuid.Parse(key)
	if err != nil {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("Idempotency-Key must be a UUID")
	}
	key = parsed.String()
	var row *repository.WorkItem
	failure := s.authorized(ctx, tenant, actor, func(r *repository.WorkflowRepository, owner bool) *apperrors.Error {
		if err := workflowOwner(ctx, owner); err != nil {
			return err
		}
		// Lock the source order before key lookup so identical concurrent creates serialize.
		order, err := r.Order(ctx, tenant, input.OrderID)
		if err != nil {
			return err
		}
		row, err = r.WorkByKey(ctx, tenant, actor, key)
		if err != nil {
			return err
		}
		if row != nil {
			if row.CreateHash != workflowHash(input) {
				return apperrors.ErrConflict(ctx, "idempotency key", "already used for different work data")
			}
			return nil
		}
		if order.Status != model.OrderStatusConfirmed {
			return apperrors.ErrConflict(ctx, "order", "must be confirmed before creating work")
		}
		if err := workflowMember(ctx, r, tenant, input.ResponsibleID); err != nil {
			return err
		}
		version, err := r.WorkflowVersion(ctx, tenant, input.WorkflowVersionID)
		if err != nil {
			return err
		}
		if len(version.Steps) == 0 {
			return apperrors.ErrConflict(ctx, "workflow", "has no published steps")
		}
		row = &repository.WorkItem{ID: uuid.NewString(), TenantID: tenant, OrderID: order.ID, WorkflowVersionID: version.ID, CurrentStep: 0, ResponsibleID: &input.ResponsibleID, State: "active", Version: 1, CreatedBy: actor, RequestKey: key, CreateHash: workflowHash(input), Workflow: *version}
		if err := r.SaveWork(ctx, row, 0, true); err != nil {
			return err
		}
		return appendWorkEvent(ctx, r, actor, "created", nil, 0, row)
	})
	return row, failure
}
func appendWorkEvent(ctx context.Context, r *repository.WorkflowRepository, actor, action string, from *string, step int, row *repository.WorkItem) *apperrors.Error {
	event := repository.WorkEvent{ID: uuid.NewString(), TenantID: row.TenantID, WorkItemID: row.ID, Version: row.Version, ActorID: actor, Action: action, FromResponsibleID: from, ResponsibleID: row.ResponsibleID, RecipientID: row.RecipientID, FromStep: step, CurrentStep: row.CurrentStep, State: row.State}
	if err := r.WorkEvent(ctx, &event); err != nil {
		return err
	}
	row.Events = append([]repository.WorkEvent{event}, row.Events...)
	if len(row.Events) > 100 {
		row.Events = row.Events[:100]
	}
	return nil
}
func (s *WorkflowService) Command(ctx context.Context, tenant, actor, id, action string, input types.WorkCommand) (*repository.WorkItem, *apperrors.Error) {
	var row *repository.WorkItem
	err := s.authorized(ctx, tenant, actor, func(r *repository.WorkflowRepository, owner bool) *apperrors.Error {
		var err *apperrors.Error
		row, err = r.WorkItem(ctx, tenant, id)
		if err != nil {
			return err
		}
		if row.Version != input.ExpectedVersion || row.State == "completed" {
			return apperrors.ErrConflict(ctx, "work item", "changed or completed; reload before retrying")
		}
		from, step := row.ResponsibleID, row.CurrentStep
		if action == "assign" {
			if err := workflowOwner(ctx, owner); err != nil {
				return err
			}
			if input.ResponsibleID != nil {
				if err := workflowMember(ctx, r, tenant, *input.ResponsibleID); err != nil {
					return err
				}
			}
			row.ResponsibleID = input.ResponsibleID
			row.RecipientID = nil
			row.State = "active"
		} else if action == "accept" {
			if row.State != "waiting" || row.RecipientID == nil || *row.RecipientID != actor {
				return apperrors.ErrForbidden(ctx).SetDetail("only the nominated recipient can accept a pending handoff")
			}
			row.ResponsibleID = row.RecipientID
			row.RecipientID = nil
			row.CurrentStep++
			row.State = "active"
		} else {
			if row.ResponsibleID == nil || *row.ResponsibleID != actor {
				return apperrors.ErrForbidden(ctx).SetDetail("only the current responsible member can progress work")
			}
			if row.State != "active" {
				return apperrors.ErrConflict(ctx, "handoff", "awaiting acceptance; owner may reassign to cancel it")
			}
			if action == "complete" {
				if row.CurrentStep != len(row.Workflow.Steps)-1 {
					return apperrors.ErrConflict(ctx, "work item", "complete each preceding step first")
				}
				row.State = "completed"
			} else if action == "handoff" {
				if row.CurrentStep >= len(row.Workflow.Steps)-1 {
					return apperrors.ErrConflict(ctx, "work item", "final step must be completed instead of handed off")
				}
				if input.RecipientID == "" || input.RecipientID == actor {
					return apperrors.ErrBadRequest(ctx).SetDetail("nominate another current tenant member")
				}
				if err := workflowMember(ctx, r, tenant, input.RecipientID); err != nil {
					return err
				}
				if row.Workflow.Steps[row.CurrentStep].HandoffMode == "acceptance" {
					row.RecipientID = &input.RecipientID
					row.State = "waiting"
				} else {
					row.ResponsibleID = &input.RecipientID
					row.CurrentStep++
				}
			} else {
				return apperrors.ErrBadRequest(ctx).SetDetail("unknown work command")
			}
		}
		row.Version++
		if err := r.SaveWork(ctx, row, input.ExpectedVersion, false); err != nil {
			return err
		}
		return appendWorkEvent(ctx, r, actor, action, from, step, row)
	})
	return row, err
}
