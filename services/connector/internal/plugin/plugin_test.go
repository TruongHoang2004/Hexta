package plugin

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fixture struct{ BasePlugin }

func (fixture) Platform() string           { return "fixture" }
func (fixture) Capabilities() Capabilities { return Capabilities{Fetch: true} }
func (fixture) Fetch(context.Context, FetchRequest) (Page, error) {
	return Page{NextCursor: "next"}, nil
}
func TestContract(t *testing.T) {
	b := BasePlugin{}
	if !errors.Is(b.Subscribe(context.Background(), ""), ErrUnsupported) {
		t.Fatal("false success")
	}
	if _, err := b.Parse(nil, time.Now()); err != ErrUnsupported {
		t.Fatal(err)
	}
	r, err := NewRegistry(fixture{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = r.Webhook("fixture"); err != ErrUnsupported {
		t.Fatal(err)
	}
	p, err := fixture{}.Fetch(context.Background(), FetchRequest{})
	if err != nil || p.NextCursor != "next" {
		t.Fatal("override")
	}
	if _, err = NewRegistry(fixture{}, fixture{}); err == nil {
		t.Fatal("duplicate")
	}
}
