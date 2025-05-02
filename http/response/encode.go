package response

import (
	"encoding/json"
	"net/http"
)

func Unauthorized(w http.ResponseWriter, err error) error {
	return Encode(w, New(http.StatusUnauthorized, err))
}

func InternalError(w http.ResponseWriter, err error) error {
	return Encode(w, New(http.StatusInternalServerError, err))
}

func ServiceUnavailable(w http.ResponseWriter, err error) error {
	return Encode(w, New(http.StatusServiceUnavailable, err))
}

func BadRequest(w http.ResponseWriter, err error) error {
	return Encode(w, New(http.StatusBadRequest, err))
}

func NotAcceptable(w http.ResponseWriter, err error) error {
	return Encode(w, New(http.StatusNotAcceptable, err))
}

func MethodNotAllowed(w http.ResponseWriter, err error) error {
	return Encode(w, New(http.StatusMethodNotAllowed, err))
}

func NotFound(w http.ResponseWriter, err error) error {
	return Encode(w, New(http.StatusNotFound, err))
}

func Forbidden(w http.ResponseWriter, err error) error {
	return Encode(w, New(http.StatusForbidden, err))
}

func Encode(w http.ResponseWriter, status *Status) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status.Code)
	return json.NewEncoder(w).Encode(status)
}
