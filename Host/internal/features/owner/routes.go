package owner

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"

	actorauth "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/actors"
	applicationaudit "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/audit"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/desktop"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/recovery"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

const settingsID = "sghhostsettings"

type settingsRequest struct {
	Port                   int    `json:"port"`
	BindAddress            string `json:"bindAddress"`
	PreferredAdapter       string `json:"preferredAdapter"`
	TrustedLANAcknowledged bool   `json:"trustedLanAcknowledged"`
	AutomaticBackups       bool   `json:"automaticBackups"`
}

type restoreRequest struct {
	Confirmation string `json:"confirmation"`
}

func Register(event *core.ServeEvent) {
	group := event.Router.Group("/api/app/v1/owner")
	group.BindFunc(actorauth.RequireOwner)
	group.GET("/settings", getSettings)
	group.PATCH("/settings", updateSettings)
	group.POST("/backups", createBackup)
	group.GET("/backups", listBackups)
	group.POST("/backups/{id}/restore", restoreBackup)
}

func RegisterJobs(app core.App) {
	app.Cron().MustAdd("sgh_daily_backup", "15 * * * *", func() {
		settings, err := EnsureSettings(app)
		if err != nil || !settings.GetBool("automatic_backups") {
			return
		}
		if err := EnsureDailyBackup(app); err != nil {
			app.Logger().Error("automatic backup failed", "error", err)
		}
	})
}

type RuntimeSettings struct {
	Port             int
	BindAddress      string
	PreferredAddress string
	AutomaticBackups bool
}

func RuntimeConfiguration(app core.App) (RuntimeSettings, error) {
	record, err := EnsureSettings(app)
	if err != nil {
		return RuntimeSettings{}, err
	}
	port := record.GetInt("port")
	if port == 0 {
		port = 8090
	}
	bindAddress := record.GetString("bind_address")
	if bindAddress == "" {
		bindAddress = "0.0.0.0"
	}
	preferred := preferredPrivateAddress(record.GetString("preferred_adapter"))
	return RuntimeSettings{
		Port:             port,
		BindAddress:      bindAddress,
		PreferredAddress: preferred,
		AutomaticBackups: record.GetBool("automatic_backups"),
	}, nil
}

func JoinURL(app core.App) string {
	settings, err := RuntimeConfiguration(app)
	if err != nil {
		return "http://127.0.0.1:8090/"
	}
	address := settings.PreferredAddress
	if address == "" {
		address = "127.0.0.1"
	}
	return fmt.Sprintf("http://%s:%d/", address, settings.Port)
}

func AllowedOrigins(app core.App) []string {
	settings, err := RuntimeConfiguration(app)
	if err != nil {
		return []string{"http://127.0.0.1:8090", "http://localhost:8090"}
	}
	origins := []string{
		fmt.Sprintf("http://127.0.0.1:%d", settings.Port),
		fmt.Sprintf("http://localhost:%d", settings.Port),
	}
	if hostname, hostnameErr := os.Hostname(); hostnameErr == nil && hostname != "" {
		origins = append(origins, fmt.Sprintf("http://%s:%d", hostname, settings.Port))
	}
	for _, address := range privateAddresses() {
		origins = append(origins, fmt.Sprintf("http://%s:%d", address["address"], settings.Port))
	}
	return origins
}

func CreateManualBackup(app core.App) (string, error) {
	name := "manual_sgh_" + time.Now().UTC().Format("20060102_150405") + ".zip"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	return name, app.CreateBackup(ctx, name)
}

