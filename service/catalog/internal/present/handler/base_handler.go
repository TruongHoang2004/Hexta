package handler

import (
	"fmt"

	"gitlab.com/ecommercehub1/lib/pkg/errors"
)

type baseHandler struct {
}

func NewBaseHandler() *baseHandler {
	return &baseHandler{}
}

func (b *baseHandler) IErrorToGRPCError(err *errors.Error) error {
	return fmt.Errorf("%s", err.ToJSon())
}
