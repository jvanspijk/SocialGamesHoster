//go:build !windows

package recovery

import (
	"context"

	"github.com/pocketbase/pocketbase/core"
)

func schedule(app core.App, name string) error {
	go app.RestoreBackup(context.Background(), name)
	return nil
}

func LaunchHelper(string, string) error {
	return nil
}

func Complete(string) error {
	return nil
}

func Relaunch(bool) error {
	return nil
}
