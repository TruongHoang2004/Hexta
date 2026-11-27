// Package persistencemap converts knowledge document, sync job and knowledge source projection rows to core values.
package persistencemap

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/knowledgemodel"
)

// ErrStoredTriggers reports an active script version whose triggers column is not a JSON string array.
var ErrStoredTriggers = errors.New("stored script triggers are not a string array")

func text(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func micros(values ...*time.Time) int64 {
	var newest int64
	for _, value := range values {
		if value != nil && value.UnixMicro() > newest {
			newest = value.UnixMicro()
		}
	}
	return newest
}

func Document(row knowledgemodel.Document) types.KnowledgeDocument {
	return types.KnowledgeDocument{ID: row.ID, TenantID: row.TenantID, Title: row.Title, Body: row.Body, Version: row.Version,
		CreatedBy: row.CreatedBy, UpdatedBy: row.UpdatedBy, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt, DeletedAt: row.DeletedAt}
}

func DocumentWithSync(row knowledgemodel.DocumentSync) types.KnowledgeDocument {
	value := Document(row.Document)
	value.SyncState, value.SyncError = text(row.SyncState), text(row.SyncError)
	return value
}

func DocumentRow(value types.KnowledgeDocument) knowledgemodel.Document {
	return knowledgemodel.Document{ID: value.ID, TenantID: value.TenantID, Title: value.Title, Body: value.Body, Version: value.Version,
		CreatedBy: value.CreatedBy, UpdatedBy: value.UpdatedBy, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt, DeletedAt: value.DeletedAt}
}

func Job(row knowledgemodel.SyncJob) types.KnowledgeJob {
	return types.KnowledgeJob{TenantID: row.TenantID, SourceType: row.SourceType, SourceID: row.SourceID, Fence: row.Fence, Attempts: row.Attempts}
}

// Product maps a product and its live variants. The revision is the newest change of the product, a variant or a
// variant's inventory, so any of them yields a newer revision.
func Product(row knowledgemodel.Product, variants []knowledgemodel.Variant, active string) types.KnowledgeProduct {
	value := types.KnowledgeProduct{ID: row.ID, Name: row.Name, Category: text(row.Category), Description: text(row.Description),
		Removed: row.Status != active || row.DeletedAt != nil, Revision: micros(row.UpdatedAt, row.DeletedAt)}
	for _, variant := range variants {
		value.Revision = max(value.Revision, micros(variant.UpdatedAt, variant.InventoryUpdateAt))
		value.Variants = append(value.Variants, types.KnowledgeVariant{SKU: variant.SKU, Name: variant.VariantName, Price: variant.Price,
			PriceProvided: variant.PriceProvided, Stock: variant.AvailableQty})
	}
	return value
}

func Script(row knowledgemodel.Script) (types.KnowledgeScript, error) {
	value := types.KnowledgeScript{ID: row.ID, Name: row.Name, Status: row.Status, ActiveVersionID: text(row.ActiveVersionID),
		Answer: text(row.Answer), Revision: row.Version}
	if row.Triggers != nil {
		if err := json.Unmarshal([]byte(*row.Triggers), &value.Triggers); err != nil {
			return types.KnowledgeScript{}, ErrStoredTriggers
		}
	}
	return value, nil
}
