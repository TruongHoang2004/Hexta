package e2e

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"gitlab.com/ecommercehub1/test/config"

	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/assert"
)

// Configurable constants for the E2E
var (
	apiURL = config.BaseURL
)

var (
	e *httpexpect.Expect
)

func initExpect(t *testing.T) *httpexpect.Expect {
	return httpexpect.Default(t, apiURL)
}

func generateRandomString(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

func TestE2E_AuthFlow(t *testing.T) {
	e = initExpect(t)

	// Variables to store across scopes
	var accessToken string
	var refreshToken string
	userName := generateRandomString("testuser")
	email := fmt.Sprintf("%s@example.com", userName)
	password := "TestPass123!"

	t.Run("1. Register User", func(t *testing.T) {
		payload := map[string]interface{}{
			"user_name":     userName,
			"full_name":     "Test E2E User",
			"email":         email,
			"password":      password,
			"gender":        "male",
			"phone":         fmt.Sprintf("0123%06d", time.Now().UnixNano()%1000000),
			"date_of_birth": "1990-01-01T00:00:00Z",
		}

		res := e.POST("/auth/register").
			WithJSON(payload).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		res.Value("data").Object().Value("user_name").String().IsEqual(userName)
		res.Value("data").Object().Value("email").String().IsEqual(email)
	})

	t.Run("2. Login User", func(t *testing.T) {
		payload := map[string]interface{}{
			"user_name": userName,
			"password":  password,
		}

		res := e.POST("/auth/login").
			WithJSON(payload).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		data := res.Value("data").Object()
		data.ContainsKey("access_token")
		data.ContainsKey("refresh_token")

		accessToken = data.Value("access_token").String().Raw()
		refreshToken = data.Value("refresh_token").String().Raw()

		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
	})

	t.Run("3. Validate Token", func(t *testing.T) {
		// Validating Authorization header
		res := e.GET("/auth/validate-token").
			WithHeader("Authorization", "Bearer "+accessToken).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		res.Value("data").Object().ContainsKey("user_id")
		res.Value("data").Object().ContainsKey("session_id")
	})

	t.Run("4. Validate Token - Error Without Header", func(t *testing.T) {
		e.GET("/auth/validate-token").
			Expect().
			Status(http.StatusUnauthorized)
	})

	t.Run("5. Refresh Token", func(t *testing.T) {
		payload := map[string]interface{}{
			"refresh_token": refreshToken,
		}

		res := e.POST("/auth/refresh").
			WithJSON(payload).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		data := res.Value("data").Object()
		data.ContainsKey("access_token")
		data.ContainsKey("refresh_token")

		newAccessToken := data.Value("access_token").String().Raw()
		newRefreshToken := data.Value("refresh_token").String().Raw()

		// The tokens should be different due to the rotation policy
		fmt.Println("Old access token: ", accessToken)
		fmt.Println("New access token: ", newAccessToken)
		fmt.Println("Old refresh token: ", refreshToken)
		fmt.Println("New refresh token: ", newRefreshToken)
		assert.NotEqual(t, accessToken, newAccessToken)
		assert.NotEqual(t, refreshToken, newRefreshToken)

		// Overwrite access token for next tests if needed
		accessToken = newAccessToken
	})

	t.Run("6. User Profile (Me)", func(t *testing.T) {
		res := e.GET("/users/me").
			WithHeader("Authorization", "Bearer "+accessToken).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		data := res.Value("data").Object()
		data.Value("user_name").String().IsEqual(userName)
		data.Value("email").String().IsEqual(email)
	})

	t.Run("7. List Users", func(t *testing.T) {
		res := e.GET("/users/").
			WithHeader("Authorization", "Bearer "+accessToken).
			WithQuery("page", 1).
			WithQuery("size", 10).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		data := res.Value("data").Object()
		data.ContainsKey("data")
		data.ContainsKey("total")
		data.ContainsKey("page")
		data.ContainsKey("size")

		items := data.Value("data").Array()
		items.Length().Ge(1) // Should have at least the one we just registered
	})
}
