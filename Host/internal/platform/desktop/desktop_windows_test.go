//go:build windows

package desktop

import "testing"

func TestSingleInstanceLockContract(t *testing.T) {
	first, releaseFirst, err := AcquireSingleInstance(t.TempDir(), false)
	if err != nil || !first {
		t.Fatalf("first isolated host = (%v, %v), want acquired", first, err)
	}
	defer releaseFirst()

	second, releaseSecond, err := AcquireSingleInstance(t.TempDir(), false)
	if err != nil || !second {
		t.Fatalf("a different isolated data directory = (%v, %v), want acquired", second, err)
	}
	releaseSecond()
}
