package utils_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.com/ecommercehub1/api/utils"
)

type TestPayload struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

func TestGenerateAndDecodeJWT(t *testing.T) {
	secret := "my-secret-key-12345"
	payload := TestPayload{
		UserID: "user-123",
		Role:   "admin",
	}

	// 1. Valid token
	token, err := utils.GenerateJWT(payload, secret, 60)
	assert.Nil(t, err)
	assert.NotEmpty(t, token)

	claims, err := utils.DecodeJWT[TestPayload](token, secret)
	assert.Nil(t, err)
	assert.Equal(t, "user-123", claims.Payload.UserID)
	assert.Equal(t, "admin", claims.Payload.Role)

	// 2. Decode with wrong secret
	_, err = utils.DecodeJWT[TestPayload](token, "wrong-secret")
	assert.NotNil(t, err)

	// 3. Expired token (1 second)
	shortLivedToken, err := utils.GenerateJWT(payload, secret, 1)
	assert.Nil(t, err)
	time.Sleep(1200 * time.Millisecond)

	_, err = utils.DecodeJWT[TestPayload](shortLivedToken, secret)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "expired")
}
