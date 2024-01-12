package tests

import (
	"github.com/avorty/spito/pkg/vrct"
	"github.com/avorty/spito/pkg/vrct/vrctFs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type Configs struct {
	paths           []string
	resultPath      string
	destinationPath string
	configType      int
}

func TestConfigsMatrix(t *testing.T) {
	ruleVrct, err := vrct.NewRuleVRCT()
	if err != nil {
		t.Fatal("Failed to Create VRCT instance")
	}
	fsVrct := &ruleVrct.Fs

	tmpPath, err := os.MkdirTemp("/tmp", "spito-test-")
	if err != nil {
		t.Fatal("Failed to create temporary test directory\n", err)
	}

	configs := []Configs{
		{
			paths:           []string{"yaml/empty-extrepo.yaml", "yaml/full-extrepo.yaml"},
			resultPath:      "yaml/full-extrepo.yaml",
			destinationPath: tmpPath + "/new_dir/extrepo.yaml",
			configType:      vrctFs.YamlConfig,
		},
	}

	for _, config := range configs {
		testConfigs(t, fsVrct, config)
	}

	// cleanup
	_ = os.RemoveAll(tmpPath)
}

func testConfigs(t *testing.T, vrct *vrctFs.FsVRCT, configs Configs) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to obtain working directory: '%s'", wd)
	}
	wd = filepath.Join(wd, "config_data")
	for _, path := range configs.paths {
		workingPath := filepath.Join(wd, path)
		configTestData, err := os.ReadFile(workingPath)
		if err != nil {
			t.Fatalf("Failed to open test data '%s': %s", workingPath, err)
		}

		err = vrct.CreateFile(configs.destinationPath, configTestData, nil, false, configs.configType)
		if err != nil {
			t.Fatal("Failed trying to override file "+configs.destinationPath+"\n", err)
		}
	}
	workingResPath := filepath.Join(wd, configs.resultPath)
	desiredRawResult, err := os.ReadFile(workingResPath)
	if err != nil {
		t.Fatalf("Failed to open result data '%s': %s", workingResPath, err)
	}

	obtainedRawResult, err := vrct.ReadFile(configs.destinationPath)
	if err != nil {
		t.Fatalf("Failed to read file destinationPath %s: %s\n", configs.destinationPath, err)
	}

	err = vrct.Apply()
	if err != nil {
		t.Fatal("Failed to apply VRCT\n", err)
	}

	obtainedRealRawResult, err := os.ReadFile(configs.destinationPath)
	if err != nil {
		t.Fatal("Failed to read from real fs file "+configs.destinationPath+"\n", err)
	}

	desiredResult, err := vrctFs.GetMapFromBytes(desiredRawResult, configs.configType)
	if err != nil {
		t.Fatal(err)
	}

	obtainedResult, err := vrctFs.GetMapFromBytes(obtainedRawResult, configs.configType)
	if err != nil {
		t.Fatal(err)
	}

	obtainedRealResult, err := vrctFs.GetMapFromBytes(obtainedRealRawResult, configs.configType)
	if err != nil {
		t.Fatal(err)
	}

	eq := reflect.DeepEqual(desiredResult, obtainedResult)
	if !eq {
		t.Fatal("Failed to properly simulate " + configs.destinationPath + " file content")
	}

	eq = reflect.DeepEqual(desiredResult, obtainedRealResult)
	if !eq {
		t.Fatal("Failed to properly simulate " + configs.destinationPath + " file content")
	}
}
