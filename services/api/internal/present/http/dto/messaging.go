package dto

import "time"

type MessagingQuery struct {
	Cursor string `form:"cursor" validate:"omitempty,uuid"`
	Filter string `form:"filter" validate:"omitempty,oneof=all mine unassigned needs_human"`
	Limit  int    `form:"limit" validate:"min=1,max=100"`
}
type MessagingInboundRequest struct {
	ExternalKey string `json:"external_key" validate:"required,max=120"`
	DisplayName string `json:"display_name" validate:"required,max=120"`
	Body        string `json:"body" validate:"required,max=4000"`
	RequestKey  string `json:"request_key" validate:"required,uuid"`
}
type MessagingCommandRequest struct {
	ExpectedVersion int64   `json:"expected_version" validate:"min=1"`
	AssigneeID      *string `json:"assignee_id" validate:"omitempty,min=1,max=255"`
}
type MessagingSendRequest struct {
	ExpectedVersion int64  `json:"expected_version" validate:"min=1"`
	ContextRevision int64  `json:"context_revision" validate:"min=1"`
	Body            string `json:"body" validate:"required,max=4000"`
	RequestKey      string `json:"request_key" validate:"required,uuid"`
}
type MessagingThread struct {
	ID          string    `json:"id"`
	DisplayName string    `json:"display_name"`
	ExternalKey string    `json:"external_key"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
type MessagingTurn struct {
	ID                string    `json:"id"`
	ThreadID          string    `json:"thread_id"`
	Lifecycle         string    `json:"lifecycle"`
	Control           string    `json:"control"`
	AssigneeID        *string   `json:"assignee_id"`
	NeedsHuman        bool      `json:"needs_human"`
	ContextRevision   int64     `json:"context_revision"`
	ControlGeneration int64     `json:"control_generation"`
	Version           int64     `json:"version"`
	LastActivityAt    time.Time `json:"last_activity_at"`
}
type MessagingMessage struct {
	ID        string    `json:"id"`
	ThreadID  string    `json:"thread_id"`
	TurnID    string    `json:"turn_id"`
	Direction string    `json:"direction"`
	ActorID   string    `json:"actor_id"`
	Body      string    `json:"body"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type MessagingConversation struct {
	Thread MessagingThread `json:"thread"`
	Turn   *MessagingTurn  `json:"turn"`
	Source string          `json:"source"`
}
type MessagingPage struct {
	Items      []MessagingConversation `json:"items"`
	NextCursor string                  `json:"next_cursor"`
}
type MessagingDetail struct {
	Conversation MessagingConversation `json:"conversation"`
	Messages     []MessagingMessage    `json:"messages"`
	NextCursor   string                `json:"next_cursor"`
}
