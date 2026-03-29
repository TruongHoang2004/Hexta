package utils

import (
	"crypto/rand"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	entropyMu sync.Mutex
	entropy   = ulid.Monotonic(rand.Reader, 0)
)

// NewULID generates a lexicographically sortable ULID string.
func NewULID() string {
	entropyMu.Lock()
	defer entropyMu.Unlock()

	return ulid.MustNew(
		ulid.Timestamp(time.Now().UTC()),
		entropy,
	).String()
}
