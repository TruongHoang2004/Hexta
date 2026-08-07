package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var (
	serviceName string
	traceIdKey  interface{} // Key to retrieve trace ID from context
)

// Init configures the errors package with the service name and trace ID key.
// This should be called once at the application startup.
func Init(svcName string, tIdKey interface{}) {
	serviceName = svcName
	traceIdKey = tIdKey
}

func GetTraceId(ctx context.Context) string {
	if ctx == nil || traceIdKey == nil {
		return ""
	}
	if v := ctx.Value(traceIdKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

const (
	ServiceName string = "ServiceName"
)

type CodeResponse string

type ErrorResponse struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	TraceID    string `json:"trace_id,omitempty"`
	Detail     string `json:"detail,omitempty"`
	Source     string `json:"source"`
	HTTPStatus int    `json:"http_status"`
}

const (
	//internal
	ErrorCodeBadRequest   CodeResponse = "BAD_REQUEST"
	ErrorCodeUnauthorized CodeResponse = "UNAUTHORIZED"
	ErrorCodeForbidden    CodeResponse = "FORBIDDEN"
	ErrorCodeNotFound     CodeResponse = "NOT_FOUND"
	ErrorCodeConflict     CodeResponse = "CONFLICT"
	ErrorCodeSystemError  CodeResponse = "INTERNAL_SERVER_ERROR"
)

const (
	DefaultServerErrorMessage  = "Something has gone wrong, please contact admin"
	DefaultBadRequestMessage   = "Invalid request"
	DefaultUnauthorizedMessage = "Token invalid"
	DefaultForbiddenMessage    = "Forbidden"
	DefauultConflict           = "Conflict"
)

type Source string

type Error struct {
	Code       CodeResponse `json:"code"`
	Message    string       `json:"message"`
	TraceID    string       `json:"trace_id,omitempty"`
	Detail     string       `json:"detail"`
	Source     Source       `json:"source"`
	HTTPStatus int          `json:"http_status"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("code:[%s], message:[%s], detail:[%s], source:[%s]", e.Code, e.Message, e.Detail, e.Source)
}

func (e *Error) GetHttpStatus() int {
	return e.HTTPStatus
}

func (e *Error) GetCode() CodeResponse {
	return e.Code
}

func (e *Error) GetMessage() string {
	return e.Message
}

func (e *Error) SetTraceId(traceId string) *Error {
	e.TraceID = fmt.Sprintf("%s:%d", traceId, time.Now().Unix())
	return e
}

func (e *Error) SetHTTPStatus(status int) *Error {
	e.HTTPStatus = status
	return e
}

func (e *Error) SetMessage(msg string) *Error {
	e.Message = msg
	return e
}

func (e *Error) SetDetail(detail string) *Error {
	e.Detail = detail
	return e
}

func (e *Error) GetDetail() string {
	return e.Detail
}

func (e *Error) SetSource(source Source) *Error {
	if source == "" {
		e.Source = Source(serviceName)
	}
	return e
}

func (e *Error) ToJSon() string {
	data, err := json.Marshal(e)
	if err != nil {
		//Todo fix this
		return "marshal error failed"
	}
	return string(data)
}

// Helper to create base error with current context
func newError(ctx context.Context, code CodeResponse, msg string, status int) *Error {
	return &Error{
		Code:       code,
		Message:    msg,
		TraceID:    GetTraceId(ctx),
		Source:     Source(serviceName),
		HTTPStatus: status,
	}
}

var (
	// Status 4xx ********

	ErrUnauthorized = func(ctx context.Context) *Error {
		return newError(ctx, ErrorCodeUnauthorized, DefaultUnauthorizedMessage, http.StatusUnauthorized)
	}

	ErrNotFound = func(ctx context.Context, object, status string) *Error {
		return newError(ctx, ErrorCodeNotFound, getMsg(object, status), http.StatusNotFound)
	}

	ErrBadRequest = func(ctx context.Context) *Error {
		return newError(ctx, ErrorCodeBadRequest, DefaultBadRequestMessage, http.StatusBadRequest)
	}

	ErrConflict = func(ctx context.Context, object, status string) *Error {
		return newError(ctx, ErrorCodeConflict, getMsg(object, status), http.StatusConflict)
	}

	// Status 5xx *******

	ErrSystemError = func(ctx context.Context, detail string) *Error {
		err := newError(ctx, ErrorCodeSystemError, DefaultServerErrorMessage, http.StatusInternalServerError)
		err.Detail = detail
		return err
	}

	ErrForbidden = func(ctx context.Context) *Error {
		return newError(ctx, ErrorCodeForbidden, DefaultForbiddenMessage, http.StatusForbidden)
	}
)

func getMsg(object, status string) string {
	return fmt.Sprintf("%s %s", object, status)
}

func ConvertErrorToResponse(err *Error) *ErrorResponse {
	return &ErrorResponse{
		Code:       string(err.Code),
		Message:    err.Message,
		TraceID:    err.TraceID,
		Detail:     err.Detail,
		Source:     string(err.Source),
		HTTPStatus: err.HTTPStatus,
	}
}

func IsClientError(err *Error) bool {
	if err == nil {
		return false
	}
	if http.StatusBadRequest <= err.GetHttpStatus() && err.GetHttpStatus() < http.StatusInternalServerError {
		return true
	}
	return false
}

func IsInternalError(err *Error) bool {
	if err == nil {
		return false
	}
	if err.GetHttpStatus() >= http.StatusInternalServerError {
		return true
	}
	return false
}
