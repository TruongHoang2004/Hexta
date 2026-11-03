package service

import (
	"context"
	"encoding/json"
	"github.com/TruongHoang2004/Hexta/packages/shared/ingress"
)

type Publisher interface {
	Publish(context.Context, string, []byte) error
}

func Publish(ctx context.Context, p Publisher, payload []byte) error {
	var e ingress.Event
	if err := json.Unmarshal(payload, &e); err != nil {
		return err
	}
	if err := e.Validate(); err != nil {
		return err
	}
	return p.Publish(ctx, e.Key(), payload)
}
