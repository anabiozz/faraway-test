package tcp

import (
	"errors"
	"fmt"
)

var (
	// Protocol errors
	ErrInvalidProtocol    = errors.New("invalid protocol format")
	ErrInvalidMessageSize = errors.New("invalid message size")

	// Connection errors
	ErrConnectionClosed = errors.New("connection closed")
	ErrReadTimeout      = errors.New("read operation timeout")
	ErrWriteTimeout     = errors.New("write operation timeout")

	// Challenge errors
	ErrInvalidChallenge     = errors.New("invalid challenge format")
	ErrSolutionNotFound     = errors.New("solution not found")
	ErrInvalidChallengeType = errors.New("invalid challenge type")

	// System errors
	ErrMaxRetriesExceeded = errors.New("maximum retry attempts exceeded")
)

type ClientError struct {
	Op   string
	Err  error
	Info string
}

func (e *ClientError) Error() string {
	if e.Info != "" {
		return fmt.Sprintf("%s: %v (%s)", e.Op, e.Err, e.Info)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *ClientError) Unwrap() error {
	return e.Err
}

func NewClientError(op string, err error, info string) error {
	return &ClientError{
		Op:   op,
		Err:  err,
		Info: info,
	}
}

// Helper functions
func IsRetryableError(err error) bool {
	var clientErr *ClientError
	if errors.As(err, &clientErr) {
		switch {
		case errors.Is(err, ErrConnectionClosed):
			return true
		case errors.Is(err, ErrReadTimeout):
			return true
		case errors.Is(err, ErrWriteTimeout):
			return true
		default:
			return false
		}
	}
	return false
}
