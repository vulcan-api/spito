package api

import "testing"

func TestPathExists(t *testing.T) {
	testPath := "/home"
	pathExists := PathExists(testPath)

	if !pathExists {
		t.Fatalf("Path '%s' doesn't exist", testPath)
	}

	testPath = "/etc/bash.bashrc"
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

	testPath = "/etc/bash.bashrc"
	pathExists = FileExists(testPath, false)

	if !pathExists {
		t.Fatalf("File '%s' doesn't exist or it's directory", testPath)
	}
}

func TestFileContains(t *testing.T) {
	testPath := "/etc/bash.bashrc"
	testContent := "shopt -s checkwinsize"

	contains, err := FileContains(testPath, testContent)
	if err != nil {
		t.Fatalf("Error occured during opening file: %s", err)
	}

	if !contains {
		t.Fatalf("Your string: %s doesn't exist in your file: %s", testContent, testPath)
	}
}
