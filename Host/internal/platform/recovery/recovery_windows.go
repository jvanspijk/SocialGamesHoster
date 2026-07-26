//go:build windows

package recovery

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

func schedule(_ core.App, name string) error {
	return scheduleConfiguredHandler(name)
}

func LaunchHelper(dataDir, backupName string) error {
	configDir := filepath.Join(filepath.Dir(dataDir), "config")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return err
	}
	markerPath := filepath.Join(configDir, "pending-restore.json")
	data, err := json.Marshal(Marker{DataDir: dataDir, BackupName: backupName})
	if err != nil {
		return err
	}
	if err := os.WriteFile(markerPath, data, 0o600); err != nil {
		return err
	}
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	command := exec.Command(executable, "--complete-restore="+markerPath)
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return command.Start()
}

func Complete(markerPath string) (finalErr error) {
	markerPath, err := filepath.Abs(markerPath)
	if err != nil {
		return err
	}
	configDir := filepath.Dir(markerPath)
	backupName := ""
	defer func() {
		report := Report{
			Status: "success", BackupName: backupName, Message: "The backup was restored successfully.",
			FinishedAt: time.Now().UTC(),
		}
		if finalErr != nil {
			report.Status = "failed"
			report.Message = "The backup could not be restored. The previous data was preserved."
		}
		writeReport(configDir, report)
	}()
	markerBytes, err := os.ReadFile(markerPath)
	if err != nil {
		return err
	}
	var marker Marker
	if err := json.Unmarshal(markerBytes, &marker); err != nil {
		return err
	}
	backupName = filepath.Base(marker.BackupName)
	dataDir, err := filepath.Abs(marker.DataDir)
	if err != nil {
		return err
	}
	parent := filepath.Dir(dataDir)
	if dataDir == parent || filepath.Base(dataDir) == "." {
		return errors.New("restore target is invalid")
	}
	expectedConfigDir := filepath.Join(parent, "config")
	if filepath.Dir(markerPath) != expectedConfigDir || filepath.Base(dataDir) != "data" {
		return errors.New("restore marker is outside the application data directory")
	}
	backupPath := filepath.Join(dataDir, core.LocalBackupsDirName, filepath.Base(marker.BackupName))
	stage, err := os.MkdirTemp(parent, ".sgh-restore-stage-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	if err := extractBackup(backupPath, stage); err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(stage, "data.db")); err != nil {
		return errors.New("backup does not contain data.db")
	}

	oldDir := filepath.Join(parent, ".sgh-restore-rollback-"+time.Now().UTC().Format("20060102-150405"))
	var moveErr error
	for attempt := 0; attempt < 60; attempt++ {
		moveErr = os.Rename(dataDir, oldDir)
		if moveErr == nil {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	if moveErr != nil {
		return fmt.Errorf("could not stop the previous data process: %w", moveErr)
	}
	if err := os.Rename(stage, dataDir); err != nil {
		_ = os.Rename(oldDir, dataDir)
		return fmt.Errorf("could not activate restored data: %w", err)
	}
	oldBackups := filepath.Join(oldDir, core.LocalBackupsDirName)
	newBackups := filepath.Join(dataDir, core.LocalBackupsDirName)
	if _, err := os.Stat(oldBackups); err == nil {
		if err := os.Rename(oldBackups, newBackups); err != nil {
			_ = os.Rename(dataDir, stage)
			_ = os.Rename(oldDir, dataDir)
			return fmt.Errorf("could not preserve backup files: %w", err)
		}
	}
	if err := os.Remove(markerPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	_ = os.RemoveAll(oldDir)
	return Relaunch(false)
}

func extractBackup(backupPath, destination string) error {
	reader, err := zip.OpenReader(backupPath)
	if err != nil {
		return err
	}
	defer reader.Close()
	if len(reader.File) > 10_000 {
		return errors.New("backup contains too many files")
	}
	var total uint64
	for _, entry := range reader.File {
		clean := filepath.Clean(entry.Name)
		if clean == "." || filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(filepath.Separator)) ||
			strings.Contains(entry.Name, `\`) || entry.FileInfo().Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("backup contains unsafe path %q", entry.Name)
		}
		target := filepath.Join(destination, clean)
		relative, err := filepath.Rel(destination, target)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return fmt.Errorf("backup path escapes restore directory")
		}
		if entry.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o700); err != nil {
				return err
			}
			continue
		}
		if entry.UncompressedSize64 > 256<<20 {
			return fmt.Errorf("backup entry %q exceeds the safety limit", entry.Name)
		}
		total += entry.UncompressedSize64
		if total > 2<<30 {
			return errors.New("backup exceeds the total restore safety limit")
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return err
		}
		source, err := entry.Open()
		if err != nil {
			return err
		}
		destinationFile, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			source.Close()
			return err
		}
		written, copyErr := io.Copy(destinationFile, io.LimitReader(source, int64(entry.UncompressedSize64)+1))
		closeErr := destinationFile.Close()
		sourceCloseErr := source.Close()
		if copyErr != nil || closeErr != nil || sourceCloseErr != nil {
			return errors.Join(copyErr, closeErr, sourceCloseErr)
		}
		if written != int64(entry.UncompressedSize64) {
			return fmt.Errorf("backup entry %q has an invalid decompressed size", entry.Name)
		}
	}
	return nil
}

func Relaunch(diagnostics bool) error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	arguments := []string{}
	if diagnostics {
		arguments = append(arguments, "--diagnostics")
	}
	command := exec.Command(executable, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return command.Start()
}
