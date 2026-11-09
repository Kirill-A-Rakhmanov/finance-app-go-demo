package error

import "time"

type Kind string

const (
	KindInvalid         Kind = "invalid"
	KindNotFound        Kind = "not_found"
	KindConflict        Kind = "conflict"
	KindForbidden       Kind = "forbidden"
	KindUnauthorized    Kind = "unauthorized"
	KindTooManyRequests Kind = "too_many_requests"
)

type Error struct {
	Kind       Kind
	Message    string
	RetryAfter time.Duration
}

func (e *Error) Error() string {
	return e.Message
}

func Invalid(message string) *Error {
	return &Error{Kind: KindInvalid, Message: message}
}

func NotFound(message string) *Error {
	return &Error{Kind: KindNotFound, Message: message}
}

func Conflict(message string) *Error {
	return &Error{Kind: KindConflict, Message: message}
}

func Forbidden(message string) *Error {
	return &Error{Kind: KindForbidden, Message: message}
}

func Unauthorized(message string) *Error {
	return &Error{Kind: KindUnauthorized, Message: message}
}

func TooManyRequests(message string, retryAfter time.Duration) *Error {
	return &Error{Kind: KindTooManyRequests, Message: message, RetryAfter: retryAfter}
}
