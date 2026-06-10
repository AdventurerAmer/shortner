package errs

import (
	"fmt"
)

type Code int

const (
	CodeInternal Code = iota
	CodeResourceNotFound
)

func (c Code) String() string {
	switch c {
	case CodeInternal:
		return "INTERNAL"
	case CodeResourceNotFound:
		return "RESOURCE_NOT_FOUND"
	}
	return ""
}

type Error struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func New(code Code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

func Wrap(err error, code Code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %+v", e.Message, e.Err)
	}
	return e.Message
}

func (e *Error) Is(target error) bool {
	_, ok := target.(*Error)
	return ok
}

func (e *Error) Unwrap() error {
	return e.Err
}
