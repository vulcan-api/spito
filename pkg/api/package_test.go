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

/*
func TestInstallPackages(t *testing.T) {
	err := InstallPackages("opentimer", "vim")
	if err != nil {
		t.Fatalf("the installation of packages 'opentimer' and 'vim' failed")
	}
}

func TestRemovePackages(t *testing.T) {
	err := RemovePackages("opentimer")
	if err != nil {
		t.Fatalf("couldn't remove package 'opentimer'")
	}
}
*/
