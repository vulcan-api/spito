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
	const testFileWithComments = "example/*multi*/#single\ndata/*multi*/#single"
	const testFileWithoutComments = "example\ndata"
	const testFileWithoutSingleLineComment = "example/*multi*/\ndata/*multi*/"
	const testFileWithoutMultilineComment = "example#single\ndata#single"

	file := RemoveComments(testFileWithComments, "#", "/*", "*/")
	if file != testFileWithoutComments {
		t.Fatal("Transformed file doesn't match given one")
	}

	file = RemoveComments(testFileWithComments, "#", "", "")
	if file != testFileWithoutSingleLineComment {
		t.Fatal("Transformed file doesn't match given one")
	}

	file = RemoveComments(testFileWithComments, "", "/*", "*/")
	if file != testFileWithoutMultilineComment {
		t.Fatal("Transformed file doesn't match given one")
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

const regexTestData = "peach"
const multilineRegexTestData = "peach\nlunch\npinch"
const testRegex = "p([a-z]+)ch"

func TestFind(t *testing.T) {
	index, err := Find(testRegex, regexTestData)
	if err != nil {
		t.Fatalf("Your regex is broken: %s", fmt.Sprint(err))
	}
	if index == nil {
		t.Fatal("Test regex doesn't match test string")
	}
}

func TestFindAll(t *testing.T) {
	indexes, err := FindAll(testRegex, multilineRegexTestData)
	if err != nil {
		t.Fatalf("Your regex is broken: %s", fmt.Sprint(err))
	}
	t.Log(indexes)
	if indexes == nil || indexes[0] == nil || indexes[1] == nil {
		t.Fatal("Test regex doesn't match test string")
	}
}

func TestGetProperLines(t *testing.T) {
	multilineRegexResults := []string{"peach", "pinch"}
	lines, err := GetProperLines(testRegex, multilineRegexTestData)
	if err != nil {
		t.Fatalf("Your regex is broken %s", fmt.Sprint(err))
	}
	for i, testLine := range lines {
		properLine := multilineRegexResults[i]
		if testLine != properLine {
			t.Fatalf("Your %dth line doesn't match specified one\nproper: '%s'\ngiven: '%s'", i, properLine, testLine)
		}
	}
}
