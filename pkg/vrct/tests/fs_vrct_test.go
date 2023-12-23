package tests

import (
	"github.com/nasz-elektryk/spito/pkg/vrct"
	"os"
	"testing"
)

func TestCreatingFile(t *testing.T) {
	ruleVrct, err := vrct.NewRuleVRCT()
	if err != nil {
		t.Fatal("Failed to Create VRCT instance")
	}
	fsVrct := &ruleVrct.Fs

	tmpPath, err := os.MkdirTemp("/tmp", "spito-test-")
	if err != nil {
		t.Fatal("Failed to create temporary test directory\n", err)
	}

	testFilePath := tmpPath + "/new_dir/file.txt"

	err = fsVrct.CreateFile(testFilePath, []byte("test value"), false)
	if err != nil {
		t.Fatal("Failed to create file "+testFilePath+"\n", err)
	}

	err = fsVrct.CreateFile(testFilePath, []byte("this should be overridden"), true)
	if err != nil {
		t.Fatal("Failed trying to override file "+testFilePath+"\n", err)
	}

	file, err := fsVrct.ReadFile(testFilePath)
	if err != nil {
		t.Fatal("Failed to read file "+testFilePath+"\n", err)
	}

	if string(file) != "test value" {
		t.Fatal("Failed to properly simulate " + testFilePath + " file content")
	}

	err = fsVrct.Apply()
	if err != nil {
		t.Fatal("Failed to apply VRCT\n", err)
	}

	file, err = os.ReadFile(testFilePath)
	if err != nil {
		t.Fatal("Failed to read from real fs file "+testFilePath+"\n", err)
	}

	if string(file) != "test value" {
		t.Fatal("Failed to properly merge " + testFilePath + " file content")
	}

	// cleanup
	_ = os.RemoveAll(tmpPath)
}
