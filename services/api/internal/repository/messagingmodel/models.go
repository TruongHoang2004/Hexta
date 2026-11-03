package messagingmodel

import "time"

type Thread struct {
	ID          string
	TenantID    string
	ExternalKey string
	DisplayName string
	CreatedAt   time.Time
	UpdatedAt   time.Time
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
	ContextRevision   int64
	ControlGeneration int64
	Version           int64
	LastActivityAt    time.Time
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
}

func (Message) TableName() string { return "messaging_messages" }
