package diagnostics

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

func TestProjectedLogsRedactSensitiveFields(t *testing.T) {
	logs := []*core.Log{{
		Message: "request failed",
		Data: types.JSONMap[any]{
			"type": "request", "method": "POST", "url": "/safe", "status": 500,
			"authorization": "Bearer secret-token", "body": "private chat body",
			"password": "correct horse battery staple", "error": "token secret-token leaked here",
		},
	}}
	data, err := json.Marshal(projectLogs(logs))
	if err != nil {
		t.Fatal(err)
	}
	value := string(data)
	for _, forbidden := range []string{"secret-token", "private chat body", "correct horse battery staple", "authorization", "password"} {
		if strings.Contains(value, forbidden) {
			t.Fatalf("support log projection leaked %q", forbidden)
		}
	}
	if !strings.Contains(value, `"method":"POST"`) {
		t.Fatal("support log projection omitted safe request metadata")
	}
	if !strings.Contains(value, `"errorFingerprint"`) {
		t.Fatal("support log projection omitted the safe error fingerprint")
	}
}
