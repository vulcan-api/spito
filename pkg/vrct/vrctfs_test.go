package vrct

import (
	"github.com/avorty/spito/pkg/vrct/vrctFs"
	"github.com/nsf/jsondiff"
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

	err = fsVrct.CreateFile(testFilePath, []byte("this should be overridden"), nil, true, vrctFs.TextFile)
	if err != nil {
		t.Fatal("Failed trying to override file "+testFilePath+"\n", err)
	}

	err = fsVrct.CreateFile(testFilePath, []byte("test value"), nil, false, vrctFs.TextFile)
	if err != nil {
		t.Fatal("Failed to create file "+testFilePath+"\n", err)
	}

	file, err := fsVrct.ReadFile(testFilePath)
	if err != nil {
		t.Fatal("Failed to read file "+testFilePath+"\n", err)
	}

	if string(file) != "test value" {
		t.Fatal("Failed to properly simulate " + testFilePath + " file content")
	}

	testConfigPath := tmpPath + "/new_dir/config.json"

	err = fsVrct.CreateFile(testConfigPath, []byte(`{"subObject": {"key": "value"}}`), nil, false, vrctFs.JsonConfig)
	if err != nil {
		t.Fatal("Failed to create file "+testConfigPath+"\n", err)
	}

	err = fsVrct.CreateFile(testConfigPath, []byte(`{"key":"value", "subObject" : {"nextKey": "value"}}`), []byte(`{"key": true}`), false, vrctFs.JsonConfig)
	if err != nil {
		t.Fatal("Failed trying to override file "+testConfigPath+"\n", err)
	}

	err = fsVrct.CreateFile(testConfigPath, []byte(`{"key":"value", "nextKey": "nextValue"}`), nil, false, vrctFs.JsonConfig)
	if err != nil {
		t.Fatal("Failed trying to override file "+testConfigPath+"\n", err)
	}
	// TODO: better tests

	//err = fsVrct.CreateFile(testConfigPath, []byte(`{"key":"value"}`), false, vrctFs.JsonConfig)
	//if !errors.Is(err, vrctFs.ErrConfigsCannotBeMerged) {
	//	t.Fatal("Failed trying to override file "+testConfigPath+"\n", err)
	//}

	config, err := fsVrct.ReadFile(testConfigPath)
	if err != nil {
		t.Fatal("Failed to read file "+testConfigPath+"\n", err)
	}

	res, _ := jsondiff.Compare([]byte(`{"key": "value", "subObject": {"nextKey":"value", "key": "value"}, "nextKey": "nextValue"}`), config, &jsondiff.Options{})
	if res != jsondiff.FullMatch {
		t.Fatal("Failed to properly simulate " + testConfigPath + " file content")
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

	config, err = os.ReadFile(testConfigPath)
	if err != nil {
		t.Fatal("Failed to read from real fs file "+testConfigPath+"\n", err)
	}

	res, _ = jsondiff.Compare([]byte(`{"key": "value", "subObject": {"nextKey":"value", "key": "value"}, "nextKey": "nextValue"}`), config, &jsondiff.Options{})
	if res != jsondiff.FullMatch {
		t.Fatal("Failed to properly simulate " + testConfigPath + " file content")
	}

	// cleanup
	_ = os.RemoveAll(tmpPath)
}
