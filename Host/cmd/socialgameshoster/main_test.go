package main

import "testing"

func TestExplicitDataDirectoryIsIsolatedFromTheInstalledHost(t *testing.T) {
	t.Setenv("SGH_DATA_DIR", "")

	installed := parseOptions(nil)
	if !installed.installerManaged {
		t.Fatal("the default data directory must participate in installer shutdown checks")
	}

	isolated := parseOptions([]string{"--dir", t.TempDir()})
	if isolated.installerManaged {
		t.Fatal("an explicit isolated data directory must not block the installed host")
	}
}
