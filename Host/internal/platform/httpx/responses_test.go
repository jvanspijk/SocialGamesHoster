package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

func TestWriteErrorOmitsTraceForExpectedErrors(t *testing.T) {
	recorder := httptest.NewRecorder()
	event := &core.RequestEvent{}
	event.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	event.Response = recorder
	event.Set(TraceIDKey, "trace-should-not-be-visible")

	if err := WriteError(event, result.Invalid("profile.invalid_name", "Enter a valid profile name.", nil)); err != nil {
		t.Fatal(err)
	}

	var response ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnprocessableEntity)
	}
	if response.TraceID != "" {
		t.Fatalf("traceId = %q, want omitted", response.TraceID)
	}
}

func TestWriteErrorIncludesTraceForInternalErrors(t *testing.T) {
	recorder := httptest.NewRecorder()
	event := &core.RequestEvent{}
	event.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	event.Response = recorder
	event.Set(TraceIDKey, "trace-123")

	if err := WriteError(event, result.Internal(errors.New("database unavailable"))); err != nil {
		t.Fatal(err)
	}

	var response ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if response.TraceID != "trace-123" {
		t.Fatalf("traceId = %q, want %q", response.TraceID, "trace-123")
	}
}
