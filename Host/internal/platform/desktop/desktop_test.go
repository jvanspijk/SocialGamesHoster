package desktop

import (
	"path/filepath"
	"testing"
)

func TestSingleInstanceScopeFollowsTheDataDirectory(t *testing.T) {
	t.Parallel()
	dataDir := filepath.Join(t.TempDir(), "party-data")
	sameDataDir := filepath.Join(dataDir, ".")
	otherDataDir := filepath.Join(t.TempDir(), "party-data")

	if instanceID(dataDir) != instanceID(sameDataDir) {
		t.Fatal("equivalent data-directory paths must share a single-instance scope")
	}
	if instanceID(dataDir) == instanceID(otherDataDir) {
		t.Fatal("isolated data directories must not block each other")
	}
}
