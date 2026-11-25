package persistencemap

import (
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/ordercasemodel"
	"github.com/stretchr/testify/require"
)

func TestUnknownStoredValuesFailInsteadOfInventingState(t *testing.T) {
	_, err := Case(ordercasemodel.Case{CaseType: "refund", Status: types.OrderCaseStatusOpen}, nil, nil)
	require.ErrorIs(t, err, ErrUnknownValue)
	_, err = Case(ordercasemodel.Case{CaseType: types.OrderCaseTypeReturn, Status: "reopened"}, nil, nil)
	require.ErrorIs(t, err, ErrUnknownValue)
	_, err = Case(ordercasemodel.Case{CaseType: types.OrderCaseTypeReturn, Status: types.OrderCaseStatusOpen}, nil, []ordercasemodel.Event{{Action: "deleted", Status: types.OrderCaseStatusOpen}})
	require.ErrorIs(t, err, ErrUnknownValue)
}

func TestEventLinesRoundTripAndFinalizerIsPreserved(t *testing.T) {
	at := time.Date(2026, 11, 26, 1, 2, 3, 456789000, time.UTC)
	lines := []types.OrderCaseLine{{OrderItemID: "item", OriginalQuantity: 2, Quantity: 1}}
	row := EventRow("tenant", "case", types.OrderCaseEvent{ID: "e", ActorID: "owner", Action: types.OrderCaseActionFinalized, Status: types.OrderCaseStatusFinalized, Version: 3, Lines: lines, CreatedAt: at})
	back, err := Event(row)
	require.NoError(t, err)
	require.Equal(t, lines, back.Lines)
	require.Equal(t, at, back.CreatedAt)
	// An event without lines stores an empty array, never null.
	require.Equal(t, "[]", EventRow("tenant", "case", types.OrderCaseEvent{}).Lines)

	owner := "owner"
	value := types.OrderCase{ID: "case", Type: types.OrderCaseTypeReturn, Status: types.OrderCaseStatusFinalized, FinalizedBy: &owner, FinalizedAt: &at, Version: 3}
	mapped, err := Case(CaseRow(value), nil, nil)
	require.NoError(t, err)
	require.Equal(t, &owner, mapped.FinalizedBy)
	require.Equal(t, &at, mapped.FinalizedAt)
	require.Empty(t, mapped.Lines)
	require.NotNil(t, mapped.Lines)
}
