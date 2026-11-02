package types

import "time"

type MessagingScope struct{ Active, Member, Owner bool }
type MessagingThread struct {
	ID, TenantID, ExternalKey, DisplayName string
	CreatedAt, UpdatedAt                   time.Time
}
type MessagingTurn struct {
	ID, TenantID, ThreadID, Lifecycle, Control  string
	AssigneeID                                  *string
	NeedsHuman                                  bool
	ContextRevision, ControlGeneration, Version int64
	LastActivityAt                              time.Time
}
type MessagingMessage struct {
	ID, TenantID, ThreadID, TurnID, Direction, ActorID, Body, RequestKey, RequestHash, Status string
	CreatedAt                                                                                 time.Time
}
type MessagingConversation struct {
	Thread MessagingThread
	Turn   *MessagingTurn
}
type MessagingDetail struct {
	Conversation MessagingConversation
	Messages     []MessagingMessage
	NextCursor   string
}
type MessagingPage struct {
	Items      []MessagingConversation
	NextCursor string
}
type MessagingQuery struct {
	Cursor, Filter string
	Limit          int
}
type MessagingInbound struct{ ExternalKey, DisplayName, Body, RequestKey string }
type MessagingCommand struct {
	ExpectedVersion, ContextRevision int64
	AssigneeID                       *string
	Body, RequestKey                 string
}
