package api

import (
	"testing"
)

func TestPackageMatrix(t *testing.T) {
	testPackages := []string{
		"pacman",
		// Commented for test development speed
		//"curl",
		//"bash",
	}

	for _, packageName := range testPackages {
		testPackageInfo(packageName, t)
	}
}

func testPackageInfo(packageName string, t *testing.T) {
	p := Package{}
	p.Get(packageName)

	if p.Name == "" {
		t.Fatalf("Couldn't resolve \"%s\" package name", packageName)
	}

	if p.Version == "" {
		t.Fatalf("Couldn't resolve \"%s\" package version", packageName)
	}

	if p.InstallDate == "" {
		t.Fatalf("Couldn't resolve \"%s\" package install date", packageName)
	}
}