func EnsurePreMigrationBackup(app core.App, applicationVersion string) error {
	if _, err := os.Stat(filepath.Join(app.DataDir(), "data.db")); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	safeVersion := strings.NewReplacer(".", "_", "-", "_", "+", "_").Replace(applicationVersion)
	if safeVersion == "" {
		safeVersion = "unknown"
	}
	name := "pre_upgrade_sgh_" + safeVersion + ".zip"
	fsys, err := app.NewBackupsFilesystem()
	if err != nil {
		return err
	}
	exists, existsErr := fsys.Exists(name)
	fsys.Close()
	if existsErr == nil && exists {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	return app.CreateBackup(ctx, name)
}

func getSettings(event *core.RequestEvent) error {
	settings, err := EnsureSettings(event.App)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	return event.JSON(http.StatusOK, projectSettings(event.App, settings))
}

func updateSettings(event *core.RequestEvent) error {
	var request settingsRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("settings.invalid", "The host settings could not be read.", nil))
	}
	if request.Port < 1 || request.Port > 65535 {
		return httpx.WriteError(event, result.Invalid("settings.invalid_port", "Choose a port between 1 and 65535.", nil))
	}
	request.BindAddress = strings.TrimSpace(request.BindAddress)
	if request.BindAddress != "" && net.ParseIP(request.BindAddress) == nil {
		return httpx.WriteError(event, result.Invalid("settings.invalid_address", "Choose a valid local IP address.", nil))
	}
	settings, err := EnsureSettings(event.App)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	if settings.GetInt("port") != request.Port {
		if err := desktop.UpdateFirewallPort(request.Port); err != nil {
			return httpx.WriteError(event, result.Conflict("settings.firewall_update_failed", "Windows did not allow the private-network firewall rule to be updated. No settings were changed."))
		}
	}
	settings.Set("port", request.Port)
	settings.Set("bind_address", request.BindAddress)
	settings.Set("preferred_adapter", strings.TrimSpace(request.PreferredAdapter))
	settings.Set("trusted_lan_acknowledged", request.TrustedLANAcknowledged)
	settings.Set("automatic_backups", request.AutomaticBackups)
	if err := event.App.RunInTransaction(func(tx core.App) error {
		if err := tx.Save(settings); err != nil {
			return err
		}
		return applicationaudit.Record(tx, event.Auth, "", "host.settings_updated", "host_settings", settings.Id,
			map[string]any{"port": request.Port, "automaticBackups": request.AutomaticBackups}, event.Get(httpx.TraceIDKey))
	}); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	return event.JSON(http.StatusOK, projectSettings(event.App, settings))
}

func createBackup(event *core.RequestEvent) error {
	name, err := CreateManualBackup(event.App)
	if err != nil {
		event.App.Logger().Error("manual backup failed", "error", err)
		return httpx.WriteError(event, result.Conflict("backup.failed", "The backup could not be created. Try again when the host is idle."))
	}
	backups, err := backupFiles(event.App)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	for _, backup := range backups {
		if backup["id"] == name {
			// Backup creation is an external filesystem operation and cannot be
			// rolled back with SQLite. Audit failure is therefore best-effort;
			// the durable backup is retained and the failure is logged.
			_ = applicationaudit.Record(event.App, event.Auth, "", "backup.created", "backup", name,
				nil, event.Get(httpx.TraceIDKey))
			return event.JSON(http.StatusCreated, backup)
		}
	}
	return event.JSON(http.StatusCreated, map[string]any{"id": name})
}

func listBackups(event *core.RequestEvent) error {
	backups, err := backupFiles(event.App)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	return event.JSON(http.StatusOK, backups)
}

func restoreBackup(event *core.RequestEvent) error {
	name := event.Request.PathValue("id")
	var request restoreRequest
	if err := event.BindBody(&request); err != nil || request.Confirmation != "RESTORE "+name {
		return httpx.WriteError(event, result.Invalid("backup.confirmation_required", `Type "RESTORE `+name+`" to confirm.`, nil))
	}
	fsys, err := event.App.NewBackupsFilesystem()
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	defer fsys.Close()
	exists, err := fsys.Exists(name)
	if err != nil || !exists {
		return httpx.WriteError(event, result.AppError{Code: "backup.not_found", Message: "Backup not found.", Status: http.StatusNotFound})
	}
	rollback := "pre_restore_sgh_" + time.Now().UTC().Format("20060102_150405") + ".zip"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	if err := event.App.CreateBackup(ctx, rollback); err != nil {
		return httpx.WriteError(event, result.Conflict("backup.rollback_failed", "A rollback backup could not be created, so restore was cancelled."))
	}
	if err := recovery.Schedule(event.App, name); err != nil {
		return httpx.WriteError(event, result.Conflict("backup.restore_unavailable", "The restore could not be scheduled. The current data remains unchanged."))
	}
	// Restore scheduling crosses a process/filesystem boundary. The scheduled
	// compensating rollback backup remains available if this best-effort audit fails.
	_ = applicationaudit.Record(event.App, event.Auth, "", "backup.restore_scheduled", "backup", name,
		map[string]any{"rollbackBackup": rollback}, event.Get(httpx.TraceIDKey))
	return event.JSON(http.StatusAccepted, map[string]any{
		"status": "restarting", "backup": name, "rollbackBackup": rollback,
	})
}

