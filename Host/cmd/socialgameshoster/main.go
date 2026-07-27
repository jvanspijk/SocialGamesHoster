package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/embedded"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/achievements"
	authfeature "github.com/jvanspijk/SocialGamesHoster/Host/internal/features/auth"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/chat"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/diagnostics"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/games"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/owner"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/profiles"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/setup"
	timerfeature "github.com/jvanspijk/SocialGamesHoster/Host/internal/features/timer"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/desktop"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/realtime"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/recovery"
	_ "github.com/jvanspijk/SocialGamesHoster/Host/migrations"
)

var version = "dev"

type options struct {
	dataDir          string
	httpAddress      string
	diagnosticMode   bool
	withoutTray      bool
	completeRestore  string
	installerManaged bool
}

type serverController struct {
	app     core.App
	address string
	origins []string
	mu      sync.RWMutex
	server  *http.Server
	running bool
}

func main() {
	startedAt := time.Now().UTC()
	configuration := parseOptions(os.Args[1:])
	if configuration.completeRestore != "" {
		if err := recovery.Complete(configuration.completeRestore); err != nil {
			desktop.ShowError(
				"Social Games Hoster restore failed",
				"The backup could not be restored. The previous data was preserved.\n\n"+err.Error(),
			)
			_ = recovery.Relaunch(true)
		}
		return
	}
	first, release, err := desktop.AcquireSingleInstance(configuration.dataDir, configuration.installerManaged)
	if err != nil {
		log.Fatal(err)
	}
	if !first {
		_ = desktop.OpenURL(loopbackURL(configuration.httpAddress) + "admin")
		return
	}
	defer release()

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir:  configuration.dataDir,
		HideStartBanner: true,
	})
	profiles.RegisterJobs(app)
	owner.RegisterJobs(app)
	realtime.RegisterAuthorization(app)
	timerService := timerfeature.NewService(app)

	web, err := fs.Sub(embedded.Web, "web")
	if err != nil {
		log.Fatal(err)
	}
	controller := &serverController{app: app}
	recovery.ConfigureRestoreHandler(func(name string) {
		time.Sleep(time.Second)
		_ = controller.stop()
		_ = app.ResetBootstrapState()
		if err := recovery.LaunchHelper(app.DataDir(), name); err != nil {
			app.Logger().Error("restore helper could not be launched", "backup", name, "error", err)
			desktop.ShowError(
				"Social Games Hoster restore failed",
				"The restore helper could not be started. Your current data and rollback backup are unchanged.",
			)
		}
		os.Exit(0)
	})
	app.OnServe().BindFunc(func(event *core.ServeEvent) error {
		event.InstallerFunc = nil
		event.Router.BindFunc(httpx.SecurityMiddleware)
		setup.Register(event, version)
		authfeature.Register(event)
		profiles.Register(event)
		rulesets.RegisterRoutes(event, version)
		games.Register(event)
		owner.Register(event)
		timerfeature.Register(event, timerService)
		chat.Register(event)
		achievements.Register(event)
		diagnostics.Register(event, configuration.diagnosticMode, startedAt, version)
		event.Router.GET("/{path...}", apis.Static(web, true)).BindFunc(httpx.StaticCacheMiddleware)
		controller.attach(event.Server)
		timerService.Reconcile()
		return event.Next()
	})

	if err := app.Bootstrap(); err != nil {
		log.Fatal(err)
	}
	defer app.ResetBootstrapState()
	if err := owner.EnsurePreMigrationBackup(app, version); err != nil {
		app.Logger().Error("pre-migration backup failed", "error", err)
		log.Fatal("The safety backup could not be created. No upgrade was applied.")
	}
	if err := app.RunAllMigrations(); err != nil {
		app.Logger().Error("migration failed", "error", err)
		log.Fatal("The data upgrade could not be completed. The previous data was preserved.")
	}
	if app.Settings().Logs.MaxDays != 14 {
		app.Settings().Logs.MaxDays = 14
		if err := app.Save(app.Settings()); err != nil {
			log.Fatal("The local log-retention setting could not be applied.")
		}
	}
	runtimeSettings, err := owner.RuntimeConfiguration(app)
	if err != nil {
		log.Fatal(err)
	}
	if configuration.httpAddress == "" {
		configuration.httpAddress = fmt.Sprintf("%s:%d", runtimeSettings.BindAddress, runtimeSettings.Port)
	}
	controller.address = configuration.httpAddress
	localBase := loopbackURL(configuration.httpAddress)
	controller.origins = owner.AllowedOrigins(app)
	controller.origins = append(controller.origins, strings.TrimSuffix(localBase, "/"))

	if err := controller.start(); err != nil {
		desired := configuration.httpAddress
		fallback, fallbackErr := availableLoopbackAddress(8091, 8190)
		if fallbackErr != nil {
			desktop.ShowError(
				"Social Games Hoster could not start",
				"The configured network port is already in use and no local recovery port was available. Close the other application or restart Windows, then try again.",
			)
			log.Fatal("configured port unavailable")
		}
		controller.address = fallback
		configuration.httpAddress = fallback
		controller.origins = []string{"http://" + fallback}
		if fallbackErr = controller.start(); fallbackErr != nil {
			log.Fatal("local recovery page could not start")
		}
		desktop.ShowError(
			"Social Games Hoster port is occupied",
			"The configured address "+desired+" is already in use. The dashboard will open locally so the owner can choose another port under Installation.",
		)
	}
	if runtimeSettings.AutomaticBackups {
		go func() {
			if err := owner.EnsureDailyBackup(app); err != nil {
				app.Logger().Error("automatic startup backup failed", "error", err)
			}
		}()
	}
	hasOwner, err := app.CountRecords("game_masters")
	if err != nil {
		log.Fatal(err)
	}
	if desktop.Supported() && !configuration.withoutTray {
		go openWhenReady(localBase, func() {
			if hasOwner == 0 {
				_ = desktop.OpenURL(localBase)
			} else {
				_ = desktop.OpenURL(localBase + "admin")
			}
		})
		exit := make(chan struct{})
		exitOnce := sync.Once{}
		err = desktop.Run(desktop.Actions{
			DashboardURL:      func() string { return localBase + "admin" },
			JoinURL:           func() string { return owner.JoinURL(app) },
			DiagnosticsURL:    func() string { return localBase + "admin?tab=owner" },
			IsHosting:         controller.isRunning,
			StartHosting:      controller.start,
			StopHosting:       controller.stop,
			CreateBackup:      func() (string, error) { return owner.CreateManualBackup(app) },
			DiagnosticsActive: configuration.diagnosticMode,
			Exit:              func() { exitOnce.Do(func() { close(exit) }) },
		})
		exitOnce.Do(func() { close(exit) })
		<-exit
		if err != nil {
			app.Logger().Error("tray stopped unexpectedly", "error", err)
		}
		_ = controller.stop()
		return
	}

	waitForSignal()
	_ = controller.stop()
}

