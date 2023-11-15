package api

import (
	"testing"
)

func TestPackageMatrix(t *testing.T) {
	testPackages := []string{
		"pacman",
		"curl",
		"bash",
	}

	for _, packageName := range testPackages {
		testPackageInfo(packageName, t)
	}
}

func testPackageInfo(packageName string, t *testing.T) {
	p, err := GetPackage(packageName)
	if err != nil {
		panic(err)
	}

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
