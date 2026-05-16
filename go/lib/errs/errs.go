package errs

import (
	"database/sql"
	"errors"
	"fmt"
)

type Code int

const (
	OK                 Code = 0
	Canceled           Code = 1
	Unknown            Code = 2
	InvalidArgument    Code = 3
	DeadlineExceeded   Code = 4
	NotFound           Code = 5
	AlreadyExists      Code = 6
	PermissionDenied   Code = 7
	ResourceExhausted  Code = 8
	FailedPrecondition Code = 9
	Aborted            Code = 10
	OutOfRange         Code = 11
	Unimplemented      Code = 12
	Internal           Code = 13
	Unavailable        Code = 14
	DataLoss           Code = 15
	Unauthenticated    Code = 16
)

func (c Code) String() string {
	switch c {
	case OK:
		return "OK"
	case Canceled:
		return "CANCELED"
	case Unknown:
		return "UNKNOWN"
	case InvalidArgument:
		return "INVALID_ARGUMENT"
	case DeadlineExceeded:
		return "DEADLINE_EXCEEDED"
	case NotFound:
		return "NOT_FOUND"
	case AlreadyExists:
		return "ALREADY_EXISTS"
	case PermissionDenied:
		return "PERMISSION_DENIED"
	case ResourceExhausted:
		return "RESOURCE_EXHAUSTED"
	case FailedPrecondition:
		return "FAILED_PRECONDITION"
	case Aborted:
		return "ABORTED"
	case OutOfRange:
		return "OUT_OF_RANGE"
	case Unimplemented:
		return "UNIMPLEMENTED"
	case Internal:
		return "INTERNAL"
	case Unavailable:
		return "UNAVAILABLE"
	case DataLoss:
		return "DATA_LOSS"
	case Unauthenticated:
		return "UNAUTHENTICATED"
	default:
		return fmt.Sprintf("CODE(%d)", c)
	}
}

func (c Code) HTTPStatus() int {
	switch c {
	case OK:
		return 200
	case Canceled:
		return 499
	case InvalidArgument:
		return 400
	case DeadlineExceeded:
		return 504
	case NotFound:
		return 404
	case AlreadyExists:
		return 409
	case PermissionDenied:
		return 403
	case ResourceExhausted:
		return 429
	case FailedPrecondition:
		return 400
	case Aborted:
		return 409
	case OutOfRange:
		return 400
	case Unimplemented:
		return 501
	case Internal:
		return 500
	case Unavailable:
		return 503
	case DataLoss:
		return 500
	case Unauthenticated:
		return 401
	default:
		return 500
	}
}

// ErrNotFound is a sentinel error indicating a requested entity was not found.
var ErrNotFound = errors.New("not found")

// IsNotFound reports whether err is or wraps ErrNotFound.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// WrapNotFound returns ErrNotFound if err is sql.ErrNoRows, otherwise wraps with msg.
func WrapNotFound(err error, msg string) error {
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%s: %w", msg, ErrNotFound)
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// Error is a structured error with a gRPC-style code, an internal error
// for logging/wrapping, and an optional user-facing message.
type Error struct {
	code Code
	msg  string
	err  error
}

func (e *Error) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s: %v", e.code, e.err)
	}
	if e.msg != "" {
		return fmt.Sprintf("%s: %s", e.code, e.msg)
	}
	return e.code.String()
}

func (e *Error) Unwrap() error {
	return e.err
}

func (e *Error) Code() Code {
	return e.code
}

// Message returns the user-facing message. If none was set, it falls back
// to the code's string representation.
func (e *Error) Message() string {
	if e.msg != "" {
		return e.msg
	}
	return e.code.String()
}

// New creates an error with the given code wrapping an internal error.
func New(code Code, err error, msg string) *Error {
	return &Error{code: code, err: err, msg: msg}
}

func NewNotFound(err error, msg string) *Error {
	return &Error{code: NotFound, err: err, msg: msg}
}

func NewPermissionDenied(err error, msg string) *Error {
	return &Error{code: PermissionDenied, err: err, msg: msg}
}

func NewInternal(err error, msg string) *Error {
	return &Error{code: Internal, err: err, msg: msg}
}

func NewUnauthenticated(err error, msg string) *Error {
	return &Error{code: Unauthenticated, err: err, msg: msg}
}
