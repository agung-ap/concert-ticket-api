package errors

import (
	"errors"
	"fmt"
)

// Common errors
var (
	ErrNotFound                = errors.New("resource not found")
	ErrUnauthorized            = errors.New("unauthorized")
	ErrForbidden               = errors.New("forbidden")
	ErrInternalServer          = errors.New("internal server error")
	ErrOptimisticLockFailed    = errors.New("optimistic lock failed")
	ErrUpdateFailed            = errors.New("update failed")
	ErrInsufficientTickets     = errors.New("insufficient tickets")
	ErrBookingClosed           = errors.New("booking is closed")
	ErrBookingAlreadyCancelled = errors.New("booking is already cancelled")
)

// ErrorWithMessage represents an error with a message
type ErrorWithMessage struct {
	err     error
	message string
}

// NewErrorWithMessage creates a new error with a message
func NewErrorWithMessage(err error, message string) *ErrorWithMessage {
	return &ErrorWithMessage{
		err:     err,
		message: message,
	}
}

// Error returns the error message
func (e *ErrorWithMessage) Error() string {
	return e.message
}

// Unwrap returns the wrapped error
func (e *ErrorWithMessage) Unwrap() error {
	return e.err
}

// Message returns the message
func (e *ErrorWithMessage) Message() string {
	return e.message
}

// Is checks if the target error is the same as this error
func (e *ErrorWithMessage) Is(target error) bool {
	if target == nil {
		return false
	}

	// Check if the target is the same as the wrapped error
	if errors.Is(e.err, target) {
		return true
	}

	// Check if the target is the same type
	_, ok := target.(*ErrorWithMessage)
	return ok
}

// ErrInvalidInput creates a new invalid input error
func ErrInvalidInput(message string) error {
	return NewErrorWithMessage(errors.New("invalid input"), message)
}

// IsInvalidInput checks if the error is an invalid input error
func IsInvalidInput(err error) bool {
	var invalidErr *ErrorWithMessage
	if errors.As(err, &invalidErr) {
		return errors.Is(invalidErr.Unwrap(), errors.New("invalid input"))
	}
	return false
}

// DBError represents a database error
type DBError struct {
	err     error
	message string
	code    string
}

// NewDBError creates a new database error
func NewDBError(err error, message, code string) *DBError {
	return &DBError{
		err:     err,
		message: message,
		code:    code,
	}
}

// Error returns the error message
func (e *DBError) Error() string {
	return fmt.Sprintf("database error: %s", e.message)
}

// Unwrap returns the wrapped error
func (e *DBError) Unwrap() error {
	return e.err
}

// Code returns the error code
func (e *DBError) Code() string {
	return e.code
}

// Message returns the message
func (e *DBError) Message() string {
	return e.message
}
