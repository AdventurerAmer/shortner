package errs

import "errors"

func NewInternal(err error) *Error {
	return &Error{
		Code:    CodeInternal,
		Message: "internal server error",
		Err:     err,
	}
}

func NewValidation(fields map[string]string) *Error {
	err := New(CodeValidation, "one or more invalid fields")
	err.Fields = fields
	return err
}

func NewNotFound(err error, message string) *Error {
	return Wrap(err, CodeResourceNotFound, message)
}

func NewAlreadyExists(err error, message string) *Error {
	return Wrap(err, CodeResourceAlreadyExists, message)
}

func NewTimeout(err error) *Error {
	return Wrap(err, CodeTimeout, "timeout")
}

func IsNotFound(err error) bool {
	var e *Error
	if errors.As(err, &e) && e.Code == CodeResourceNotFound {
		return true
	}
	return false
}

func IsValidation(err error) bool {
	var e *Error
	if errors.As(err, &e) && e.Code == CodeValidation {
		return true
	}
	return false
}

func IsAlreadyExists(err error) bool {
	var e *Error
	if errors.As(err, &e) && e.Code == CodeResourceAlreadyExists {
		return true
	}
	return false
}

func IsTimeout(err error) bool {
	var e *Error
	if errors.As(err, &e) && e.Code == CodeTimeout {
		return true
	}
	return false
}
