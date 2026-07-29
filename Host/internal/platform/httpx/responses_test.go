package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
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

	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnprocessableEntity)
	}
	if _, present := response["traceId"]; present {
		t.Fatalf("traceId must be omitted from expected-error responses: %#v", response)
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

func TestWriteErrorFromTranslatesApplicationAndUnexpectedErrors(t *testing.T) {
	valueError := result.Invalid("profile.invalid_name", "Enter a valid profile name.", result.FieldErrors{"name": {"Enter a name."}})
	pointerError := result.Forbidden("profile.forbidden", "This profile is unavailable.")

	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantCode    string
		wantMessage string
		wantFields  result.FieldErrors
		wantTraceID string
	}{
		{
			name:        "direct value application error",
			err:         valueError,
			wantStatus:  http.StatusUnprocessableEntity,
			wantCode:    valueError.Code,
			wantMessage: valueError.Message,
			wantFields:  valueError.FieldErrors,
		},
		{
			name:        "wrapped value application error",
			err:         fmt.Errorf("request failed: %w", valueError),
			wantStatus:  http.StatusUnprocessableEntity,
			wantCode:    valueError.Code,
			wantMessage: valueError.Message,
			wantFields:  valueError.FieldErrors,
		},
		{
			name:        "direct pointer application error",
			err:         &pointerError,
			wantStatus:  http.StatusForbidden,
			wantCode:    pointerError.Code,
			wantMessage: pointerError.Message,
		},
		{
			name:        "wrapped pointer application error",
			err:         fmt.Errorf("request failed: %w", &pointerError),
			wantStatus:  http.StatusForbidden,
			wantCode:    pointerError.Code,
			wantMessage: pointerError.Message,
		},
		{
			name:        "unexpected error",
			err:         errors.New("database unavailable"),
			wantStatus:  http.StatusInternalServerError,
			wantCode:    "internal.unexpected",
			wantMessage: "Something went wrong. Please try again.",
			wantTraceID: "trace-123",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			event := &core.RequestEvent{}
			event.Request = httptest.NewRequest(http.MethodPost, "/", nil)
			event.Response = recorder
			event.Set(TraceIDKey, "trace-123")

			if err := WriteErrorFrom(event, test.err); err != nil {
				t.Fatal(err)
			}

			var response ErrorResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, test.wantStatus)
			}
			if response.Code != test.wantCode {
				t.Errorf("code = %q, want %q", response.Code, test.wantCode)
			}
			if response.Message != test.wantMessage {
				t.Errorf("message = %q, want %q", response.Message, test.wantMessage)
			}
			if !reflect.DeepEqual(response.FieldErrors, test.wantFields) {
				t.Errorf("fieldErrors = %#v, want %#v", response.FieldErrors, test.wantFields)
			}
			if response.TraceID != test.wantTraceID {
				t.Errorf("traceId = %q, want %q", response.TraceID, test.wantTraceID)
			}
		})
	}
}
