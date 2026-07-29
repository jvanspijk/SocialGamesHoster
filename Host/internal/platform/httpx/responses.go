package httpx

import (
	"errors"
	"net/http"

	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

const TraceIDKey = "app.trace_id"

type ErrorResponse struct {
	Code        string             `json:"code"`
	Message     string             `json:"message"`
	FieldErrors result.FieldErrors `json:"fieldErrors,omitempty"`
	TraceID     string             `json:"traceId,omitempty"`
}

func WriteError(event *core.RequestEvent, appError result.AppError) error {
	// A trace is for unexpected failures that need investigation. Expected
	// request problems are actionable in the form and should not look like
	// application incidents to the user or create diagnostic noise.
	traceID := ""
	if appError.Cause != nil {
		traceID, _ = event.Get(TraceIDKey).(string)
	}
	status := appError.Status
	if status == 0 {
		status = http.StatusInternalServerError
	}
	return event.JSON(status, ErrorResponse{
		Code:        appError.Code,
		Message:     appError.Message,
		FieldErrors: appError.FieldErrors,
		TraceID:     traceID,
	})
}

// WriteErrorFrom translates an arbitrary application error at the HTTP
// boundary. Expected application errors retain their public contract even when
// wrapped; unexpected errors are exposed only as an internal error.
func WriteErrorFrom(event *core.RequestEvent, err error) error {
	var appError result.AppError
	if errors.As(err, &appError) {
		return WriteError(event, appError)
	}

	var appErrorPointer *result.AppError
	if errors.As(err, &appErrorPointer) {
		return WriteError(event, *appErrorPointer)
	}

	return WriteError(event, result.Internal(err))
}
