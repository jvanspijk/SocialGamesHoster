//go:build windows

package desktop

import "testing"

func TestSingleInstanceLockContract(t *testing.T) {
	t.Run("same data directory rejects second host", func(t *testing.T) {
		dataDir := t.TempDir()
		acquireSingleInstance(t, dataDir, false, true)
		acquireSingleInstance(t, dataDir, false, false)
	})

	t.Run("distinct explicit data directories coexist", func(t *testing.T) {
		acquireSingleInstance(t, t.TempDir(), false, true)
		acquireSingleInstance(t, t.TempDir(), false, true)
	})

	t.Run("installer-managed hosts collide across data directories", func(t *testing.T) {
		acquireSingleInstance(t, t.TempDir(), true, true)
		acquireSingleInstance(t, t.TempDir(), true, false)
	})
}

func acquireSingleInstance(t *testing.T, dataDir string, installerManaged, wantFirst bool) {
	t.Helper()
	first, release, err := AcquireSingleInstance(dataDir, installerManaged)
	if err != nil {
		t.Fatalf("AcquireSingleInstance(%q, installerManaged=%v): %v", dataDir, installerManaged, err)
	}
	if first != wantFirst {
		t.Fatalf("AcquireSingleInstance(%q, installerManaged=%v) first = %v, want %v", dataDir, installerManaged, first, wantFirst)
	}
	if first {
		t.Cleanup(release)
	}
}
