package messagingmodel

import "time"

type Thread struct {
	ID           string
	TenantID     string
	ExternalKey  string
	DisplayName  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ConnectionID *string
	RecipientID  *string
}

func (Thread) TableName() string { return "messaging_threads" }

type Turn struct {
	ID                string
	TenantID          string
	ThreadID          string
	Lifecycle         string
	Control           string
	AssigneeID        *string
	NeedsHuman        bool
	NeedsHumanReason  *string
	ContextRevision   int64
	ControlGeneration int64
	Version           int64
	LastActivityAt    time.Time
	ClosedAt          *time.Time
	CloseCause        *string
	ClosedBy          *string
	Outcome           *string
	Note              *string
	LastHandlerID     *string
}

func (Turn) TableName() string { return "messaging_turns" }

type Message struct {
	ID          string
	TenantID    string
	ThreadID    string
	TurnID      string
	Direction   string
	ActorID     string
	Body        string
	RequestKey  string
	RequestHash string
	Status      string
	CreatedAt   time.Time
	ContentType string
	// PlaceholderMeta is jsonb {"media_kinds":[...]} for unsupported_media rows.
	PlaceholderMeta     *string `gorm:"type:jsonb"`
	DeliveryStatus      string
	DeliveryErrorReason *string
}

func (Message) TableName() string { return "messaging_messages" }

type Connection struct {
	ID               string
	TenantID         string
	Version          int64
	Platform         string
	AccountID        string
	Enabled          bool
	AccountName      string
	AccountUsername  string
	Status           string
	LastErrorCode    *string
	LastErrorMessage *string
	UpdatedAt        time.Time
}

func (Connection) TableName() string { return "messaging_connections" }

type OAuthState struct {
	ID                    string
	TenantID              string
	ActorID               string
	Platform              string
	StateHash             string
	RedirectURI           string
	ExpiresAt             time.Time
	ConsumedAt            *time.Time
	OAuthSessionID        *string    `gorm:"column:oauth_session_id"`
	OAuthSessionExpiresAt *time.Time `gorm:"column:oauth_session_expires_at"`
	Pages                 *string    `gorm:"type:jsonb"`
	CreatedAt             time.Time
}

func (OAuthState) TableName() string { return "channel_oauth_states" }

type OutboundIntent struct {
	ID              string
	TenantID        string
	ConnectionID    string
	MessageID       string
	ClientMessageID string
	RequestHash     string
	RecipientID     string
	Status          string
	ErrorCategory   *string
	ErrorCode       *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (OutboundIntent) TableName() string { return "outbound_intents" }

type OutboundAttempt struct {
	ID                string
	TenantID          string
	IntentID          string
	ConnectionVersion int64
	Status            string
	ProviderMessageID *string
	CreatedAt         time.Time
	CompletedAt       *time.Time
}

func (OutboundAttempt) TableName() string { return "outbound_attempts" }

// ConversationView is the read model row for one conversation and its latest turn.
type ConversationView struct {
	TenantID          string
	ThreadID          string
	ExternalKey       string
	DisplayName       string
	ThreadCreatedAt   time.Time
	ThreadUpdatedAt   time.Time
	ConnectionID      *string
	RecipientID       *string
	TurnID            string
	Lifecycle         string
	Control           string
	AssigneeID        *string
	NeedsHuman        bool
	ContextRevision   int64
	ControlGeneration int64
	TurnVersion       int64
	LastActivityAt    time.Time
	NeedsHumanReason  *string
	ClosedAt          *time.Time
	CloseCause        *string
	Outcome           *string
}

func (ConversationView) TableName() string { return "messaging_conversation_views" }

// SessionStats is one turn's message count and first message, read for the session list.
type SessionStats struct {
	TurnID         string
	MessageCount   int64
	FirstMessageID string
	OpenedAt       time.Time
}

// SessionOrder is an order linked to a turn, read for the session list.
type SessionOrder struct {
	ID              string
	OrderNumber     string
	Status          string
	MessagingTurnID string
}
