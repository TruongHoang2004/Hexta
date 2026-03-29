package mock

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"time"

	"gitlab.com/ecommercehub1/test/config"
)

type RegisterPayload struct {
	UserName    string `json:"user_name"`
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Gender      string `json:"gender"`
	Phone       string `json:"phone"`
	DateOfBirth string `json:"date_of_birth"`
}

var baseURL = config.BaseURL

func randomString(prefix string) string {
	n, _ := rand.Int(rand.Reader, big.NewInt(100000))
	return fmt.Sprintf("%s_%d_%d", prefix, time.Now().UnixMilli(), n.Int64())
}

func randomPhone() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("0123%06d", n.Int64()+100000)
}

func RegisterMockUsers(count int) {
	log.Printf("Registering %d mock users...", count)

	for i := 0; i < count; i++ {
		userName := randomString("mockuser")
		payload := RegisterPayload{
			UserName:    userName,
			FullName:    fmt.Sprintf("Mock User %d", i+1),
			Email:       fmt.Sprintf("%s@example.com", userName),
			Password:    "TestPass123!",
			Gender:      "male",
			Phone:       randomPhone(),
			DateOfBirth: "1990-01-01T00:00:00Z",
		}

		body, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Failed to marshal payload for %s: %v", userName, err)
			continue
		}

		resp, err := http.Post(baseURL+"/auth/register", "application/json", bytes.NewBuffer(body))
		if err != nil {
			log.Printf("Failed to register %s: %v", userName, err)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			log.Printf("[%d/%d] Successfully registered mock user: %s", i+1, count, userName)
		} else {
			log.Printf("[%d/%d] Failed to register %s, status code: %d", i+1, count, userName, resp.StatusCode)
		}

		resp.Body.Close()
	}
	log.Println("Mock user registration complete.")
}
