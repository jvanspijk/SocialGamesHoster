package diagnostics

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	actorauth "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/actors"
)

func Register(event *core.ServeEvent, enabled bool, startedAt time.Time, applicationVersion string) {
	if !enabled {
		return
	}
	group := event.Router.Group("/api/app/v1/diagnostics")
	group.BindFunc(actorauth.RequireOwner)
	group.GET("/health", func(request *core.RequestEvent) error {
		return request.JSON(http.StatusOK, map[string]any{
			"status":     "ok",
			"serverTime": time.Now().UTC(),
			"uptimeMs":   time.Since(startedAt).Milliseconds(),
			"goVersion":  runtime.Version(),
			"goroutines": runtime.NumGoroutine(),
			"sseClients": len(request.App.SubscriptionsBroker().Clients()),
		})
	})
	group.GET("/errors", func(request *core.RequestEvent) error {
		logs := []*core.Log{}
		err := request.App.LogQuery().
			AndWhere(dbx.NewExp("level >= {:level}", dbx.Params{"level": int(slog.LevelError)})).
			OrderBy("created DESC").
			Limit(100).
			All(&logs)
		if err != nil {
			return request.JSON(http.StatusInternalServerError, map[string]string{"code": "diagnostics.failed", "message": "Diagnostics could not be loaded."})
		}
		return request.JSON(http.StatusOK, projectLogs(logs))
	})
	group.GET("/requests", func(request *core.RequestEvent) error {
		logs := []*core.Log{}
		err := request.App.LogQuery().
			AndWhere(dbx.NewExp("json_extract(data, '$.type') = 'request'")).
			OrderBy("created DESC").
			Limit(100).
			All(&logs)
		if err != nil {
			return request.JSON(http.StatusInternalServerError, map[string]string{"code": "diagnostics.failed", "message": "Diagnostics could not be loaded."})
		}
		return request.JSON(http.StatusOK, projectLogs(logs))
	})
	group.GET("/resources", func(request *core.RequestEvent) error {
		var memory runtime.MemStats
		runtime.ReadMemStats(&memory)
		return request.JSON(http.StatusOK, map[string]any{
			"serverTime":     time.Now().UTC(),
			"goroutines":     runtime.NumGoroutine(),
			"heapAllocated":  memory.HeapAlloc,
			"heapInUse":      memory.HeapInuse,
			"systemMemory":   memory.Sys,
			"garbageCycles":  memory.NumGC,
			"sseClients":     len(request.App.SubscriptionsBroker().Clients()),
			"databaseHealth": databaseHealth(request.App),
		})
	})
	group.POST("/support-bundle", func(request *core.RequestEvent) error {
		return supportBundle(request, startedAt, applicationVersion)
	})
}

func projectLogs(logs []*core.Log) []map[string]any {
	result := make([]map[string]any, len(logs))
	for index, log := range logs {
		safe := map[string]any{}
		for _, key := range []string{"type", "method", "url", "status", "requestId", "traceId", "elapsed"} {
			if value, exists := log.Data[key]; exists {
				safe[key] = value
			}
		}
		if raw, exists := log.Data["error"]; exists {
			digest := sha256.Sum256([]byte(fmt.Sprint(raw)))
			safe["errorFingerprint"] = hex.EncodeToString(digest[:6])
		}
		result[index] = map[string]any{
			"id": log.Id, "created": log.Created.Time().UTC(), "level": log.Level,
			"message": log.Message, "data": safe,
		}
	}
	return result
}

func databaseHealth(app core.App) string {
	var value int
	if err := app.DB().NewQuery("SELECT 1").Row(&value); err != nil || value != 1 {
		return "unavailable"
	}
	return "ok"
}
