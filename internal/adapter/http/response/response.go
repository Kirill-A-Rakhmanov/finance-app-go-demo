package response

import (
	"encoding/json"
	"errors"
	apperror "finance-app-go-demo/internal/application/error"
	"math"
	"net/http"
	"strconv"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type ValidationErrorResponse struct {
	Error  string              `json:"error"`
	Fields map[string][]string `json:"fields"`
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	data any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}

func OK(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, data)
}

func Created(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusCreated, data)
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func BadRequest(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
}

func Unauthorized(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
}

func PaymentRequired(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusPaymentRequired, ErrorResponse{Error: err.Error()})
}

func Forbidden(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusForbidden, ErrorResponse{Error: err.Error()})
}

func NotFound(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
}

func Conflict(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusConflict, ErrorResponse{Error: err.Error()})
}

func UnprocessableEntity(w http.ResponseWriter, err *apperror.ValidationError) {
	writeJSON(w, http.StatusUnprocessableEntity, ValidationErrorResponse{
		Error:  err.Error(),
		Fields: err.Fields,
	})
}

func TooManyRequests(w http.ResponseWriter, err error) {
	var appErr *apperror.Error
	if errors.As(err, &appErr) && appErr.RetryAfter > 0 {
		seconds := int(math.Ceil(appErr.RetryAfter.Seconds()))
		w.Header().Set("Retry-After", strconv.Itoa(seconds))
	}

	writeJSON(w, http.StatusTooManyRequests, ErrorResponse{Error: err.Error()})
}

type ErrorRecorder interface {
	RecordError(err error)
}

const internalErrorMessage = "internal server error"

func Internal(w http.ResponseWriter, err error) {
	if recorder, ok := w.(ErrorRecorder); ok {
		recorder.RecordError(err)
	}

	writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: internalErrorMessage})
}

var kindHandlers = map[apperror.Kind]func(http.ResponseWriter, error){
	apperror.KindInvalid:         BadRequest,
	apperror.KindNotFound:        NotFound,
	apperror.KindConflict:        Conflict,
	apperror.KindForbidden:       Forbidden,
	apperror.KindUnauthorized:    Unauthorized,
	apperror.KindTooManyRequests: TooManyRequests,
}

func FromError(w http.ResponseWriter, err error) {
	var validationErr *apperror.ValidationError
	if errors.As(err, &validationErr) {
		UnprocessableEntity(w, validationErr)
		return
	}

	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		if handler, ok := kindHandlers[appErr.Kind]; ok {
			handler(w, appErr)
			return
		}
	}

	Internal(w, err)
}
