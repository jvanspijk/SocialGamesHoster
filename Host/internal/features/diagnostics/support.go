package diagnostics

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func supportBundle(event *core.RequestEvent, startedAt time.Time, applicationVersion string) error {
	buffer := bytes.NewBuffer(nil)
	writer := zip.NewWriter(buffer)
	if err := addJSON(writer, "versions.json", map[string]any{
		"application": applicationVersion,
		"go":          runtime.Version(),
		"pocketBase":  "0.39.9",
		"schema":      1,
		"generatedAt": time.Now().UTC(),
	}); err != nil {
		return supportError(event, err)
	}
	if err := addJSON(writer, "resources.json", resourceSummary(event.App, startedAt)); err != nil {
		return supportError(event, err)
	}
	if err := addJSON(writer, "database-counts.json", databaseCounts(event.App)); err != nil {
		return supportError(event, err)
	}
	if err := addJSON(writer, "network.json", networkSummary()); err != nil {
		return supportError(event, err)
	}
	if err := addJSON(writer, "settings.json", sanitizedSettings(event.App)); err != nil {
		return supportError(event, err)
	}
	logs := []*core.Log{}
	_ = event.App.LogQuery().
		AndWhere(dbx.NewExp("level >= {:level}", dbx.Params{"level": 4})).
		OrderBy("created DESC").
		Limit(250).
		All(&logs)
	if err := addJSON(writer, "recent-logs.json", projectLogs(logs)); err != nil {
		return supportError(event, err)
	}
	if err := writer.Close(); err != nil {
		return supportError(event, err)
	}
	event.Response.Header().Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="social-games-hoster-support-%s.zip"`, time.Now().UTC().Format("20060102-150405")))
	return event.Blob(http.StatusOK, "application/zip", buffer.Bytes())
}

func addJSON(writer *zip.Writer, name string, value any) error {
	entry, err := writer.Create(name)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(entry)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func resourceSummary(app core.App, startedAt time.Time) map[string]any {
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	files := map[string]int64{}
	for _, name := range []string{"data.db", "auxiliary.db"} {
		info, err := os.Stat(filepath.Join(app.DataDir(), name))
		if err == nil {
			files[name] = info.Size()
		}
	}
	return map[string]any{
		"serverTime": time.Now().UTC(), "uptimeMs": time.Since(startedAt).Milliseconds(),
		"goroutines": runtime.NumGoroutine(), "heapAllocated": memory.HeapAlloc,
		"heapInUse": memory.HeapInuse, "systemMemory": memory.Sys, "garbageCycles": memory.NumGC,
		"sseClients": len(app.SubscriptionsBroker().Clients()), "databaseFiles": files,
	}
}

func databaseCounts(app core.App) map[string]any {
	names := []string{
		"game_masters", "player_profiles", "profile_requests", "rulesets", "ruleset_versions",
		"ruleset_assets", "games", "participants", "chat_rooms", "chat_memberships",
		"chat_messages", "achievement_awards", "game_audit",
	}
	counts := map[string]any{}
	for _, name := range names {
		count, err := app.CountRecords(name)
		if err == nil {
			counts[name] = count
		}
	}
	return counts
}

func networkSummary() []map[string]any {
	interfaces, err := net.Interfaces()
	if err != nil {
		return []map[string]any{{"error": "network adapter information unavailable"}}
	}
	result := make([]map[string]any, 0, len(interfaces))
	for _, adapter := range interfaces {
		addresses, _ := adapter.Addrs()
		values := make([]string, 0, len(addresses))
		for _, address := range addresses {
			values = append(values, address.String())
		}
		result = append(result, map[string]any{
			"name": adapter.Name, "flags": adapter.Flags.String(), "addresses": values,
		})
	}
	return result
}

func sanitizedSettings(app core.App) map[string]any {
	settings, err := app.FindRecordById("host_settings", "sghhostsettings")
	if err != nil {
		return map[string]any{"configured": false}
	}
	return map[string]any{
		"configured": true, "port": settings.GetInt("port"),
		"bindAddress":            settings.GetString("bind_address"),
		"preferredAdapter":       settings.GetString("preferred_adapter"),
		"trustedLanAcknowledged": settings.GetBool("trusted_lan_acknowledged"),
		"automaticBackups":       settings.GetBool("automatic_backups"),
	}
}

func supportError(event *core.RequestEvent, err error) error {
	event.App.Logger().Error("support bundle generation failed", "error", err)
	return event.JSON(http.StatusInternalServerError, map[string]any{
		"code": "diagnostics.support_failed", "message": "The support bundle could not be created.",
		"traceId": event.Get("app.trace_id"),
	})
}
