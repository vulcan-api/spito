package tests

import (
	"github.com/nasz-elektryk/spito/vrct"
	"testing"
)

func TestCreatingFile(t *testing.T) {
	ruleVrct, err := vrct.NewRuleVRCT()
	if err != nil {
		t.Fatal("Failed to Create VRCT instance")
	}
	fsVrct := &ruleVrct.Fs

	err = fsVrct.CreateFile("/test/file.txt", []byte("test value"), false)
	if err != nil {
		t.Fatal("Failed to create file /test/file.txt\n", err)
	}

	err = fsVrct.CreateFile("/test/file.txt", []byte("this should be overridden"), true)
	if err != nil {
		t.Fatal("Failed trying to override file /test/file.txt\n", err)
	}

	file, err := fsVrct.ReadFile("/test/file.txt")
	if err != nil {
		t.Fatal("Failed to read file /test/file.txt\n", err)
	}

	if string(file) != "test value" {
		t.Fatal("Failed to properly simulate /test/file.txt file content")
	}
}
