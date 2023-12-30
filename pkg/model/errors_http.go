package model

import "net/http"

var (
	ErrUnhealthy      = NewError(http.StatusInternalServerError, "something went wrong")
	ErrRefreshExpired = NewError(http.StatusUnauthorized, "refresh")
	ErrUnauthorized   = NewError(http.StatusUnauthorized, "user unauthorized")
	ErrInvalidBody    = NewError(http.StatusBadRequest, "request invalid body")
	ErrInvalidRole    = NewError(http.StatusBadRequest, "user invalid role")
	ErrUsernameExist  = NewError(http.StatusBadRequest, "username exist")

	ErrAlreadyUnsubscribed = NewError(http.StatusBadRequest, "you have already unsubscribed")
	ErrAlreadySubscribed   = NewError(http.StatusBadRequest, "you have already subscribed")
	ErrIncorrectSymbol     = NewError(http.StatusBadRequest, "incorrect symbol")
)

const (
	NotFound = "not found"
)

type Error interface {
	error
	Status() int
}

// StatusError represents HTTP error.
type StatusError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error allows StatusError to satisfy the error interface.
func (se StatusError) Error() string {
	return se.Message
}

// Status returns our HTTP status code.
func (se StatusError) Status() int {
	return se.Code
}

//nolint:ireturn
func NewError(code int, message string) Error {
	return StatusError{
		Code:    code,
		Message: message,
	}
}
