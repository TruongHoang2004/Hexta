package model

import (
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/events/delivery"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/messaging"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/order"
)

// Outbox event types name shared events (packages/shared/events); the relay resolves each one's single topic.
const (
	DomainEventOrderConfirmed      = order.OrderConfirmedEventType
	DomainEventOutboundIntent      = delivery.SendRequestedEventType
	DomainEventConversationChanged = messaging.ConversationChangedEventType

	DomainEventStatePending   = "pending"
	DomainEventStateLeased    = "leased"
	DomainEventStatePublished = "published"
	DomainEventStateFailed    = "failed"
	// DomainEventStateCancelled marks an event withdrawn before the relay leased it (a bot send cancelled because
	// control changed); it is never published. failed stays reserved for relay give-up.
	DomainEventStateCancelled = "cancelled"
)

// DomainEvent is a transactional outbox row, committed with the business change it describes.
type DomainEvent struct {
	ID          string     `gorm:"primaryKey;type:uuid;column:id" json:"id"`
	TenantID    string     `gorm:"column:tenant_id;type:varchar(36);not null" json:"tenant_id"`
	EventType   string     `gorm:"column:event_type;type:varchar(64);not null" json:"event_type"`
	AggregateID string     `gorm:"column:aggregate_id;type:varchar(36);not null" json:"aggregate_id"`
	Payload     string     `gorm:"column:payload;type:jsonb;not null" json:"payload"`
	State       string     `gorm:"column:state;type:varchar(20);not null;default:'pending'" json:"state"`
	Attempts    int        `gorm:"column:attempts;not null;default:0" json:"attempts"`
	Fence       int64      `gorm:"column:fence;not null;default:0" json:"fence"`
	AvailableAt time.Time  `gorm:"column:available_at;not null;default:clock_timestamp()" json:"available_at"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null;default:clock_timestamp()" json:"created_at"`
	PublishedAt *time.Time `gorm:"column:published_at" json:"published_at,omitempty"`
}

func (DomainEvent) TableName() string {
	return "domain_events"
}
