package httpx

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

func TestErrorResponseOmitsTraceForExpectedErrors(t *testing.T) {
	payload, err := json.Marshal(ErrorResponse{
		Code:    "profile.invalid_name",
		Message: "Enter a valid profile name.",
		TraceID: "trace-should-not-be-visible",
	})
	if err != nil {
		t.Fatal(err)
	}

	// WriteError only supplies TraceID for an operational cause. The response
	// shape must also keep the field absent when no trace is available.
	withoutTrace, err := json.Marshal(ErrorResponse{Code: "profile.invalid_name", Message: "Enter a valid profile name."})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(withoutTrace), "traceId") {
		t.Fatalf("expected no traceId in validation response: %s", withoutTrace)
	}
	if !strings.Contains(string(payload), "trace-should-not-be-visible") {
		t.Fatal("response contract should still allow an operational trace when explicitly supplied")
	}
}

func TestInternalErrorRetainsOperationalCause(t *testing.T) {
	err := result.Internal(errors.New("database unavailable"))
	if err.Cause == nil {
		t.Fatal("internal errors must retain their cause for diagnostics")
	}
}
