// Package connection holds channel connection lifecycle events.
package connection

import (
	"errors"

	"github.com/TruongHoang2004/Hexta/packages/shared/events"
)

// HealthChangedTopic carries HealthChanged. Producer: connector (outbox). Consumer: API channel status.
const HealthChangedTopic = "hexta.connection.health-changed.v1"

// StatusAuthExpired is currently the only reported health status: the provider rejected the stored credential.
const StatusAuthExpired = "auth_expired"

// HealthChanged reports that the connector observed a connection health change.
type HealthChanged struct {
	events.Meta
	ConnectionID      string `json:"connection_id"`
	ConnectionVersion int64  `json:"connection_version"`
	Status            string `json:"status"`
	ErrorCode         string `json:"error_code"`
}

// Key orders health changes per connection.
func (e HealthChanged) Key() string { return e.ConnectionID }

func (e HealthChanged) Validate() error {
	if err := e.Meta.Validate(); err != nil {
		return err
	}
	if !events.ValidUUID(e.TenantID, e.ConnectionID) || e.ConnectionVersion < 1 || e.Status != StatusAuthExpired || len(e.ErrorCode) > 64 {
		return errors.New("invalid connection health change")
	}
	return nil
}