func (controller *serverController) attach(server *http.Server) {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	controller.server = server
}

func (controller *serverController) isRunning() bool {
	controller.mu.RLock()
	defer controller.mu.RUnlock()
	return controller.running
}

func (controller *serverController) start() error {
	controller.mu.Lock()
	if controller.running {
		controller.mu.Unlock()
		return nil
	}
	controller.running = true
	controller.mu.Unlock()
	result := make(chan error, 1)
	go func() {
		err := apis.Serve(controller.app, apis.ServeConfig{
			HttpAddr:        controller.address,
			ShowStartBanner: false,
			AllowedOrigins:  controller.origins,
		})
		controller.mu.Lock()
		controller.running = false
		controller.server = nil
		controller.mu.Unlock()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			controller.app.Logger().Error("hosting stopped unexpectedly", "address", controller.address, "error", err)
		}
		result <- err
	}()
	select {
	case err := <-result:
		return err
	case <-time.After(400 * time.Millisecond):
		return nil
	}
}

func (controller *serverController) stop() error {
	controller.mu.RLock()
	server := controller.server
	running := controller.running
	controller.mu.RUnlock()
	if !running {
		return nil
	}
	if server == nil {
		return errors.New("hosting is still starting")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return server.Shutdown(ctx)
}

func parseOptions(arguments []string) options {
	configuration := options{
		dataDir:          resolveDataDir(),
		installerManaged: os.Getenv("SGH_DATA_DIR") == "",
	}
	for index := 0; index < len(arguments); index++ {
		argument := arguments[index]
		switch {
		case argument == "--diagnostics":
			configuration.diagnosticMode = true
		case argument == "--no-tray":
			configuration.withoutTray = true
		case strings.HasPrefix(argument, "--complete-restore="):
			configuration.completeRestore = strings.TrimPrefix(argument, "--complete-restore=")
		case strings.HasPrefix(argument, "--http="):
			configuration.httpAddress = strings.TrimPrefix(argument, "--http=")
		case argument == "--http" && index+1 < len(arguments):
			index++
			configuration.httpAddress = arguments[index]
		case strings.HasPrefix(argument, "--dir="):
			configuration.dataDir = strings.TrimPrefix(argument, "--dir=")
			configuration.installerManaged = false
		case argument == "--dir" && index+1 < len(arguments):
			index++
			configuration.dataDir = arguments[index]
			configuration.installerManaged = false
		}
	}
	return configuration
}

func loopbackURL(address string) string {
	_, port, found := strings.Cut(address, ":")
	if !found {
		port = "8090"
	}
	if _, err := strconv.Atoi(port); err != nil {
		port = "8090"
	}
	return "http://127.0.0.1:" + port + "/"
}

func openWhenReady(baseURL string, ready func()) {
	client := http.Client{Timeout: 750 * time.Millisecond}
	for attempt := 0; attempt < 40; attempt++ {
		response, err := client.Get(baseURL + "api/app/v1/setup/status")
		if err == nil {
			response.Body.Close()
			if response.StatusCode == http.StatusOK {
				ready()
				return
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func waitForSignal() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	<-signals
}

func availableLoopbackAddress(firstPort, lastPort int) (string, error) {
	for port := firstPort; port <= lastPort; port++ {
		address := "127.0.0.1:" + strconv.Itoa(port)
		listener, err := net.Listen("tcp", address)
		if err != nil {
			continue
		}
		_ = listener.Close()
		return address, nil
	}
	return "", errors.New("no local recovery port is available")
}

func resolveDataDir() string {
	if value := os.Getenv("SGH_DATA_DIR"); value != "" {
		return value
	}
	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		return filepath.Join(localAppData, "SocialGamesHoster", "data")
	}
	return filepath.Join(".", "pb_data")
}
