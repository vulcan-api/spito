package vrct

import (
	"github.com/avorty/spito/pkg/vrct/vrctFs"
	"os"
	"testing"
)

func TestVRCTFs(t *testing.T) {
	ruleVrct, err := NewRuleVRCT()
	if err != nil {
		t.Fatal("Failed to Create VRCT instance")
	}
	fsVrct := &ruleVrct.Fs

	tmpPath, err := os.MkdirTemp("/tmp", "spito-test-")
	if err != nil {
		t.Fatal("Failed to create temporary test directory\n", err)
	}

	testFilePath := tmpPath + "/new_dir/file.txt"

	err = fsVrct.CreateFile(testFilePath, []byte("test value"), false, vrctFs.TextFile)
	if err != nil {
		t.Fatal("Failed to create file "+testFilePath+"\n", err)
	}

	err = fsVrct.CreateFile(testFilePath, []byte("this should be overridden"), true, vrctFs.TextFile)
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

	//testConfigPath := tmpPath + "/new_dir/config.json"

	//err = fsVrct.CreateFile(testConfigPath, []byte(`{"name":"john"}`), false, vrctFs.JsonConfig)
	//if err != nil {
	//	t.Fatal("Failed to create file "+testConfigPath+"\n", err)
	//}

	//err = fsVrct.CreateFile(testConfigPath, []byte(`"lastname":"mcqueen"`), true, vrctFs.JsonConfig)
	//if err != nil {
	//	t.Fatal("Failed trying to override file "+testConfigPath+"\n", err)
	//}

	//config, err := fsVrct.ReadFile(testConfigPath)
	//if err != nil {
	//	t.Fatal("Failed to read file "+testConfigPath+"\n", err)
	//}

	//if strings.TrimSpace(string(config)) != `{"name":"john","lastname":"mcqueen"}` {
	//	t.Fatal("Failed to properly simulate " + testConfigPath + " file content")
	//}

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
