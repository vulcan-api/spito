package tests

import (
	"github.com/nasz-elektryk/spito/vrct"
	"os"
	"testing"
)

func TestCreatingFile(t *testing.T) {
	ruleVrct, err := vrct.NewRuleVRCT()
	if err != nil {
		t.Fatal("Failed to Create VRCT instance")
	}
	fsVrct := &ruleVrct.Fs

	err = fsVrct.CreateFile("/tmp/spito-test-2137/file.txt", []byte("test value"), false)
	if err != nil {
		t.Fatal("Failed to create file /tmp/spito-test-2137/file.txt\n", err)
	}

	err = fsVrct.CreateFile("/tmp/spito-test-2137/file.txt", []byte("this should be overridden"), true)
	if err != nil {
		t.Fatal("Failed trying to override file /tmp/spito-test-2137/file.txt\n", err)
	}

	file, err := fsVrct.ReadFile("/tmp/spito-test-2137/file.txt")
	if err != nil {
		t.Fatal("Failed to read file /tmp/spito-test-2137/file.txt\n", err)
	}

	if string(file) != "test value" {
		t.Fatal("Failed to properly simulate /tmp/spito-test-2137 file content")
	}

	err = fsVrct.Apply()
	if err != nil {
		t.Fatal("Failed to apply VRCT\n", err)
	}

	file, err = os.ReadFile("/tmp/spito-test-2137/file.txt")
	if err != nil {
		t.Fatal("Failed to read from real fs file /tmp/spito-test-2137/file.txt\n", err)
	}

	if string(file) != "test value" {
		t.Fatal("Failed to properly merge /tmp/spito-test-2137 file content")
	}

	// cleanup
	_ = os.RemoveAll("/tmp/spito-test-2137")
}