func EnsureSettings(app core.App) (*core.Record, error) {
	record, err := app.FindRecordById("host_settings", settingsID)
	if err == nil {
		return record, nil
	}
	collection, err := app.FindCollectionByNameOrId("host_settings")
	if err != nil {
		return nil, err
	}
	record = core.NewRecord(collection)
	record.Id = settingsID
	record.Set("port", 8090)
	record.Set("bind_address", "0.0.0.0")
	record.Set("trusted_lan_acknowledged", false)
	record.Set("automatic_backups", true)
	if err := app.Save(record); err != nil {
		return nil, err
	}
	return record, nil
}

func projectSettings(app core.App, record *core.Record) map[string]any {
	return map[string]any{
		"port": record.GetInt("port"), "bindAddress": record.GetString("bind_address"),
		"preferredAdapter":       record.GetString("preferred_adapter"),
		"trustedLanAcknowledged": record.GetBool("trusted_lan_acknowledged"),
		"automaticBackups":       record.GetBool("automatic_backups"),
		"privateAddresses":       privateAddresses(),
		"restartRequired":        true,
		"lastRestore":            recovery.LastReport(app.DataDir()),
	}
}

func privateAddresses() []map[string]string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return []map[string]string{}
	}
	result := make([]map[string]string, 0)
	for _, adapter := range interfaces {
		if adapter.Flags&net.FlagUp == 0 || adapter.Flags&net.FlagLoopback != 0 {
			continue
		}
		addresses, _ := adapter.Addrs()
		for _, address := range addresses {
			ip, _, err := net.ParseCIDR(address.String())
			if err == nil && ip.To4() != nil && ip.IsPrivate() {
				result = append(result, map[string]string{"adapter": adapter.Name, "address": ip.String()})
			}
		}
	}
	return result
}

func preferredPrivateAddress(adapterName string) string {
	addresses := privateAddresses()
	for _, address := range addresses {
		if address["adapter"] == adapterName {
			return address["address"]
		}
	}
	if len(addresses) > 0 {
		return addresses[0]["address"]
	}
	return ""
}

func backupFiles(app core.App) ([]map[string]any, error) {
	fsys, err := app.NewBackupsFilesystem()
	if err != nil {
		return nil, err
	}
	defer fsys.Close()
	files, err := fsys.List("")
	if err != nil {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].ModTime.After(files[j].ModTime) })
	result := make([]map[string]any, 0, len(files))
	for _, file := range files {
		if strings.HasSuffix(file.Key, ".attrs") {
			continue
		}
		result = append(result, map[string]any{
			"id": file.Key, "size": file.Size, "modifiedAt": file.ModTime.UTC(),
			"automatic": strings.HasPrefix(file.Key, "auto_sgh_"),
		})
	}
	return result, nil
}

func EnsureDailyBackup(app core.App) error {
	name := "auto_sgh_" + time.Now().UTC().Format("20060102") + ".zip"
	fsys, err := app.NewBackupsFilesystem()
	if err != nil {
		return err
	}
	exists, err := fsys.Exists(name)
	fsys.Close()
	if err == nil && exists {
		return nil
	}
	if err := app.CreateBackup(context.Background(), name); err != nil {
		return err
	}
	return retainAutomaticBackups(app, 7)
}

func retainAutomaticBackups(app core.App, keep int) error {
	fsys, err := app.NewBackupsFilesystem()
	if err != nil {
		return err
	}
	defer fsys.Close()
	files, err := fsys.List("")
	if err != nil {
		return err
	}
	automatic := make([]string, 0)
	for _, file := range files {
		if strings.HasPrefix(file.Key, "auto_sgh_") && strings.HasSuffix(file.Key, ".zip") {
			automatic = append(automatic, file.Key)
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(automatic)))
	if len(automatic) <= keep {
		return nil
	}
	for _, name := range automatic[keep:] {
		if err := fsys.Delete(name); err != nil {
			return fmt.Errorf("delete expired automatic backup %s: %w", name, err)
		}
	}
	return nil
}
