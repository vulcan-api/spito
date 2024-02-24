package tests

import (
	"github.com/avorty/spito/cmd/cmdApi"
	"github.com/avorty/spito/internal/checker"
	"github.com/avorty/spito/pkg/vrct"
	"github.com/avorty/spito/pkg/vrct/vrctFs"
	"os"
	"path/filepath"
	"testing"
)

const newContent = "new test content"
const originalContent = "original content"

func TestCreatingFile(t *testing.T) {
	ruleVrct, err := vrct.NewRuleVRCT()
	if err != nil {
		t.Fatal("Failed to Create VRCT instance")
	}
	fsVrct := &ruleVrct.Fs

	defer func() {
		if err := ruleVrct.DeleteRuntimeTemp(); err != nil {
			t.Fatal("Failed to remove temporary VRCT files", err.Error())
		}
	}()

	tmpPath, err := os.MkdirTemp("/tmp", "spito-test-")
	if err != nil {
		t.Fatal("Failed to create temporary test directory\n", err.Error())
	}

	testFilePath := tmpPath + "/new_dir/file.txt"

	// We create file right now in order to check if VRCT backup works
	_ = os.MkdirAll(filepath.Dir(testFilePath), os.ModePerm)
	testFile, err := os.Create(testFilePath)
	if err != nil {
		t.Fatal("Failed to create test file in \""+testFilePath+"\" this means test is broken not spito\n", err.Error())
	}

	if _, err = testFile.Write([]byte(originalContent)); err != nil {
		t.Fatal("Failed to write content to test file in \""+testFilePath+"\" this means test is broken not spito\n", err.Error())
	}

	if err := testFile.Close(); err != nil {
		t.Fatal("Failed to close test file in \""+testFilePath+"\" this means test is broken not spito\n", err.Error())
	}

	revertNum := makeFsChanges(t, fsVrct, testFilePath)
	revertFsChanges(t, testFilePath, revertNum)

	// cleanup
	_ = os.RemoveAll(tmpPath)
}

// Returns revertNum
func makeFsChanges(t *testing.T, fsVrct *vrctFs.VRCTFs, testFilePath string) int {
	err := fsVrct.CreateFile(testFilePath, []byte(newContent), false)
	if err != nil {
		t.Fatal("Failed to create file "+testFilePath+"\n", err)
	}

	err = fsVrct.CreateFile(testFilePath, []byte(newContent), false)
	if err != nil {
		t.Fatal("Failed to create file "+testFilePath+"\n", err)
	}

	err = fsVrct.CreateFile(testFilePath, []byte("this should result in error"), false)
	if err == nil {
		t.Fatalf("something is wrong with merging: %s", err)
	}

	err = fsVrct.CreateFile(testFilePath, []byte("this should be overridden"), true)
	if err != nil {
		t.Fatal("Failed trying to override file "+testFilePath+"\n", err)
	}

	file, err := fsVrct.ReadFile(testFilePath)
	if err != nil {
		t.Fatal("Failed to read file "+testFilePath+"\n", err)
	}

	if string(file) != newContent {
		t.Logf("content:\"%s\"\n", string(file))
		t.Logf("expected content: \"%s\"\n\n", newContent)
		t.Fatal("Failed to properly simulate " + testFilePath + " file content")
	}

	revertNum, err := fsVrct.Apply([]vrctFs.Rule{})
	if err != nil {
		t.Fatal("Failed to apply VRCT\n", err)
	}

	file, err = os.ReadFile(testFilePath)
	if err != nil {
		t.Fatal("Failed to read from real fs file "+testFilePath+"\n", err)
	}

	if string(file) != newContent {
		t.Logf("content:\"%s\"\n", string(file))
		t.Logf("expected content: \"%s\"\n\n", newContent)
		t.Fatal("Failed to properly merge " + testFilePath + " file content")
	}

	return revertNum
}

func revertFsChanges(t *testing.T, testFilePath string, revertNum int) {
	revertSteps, err := vrctFs.NewRevertSteps()
	if err != nil {
		t.Fatalf("Failed to initialize RevertSteps\n%s", err.Error())
	}

	if err := revertSteps.Deserialize(revertNum); err != nil {
		t.Fatalf(
			"Failed to deserialize RevertSteps using %d revert number \n%s",
			revertNum,
			err.Error(),
		)
	}

	revertFn := checker.GetRevertRuleFn(cmdApi.InfoApi{})

	if err := revertSteps.Apply(revertFn); err != nil {
		t.Fatalf("Failed to revert VRCT\n%s", err.Error())
	}

	file, err := os.ReadFile(testFilePath)
	if err != nil {
		t.Fatalf("Failed to read from real fs file %s\n%s", testFilePath, err.Error())
	}

	if string(file) != originalContent {
		t.Logf("content:\"%s\"\n", string(file))
		t.Logf("expected content: \"%s\"\n\n", originalContent)
		t.Fatalf("Failed to properly revert %s file content", testFilePath)
	}
}
