package resp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

type Error struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e *Error) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Field)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func MissingParameter(field string) *Error {
	return &Error{
		Status:  http.StatusBadRequest,
		Code:    "missing_parameter",
		Message: "parameter is required",
		Field:   field,
	}
}

func InvalidParameter(field, message string) *Error {
	return &Error{
		Status:  http.StatusBadRequest,
		Code:    "invalid_parameter",
		Message: message,
		Field:   field,
	}
}

type envelope struct {
	Error *Error `json:"error"`
}

func SendStruct(ctx context.Context, w http.ResponseWriter, status int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		slog.ErrorContext(ctx, "encoding response failed", "error", err)
		http.Error(w, `{"error":{"code":"internal_error","message":"internal server error"}}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if _, err := w.Write(body); err != nil {
		slog.WarnContext(ctx, "writing response failed", "error", err)
	}
}

func SendError(ctx context.Context, w http.ResponseWriter, err error) {
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		slog.ErrorContext(ctx, "unhandled error", "error", err)
		apiErr = &Error{
			Status:  http.StatusInternalServerError,
			Code:    "internal_error",
			Message: "internal server error",
		}
	}
	SendStruct(ctx, w, apiErr.Status, envelope{Error: apiErr})
}

func SendStatus(ctx context.Context, w http.ResponseWriter, status int, code, message string) {
	SendError(ctx, w, &Error{Status: status, Code: code, Message: message})
}
