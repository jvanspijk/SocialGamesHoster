package recovery

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

var (
	handlerMu sync.RWMutex
	handler   func(string)
)

func ConfigureRestoreHandler(callback func(string)) {
	handlerMu.Lock()
	defer handlerMu.Unlock()
	handler = callback
}

func scheduleConfiguredHandler(name string) error {
	handlerMu.RLock()
	callback := handler
	handlerMu.RUnlock()
	if callback == nil {
		return errors.New("restore lifecycle handler is unavailable")
	}
	go callback(name)
	return nil
}

type Marker struct {
	DataDir    string `json:"dataDir"`
	BackupName string `json:"backupName"`
}

type Report struct {
	Status     string    `json:"status"`
	BackupName string    `json:"backupName,omitempty"`
	Message    string    `json:"message"`
	FinishedAt time.Time `json:"finishedAt"`
}

func LastReport(dataDir string) *Report {
	content, err := os.ReadFile(filepath.Join(filepath.Dir(dataDir), "config", "last-restore.json"))
	if err != nil {
		return nil
	}
	var report Report
	if json.Unmarshal(content, &report) != nil {
		return nil
	}
	return &report
}

func writeReport(configDir string, report Report) {
	if os.MkdirAll(configDir, 0o700) != nil {
		return
	}
	content, err := json.Marshal(report)
	if err != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(configDir, "last-restore.json"), content, 0o600)
}

func Schedule(app core.App, name string) error {
	return schedule(app, name)
}
