package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"unicode/utf8"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/google/uuid"
)

func messagingText(value string, max int) bool {
	return strings.TrimSpace(value) != "" && utf8.ValidString(value) && utf8.RuneCountInString(value) <= max
}
func messagingKey(ctx context.Context, key string) *apperrors.Error {
	if _, err := uuid.Parse(key); err != nil {
		return apperrors.ErrBadRequest(ctx).SetDetail("request_key must be a UUID")
	}
	return nil
}
func messagingID() string { return uuid.Must(uuid.NewV7()).String() }
func messagingHash(value any) string {
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
