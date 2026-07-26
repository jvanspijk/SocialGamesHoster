package desktop

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"strings"
)

type Actions struct {
	DashboardURL      func() string
	JoinURL           func() string
	DiagnosticsURL    func() string
	IsHosting         func() bool
	StartHosting      func() error
	StopHosting       func() error
	CreateBackup      func() (string, error)
	DiagnosticsActive bool
	Exit              func()
}

func instanceID(dataDir string) string {
	absolute, err := filepath.Abs(dataDir)
	if err != nil {
		absolute = filepath.Clean(dataDir)
	}
	normalized := strings.ToLower(filepath.Clean(absolute))
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:8])
}
