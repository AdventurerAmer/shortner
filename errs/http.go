package errs

import (
	"net/http"
)

func HTTPStatus(code Code) int {
	switch code {
	case CodeInternal, CodeUnknown:
		return http.StatusInternalServerError
	case CodeValidation:
		return http.StatusBadRequest
	case CodeResourceNotFound:
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}
