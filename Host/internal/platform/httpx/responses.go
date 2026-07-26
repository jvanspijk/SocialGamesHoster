package httpx

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

const TraceIDKey = "app.trace_id"

type ErrorResponse struct {
	Code        string             `json:"code"`
	Message     string             `json:"message"`
	FieldErrors result.FieldErrors `json:"fieldErrors,omitempty"`
	TraceID     string             `json:"traceId"`
}

func WriteError(event *core.RequestEvent, appError result.AppError) error {
	traceID, _ := event.Get(TraceIDKey).(string)
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
