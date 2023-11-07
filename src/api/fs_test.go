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

func TestReadFile(t *testing.T) {
	content, err := ReadFile(testFile)
	if err != nil {
		t.Fatalf("Error occured during opening file: %s", fmt.Sprint(err))
	}
	if content == "" {
		t.Fatal("File is empty")
	}
}

func TestReadDir(t *testing.T) {
	entries, err := ReadDir(testDir)
	if err != nil {
		t.Fatalf("Error occured during getting entries: %s", fmt.Sprint(err))
	}

	if len(entries) == 0 {
		t.Fatal("Directory is empty")
	}
}

func TestRemoveComments(t *testing.T) {
	const testFileWithComments = "example file #example /*comment\nexample data /* and\ncomment */"
	const testFileWithoutComment = "example file\nexample data"

	file := RemoveComments(testFileWithComments, "#", "/*", "*/")
	if file != testFileWithoutComment {
		t.Fatal("Output file doesn't match given one")
	}
}

func TestFileContains(t *testing.T) {
	const testFileContent = "example data"
	const testData = "data"

	contains := FileContains(testFileContent, testData)
	if contains != true {
		t.Fatal("Function haven't returned true")
	}
}
