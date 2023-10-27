package api

import (
	"fmt"
	"testing"
)

const testDir = "/etc"
const testFile = "/etc/bash.bashrc"

func TestPathExists(t *testing.T) {
	pathExists := PathExists(testDir)

	if !pathExists {
		t.Fatalf("Path '%s' doesn't exist", testDir)
	}

	pathExists = PathExists(testFile)

	if !pathExists {
		t.Fatalf("File '%s' doesn't exist", testDir)
	}
}

func TestFileExists(t *testing.T) {
	dirExists := FileExists(testDir, true)

	if !dirExists {
		t.Fatalf("Path '%s' doesn't exist or it's file", testDir)
	}

	fileExists := FileExists(testFile, false)

	if !fileExists {
		t.Fatalf("File '%s' doesn't exist or it's directory", testFile)
	}
}

func TestGetFileContent(t *testing.T) {
	content, err := GetFileContent(testFile)
	if err != nil {
		t.Fatalf("Error occured during opening file: %s", fmt.Sprint(err))
	}
	if content == "" {
		t.Fatal("File is empty")
	}
}

func TestLS(t *testing.T) {
	entries, err := LS(testDir)
	if err != nil {
		t.Fatalf("Error occured during getting entries: %s", fmt.Sprint(err))
	}

	if len(entries) == 0 {
		t.Fatal("Directory is empty")
	}
}
