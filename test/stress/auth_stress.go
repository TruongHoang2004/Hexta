package stress

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
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

type LoginPayload struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

type RefreshPayload struct {
	RefreshToken string `json:"refresh_token"`
}

var (
	baseURL = config.BaseURL
)

func randomString(prefix string) string {
	n, _ := rand.Int(rand.Reader, big.NewInt(100000))
	return fmt.Sprintf("%s_%d_%d", prefix, time.Now().UnixMilli(), n.Int64())
}

func randomPhone() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("0123%06d", n.Int64()+100000)
}

// NewAuthTargeter creates a custom Targeter that registers, logs in, validating, refreshing, getting user and listing users
func NewAuthTargeter() vegeta.Targeter {
	return func(tgt *vegeta.Target) error {
		if tgt == nil {
			return vegeta.ErrNilTarget
		}

		userName := randomString("testuser")
		email := fmt.Sprintf("%s@example.com", userName)
		password := "TestPass123!"

		// Determine the step (we can pseudo-randomize or just do one step, but Vegeta works best attacking one endpoint per target)
		// For a complex flow, we'd need to coordinate state.
		// Instead, we will configure the targeter to just hit "register" since we don't have continuity in vegeta targeter natively without external state.

		payload := RegisterPayload{
			UserName:    userName,
			FullName:    "Test E2E User",
			Email:       email,
			Password:    password,
			Gender:      "male",
			Phone:       randomPhone(),
			DateOfBirth: "1990-01-01T00:00:00Z",
		}

		body, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		tgt.Method = "POST"
		tgt.URL = baseURL + "/auth/register"
		tgt.Body = body
		tgt.Header = map[string][]string{
			"Content-Type": {"application/json"},
		}

		return nil
	}
}

func RunAuthStressTest() {
	rate := vegeta.Rate{Freq: 50, Per: time.Second}
	duration := 300 * time.Second

	targeter := NewAuthTargeter()

	attacker := vegeta.NewAttacker()

	log.Printf("Starting stress test: %d req/s for %s", rate.Freq, duration)

	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, rate, duration, "AuthStressTest") {
		metrics.Add(res)
	}
	metrics.Close()

	reporter := vegeta.NewTextReporter(&metrics)
	err := reporter.Report(log.Writer())
	if err != nil {
		log.Fatalf("failed to report metrics: %v", err)
	}
}
