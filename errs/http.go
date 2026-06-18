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
	case CodeResourceAlreadyExists:
		return http.StatusConflict
	case CodeTimeout:
		return http.StatusRequestTimeout
	}
	return http.StatusInternalServerError
}
