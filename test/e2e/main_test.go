package e2e

import (
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	log.Println("Starting e2e tests setup...")

	// Assuming Docker Compose runs API on localhost:9090
	// Wait here if needed or set envs

	exitCode := m.Run()

	log.Println("Tearing down e2e tests...")
	os.Exit(exitCode)
}
