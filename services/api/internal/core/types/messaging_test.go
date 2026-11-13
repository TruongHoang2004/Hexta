package types

import "testing"

func TestMessagingUnknownControlDoesNotAuthorizeBot(t *testing.T) {
	actor := "sale"
	for _, turn := range []MessagingTurn{
		{Control: "system", AssigneeID: &actor},
		{Control: "system", NeedsHuman: true},
		{Control: "sales", AssigneeID: &actor, NeedsHuman: true},
		{Control: "unexpected"},
	} {
		if turn.HandlingState() != "UNKNOWN" {
			t.Fatalf("unsafe control combination projected as %s", turn.HandlingState())
		}
	}
}
