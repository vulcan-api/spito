package api

import "testing"

func TestPathExists(t *testing.T) {
	testPath := "/home"
	pathExists := PathExists(testPath)

	if !pathExists {
		t.Fatalf("Path '%s' doesn't exist", testPath)
	}

	testPath = "/etc/init.d/dbus"
	pathExists = PathExists(testPath)

	if !pathExists {
		t.Fatalf("File '%s' doesn't exist", testPath)
	}
}

func TestFileExists(t *testing.T) {
	testPath := "/home"
	pathExists := FileExists(testPath, true)

	if !pathExists {
		t.Fatalf("Path '%s' doesn't exist or it's file", testPath)
	}

	testPath = "/etc/init.d/dbus"
	pathExists = FileExists(testPath, false)

	if !pathExists {
		t.Fatalf("File '%s' doesn't exist or it's directory", testPath)
	}
}
