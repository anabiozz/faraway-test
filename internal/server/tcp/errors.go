package tcp

import (
	"errors"
	"fmt"
)

// Custom error types
var (
	// Protocol errors
	ErrInvalidProtocol = errors.New("invalid protocol format")
	ErrInvalidSolution = errors.New("invalid proof of work solution")

	// Connection errors
	ErrConnectionClosed = errors.New("connection closed")
	ErrReadTimeout      = errors.New("read operation timeout")
	ErrWriteTimeout     = errors.New("write operation timeout")

	// Challenge errors
	ErrChallengeFailed   = errors.New("failed to generate challenge")
	ErrChallengeDelivery = errors.New("failed to deliver challenge")

	// Solution errors
	ErrSolutionFormat     = errors.New("invalid solution format")
	ErrSolutionValidation = errors.New("solution validation failed")

	// System errors
	ErrServerShutdown = errors.New("server is shutting down")
	ErrInternal       = errors.New("internal server error")
)

// Error types with additional context
type ServerError struct {
	Op   string // Operation that failed
	Err  error  // Original error
	Info string // Additional context
}

func (e *ServerError) Error() string {
	if e.Info != "" {
		return fmt.Sprintf("%s: %v (%s)", e.Op, e.Err, e.Info)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *ServerError) Unwrap() error {
	return e.Err
}

// Error wrapping functions
func NewConnectionError(op string, err error, info string) error {
	return &ServerError{
		Op:   op,
		Err:  err,
		Info: info,
	}
}

// Helper functions for common error cases
func IsTimeoutError(err error) bool {
	return errors.Is(err, ErrReadTimeout) || errors.Is(err, ErrWriteTimeout)
}

func IsProtocolError(err error) bool {
	return errors.Is(err, ErrInvalidProtocol) || errors.Is(err, ErrInvalidSolution)
}

// Error response types
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Common error responses
var (
	ErrRespInvalidFormat = ErrorResponse{
		Code:    "INVALID_FORMAT",
		Message: "Invalid message format",
	}
	ErrRespTimeout = ErrorResponse{
		Code:    "TIMEOUT",
		Message: "Operation timed out",
	}
	ErrRespInvalidSolution = ErrorResponse{
		Code:    "INVALID_SOLUTION",
		Message: "Invalid proof of work solution",
	}
)

// Helper function to convert errors to responses
func ToErrorResponse(err error) ErrorResponse {
	switch {
	case errors.Is(err, ErrInvalidProtocol):
		return ErrRespInvalidFormat
	case IsTimeoutError(err):
		return ErrRespTimeout
	case errors.Is(err, ErrInvalidSolution):
		return ErrRespInvalidSolution
	default:
		return ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "An internal error occurred",
		}
	}
}
