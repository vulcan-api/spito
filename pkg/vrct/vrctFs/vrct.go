package vrctFs

import (
	"encoding/json"
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
	"sort"
	"strings"
)

const VirtualFsPathPrefix = "/tmp/spito-vrct/fs"
const VirtualFilePostfix = ".prototype.bson"

type FsVRCT struct {
	virtualFSPath  string
	fsRequirements []FsRequirement
	revertSteps    RevertSteps
}

func NewFsVRCT() (FsVRCT, error) {
	err := os.MkdirAll(VirtualFsPathPrefix, os.ModePerm)
	revertSteps, err := NewRevertSteps()
	if err != nil {
		return FsVRCT{}, nil
	}

	err = os.MkdirAll(VirtualFsPathPrefix, os.ModePerm)
	if err != nil {
		return FsVRCT{}, err
	}

	dir, err := os.MkdirTemp(VirtualFsPathPrefix, "")
	if err != nil {
		return FsVRCT{}, err
	}

	return FsVRCT{
		virtualFSPath:  dir,
		fsRequirements: make([]FsRequirement, 0),
		revertSteps:    revertSteps,
	}, nil
}

func (v *FsVRCT) DeleteRuntimeTemp() error {
	if err := v.revertSteps.DeleteRuntimeTemp(); err != nil {
		return err
	}
	return os.RemoveAll(v.virtualFSPath)
}

func (v *FsVRCT) checkRequirements() (bool, *FsRequirement) {
	for _, req := range v.fsRequirements {
		if req.checkRequirement() == false {
			return false, &req
		}
	}

	return true, nil
}

func (v *FsVRCT) InnerValidate() error {
	for _, requirement := range v.fsRequirements {
		if !requirement.checkRequirement() {
			return fmt.Errorf("requirement %v is not met", requirement.ruleStackTrace)
		}
	}

	return nil
}

func (v *FsVRCT) Apply() error {
	if err := v.InnerValidate(); err != nil {
		return err
	}

	mergeDir, err := os.MkdirTemp("/tmp", "spito-fs-vrct-merge")
	if err != nil {
		return err
	}

	if err := mergePrototypes(v.virtualFSPath, mergeDir); err != nil {
		return err
	}

	if err := v.mergeToRealFs(mergeDir); err != nil {
		return err
	}

	return os.RemoveAll(mergeDir)
}

func (v *FsVRCT) Revert() error {
	return v.revertSteps.Apply()
}

func (v *FsVRCT) mergeToRealFs(mergeDirPath string) error {
	splitMergePath := strings.Split(mergeDirPath, "/")[3:]
	destPath := strings.Join(splitMergePath, "/")
	if len(destPath) != 0 {
		destPath = "/" + destPath
	}

	entries, err := os.ReadDir(mergeDirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		realFsEntryPath := destPath + "/" + entry.Name()
		mergeDirEntryPath := mergeDirPath + "/" + entry.Name()

		if entry.IsDir() {
			_, err := os.Stat(realFsEntryPath)
			if err != nil && !os.IsNotExist(err) {
				return err
			}

			// If originally dir does not exist, then revert should delete it
			if os.IsNotExist(err) {
				v.revertSteps.RemoveDirAll(realFsEntryPath)
			}
			if err := os.MkdirAll(realFsEntryPath, os.ModePerm); err != nil {
				return err
			}
			if err := v.mergeToRealFs(mergeDirEntryPath); err != nil {
				return err
			}
			continue
		}

		filePrototype := FilePrototype{}
		err := filePrototype.Read(v.virtualFSPath, realFsEntryPath)
		if err != nil {
			return err
		}

		if err := v.revertSteps.BackupOldContent(realFsEntryPath); err != nil {
			return err
		}

		err = os.Remove(realFsEntryPath)
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		if err := os.Rename(mergeDirEntryPath, realFsEntryPath); err != nil {
			return err
		}

		if err := filePrototype.Save(); err != nil {
			return err
		}
	}

	return nil
}

func mergePrototypes(prototypesDirPath, destPath string) error {
	dirs, err := os.ReadDir(prototypesDirPath)
	if err != nil {
		return err
	}

	for _, dirEntry := range dirs {
		if dirEntry.IsDir() {
			dirName := dirEntry.Name()
			if err := os.MkdirAll(destPath+"/"+dirName, os.ModePerm); err != nil {
				return err
			}
			if err := mergePrototypes(prototypesDirPath+"/"+dirName, destPath+"/"+dirName); err != nil {
				return err
			}
			continue
		}
		prototypeName := dirEntry.Name()
		if strings.Contains(prototypeName, ".prototype.bson") {
			fileName := strings.ReplaceAll(prototypeName, ".prototype.bson", "")

			prototype := FilePrototype{}
			if err := prototype.Read(prototypesDirPath+"/", fileName); err != nil {
				return err
			}
			file, err := prototype.SimulateFile()
			if err != nil {
				return err
			}

			if err := os.WriteFile(destPath+"/"+fileName, file, os.ModePerm); err != nil {
				return err
			}
			continue
		}
	}

	return nil
}

type FsRequirement struct {
	// In simpler words - how this rule appeared here
	ruleStackTrace   []string
	checkRequirement func() bool
}

type PrototypeLayer struct {
	// If ContentPath is specified and file exists in real fs, real file will be later overridden by this content
	// (We don't store content as string in order to make bson lightweight and fast accessible)
	ContentPath string `bson:",omitempty"`
	OptionsPath string `bson:",omitempty"`
	IsOptional  bool
}

func (p *FilePrototype) mergeLayers() (PrototypeLayer, error) {
	switch p.FileType {
	case TextFile:
		return p.mergeTextLayers()
	default:
		return p.mergeConfigLayers()
	}
}

func (p *FilePrototype) mergeTextLayers() (PrototypeLayer, error) {
	finalLayer := PrototypeLayer{
		IsOptional: false,
	}

	sort.Slice(p.Layers, func(i, j int) bool {
		return !p.Layers[i].IsOptional && p.Layers[j].IsOptional
	})

	for i, layer := range p.Layers {
		// TODO: check if contents differ
		if finalLayer.ContentPath != "" && layer.ContentPath != "" && !layer.IsOptional {
			return finalLayer, fmt.Errorf(
				"non optional layer nr. %d is in conflict with previous layers", i,
			)
		}
		if (layer.IsOptional && finalLayer.ContentPath == "") || !layer.IsOptional {
			finalLayer.ContentPath = layer.ContentPath
		}
	}

	return finalLayer, nil
}

func (p *FilePrototype) mergeConfigLayers() (PrototypeLayer, error) {
	// create new layer first
	finalLayer, err := p.CreateLayer(nil, nil, true)
	if err != nil {
		return finalLayer, err
	}

	finalLayerContent, err := GetBsonMap(finalLayer.ContentPath)
	if err != nil {
		return finalLayer, err
	}

	finalLayerOptions, err := GetBsonMap(finalLayer.OptionsPath)
	if err != nil {
		return finalLayer, err
	}

	sort.Slice(p.Layers, func(i, j int) bool {
		return !p.Layers[i].IsOptional && p.Layers[j].IsOptional
	})

	for _, layer := range p.Layers {
		currentLayerContent, err := GetBsonMap(layer.ContentPath)
		if err != nil {
			return finalLayer, err
		}

		currentLayerOptions, err := GetBsonMap(layer.OptionsPath)
		if err != nil {
			return finalLayer, err
		}

		finalLayerContent, finalLayerOptions, err = mergeConfigs(finalLayerContent, finalLayerOptions, currentLayerContent, currentLayerOptions, layer.IsOptional)
		if err != nil {
			return finalLayer, err
		}
	}

	marshalledContent, err := bson.Marshal(finalLayerContent)
	if err != nil {
		return finalLayer, err
	}
	err = finalLayer.SetContent(marshalledContent)

	err = SaveBsonMap(finalLayerOptions, finalLayer.OptionsPath)

	return finalLayer, err
}

func getBoolMap(value interface{}, option interface{}) (map[string]interface{}, map[string]interface{}, error) {
	valueKind := reflect.ValueOf(value).Kind()
	optionKind := reflect.ValueOf(option).Kind()

	if valueKind != reflect.Map {
		return nil, nil, fmt.Errorf("trying to map interface that is not map")
	}

	mappedValue := value.(map[string]interface{})

	var mappedOption map[string]interface{}
	if optionKind == reflect.Map {
		mappedOption = option.(map[string]interface{})
	} else if option == reflect.Bool {
		for key := range mappedValue {
			mappedOption[key] = option.(bool)
		}
	} else {
		return nil, nil, fmt.Errorf("types conflict")
	}

	return mappedValue, mappedOption, nil
}

//func getBoolArray(value interface{}, option interface{}) ([]any, []any, error) {
//	valueKind := reflect.ValueOf(value).Kind()
//	optionKind := reflect.ValueOf(option).Kind()
//
//	if valueKind != reflect.Slice {
//		return nil, nil, fmt.Errorf("trying to map interface that is not map")
//	}
//	arrayedValue := value.([]any)
//
//	var arrayedOption []any
//	if optionKind == reflect.Slice {
//		arrayedOption = option.([]any)
//	} else if optionKind == reflect.Bool {
//		for range arrayedValue {
//			arrayedOption = append(arrayedOption, option.(bool))
//		}
//	} else {
//		return arrayedValue, nil, fmt.Errorf("options structure does not match config's one: types conflict '%s' and '%s'", valueKind, optionKind)
//	}
//
//	return arrayedValue, arrayedOption, nil
//}

func mergeConfigs(merger map[string]interface{}, mergerOptions map[string]interface{}, toMerge map[string]interface{}, toMergeOptions map[string]interface{}, isOptional bool) (map[string]interface{}, map[string]interface{}, error) {
	for key, toMergeVal := range toMerge {
		mergerVal, ok := merger[key]

		mergerOption, mergerOptOk := mergerOptions[key]
		toMergeOption, toMergeOptOk := toMergeOptions[key]

		// TODO: check for structure conflicts
		if !ok {
			merger[key] = toMergeVal
			if toMergeOptOk {
				mergerOptions[key] = toMergeOption
			} else {
				mergerOptions[key] = isOptional
			}
			continue
		}

		if reflect.ValueOf(toMergeVal).Kind() == reflect.Map {
			mergerMapVal, mergerOptsMapVal, err := getBoolMap(mergerVal, mergerOption)
			if err != nil {
				return merger, mergerOptions, err
			}

			toMergeMapVal, toMergeOptsMapVal, err := getBoolMap(mergerVal, mergerOption)
			if err != nil {
				return merger, mergerOptions, err
			}

			merger[key], mergerOptions[key], err = mergeConfigs(toMergeMapVal, mergerOptsMapVal, mergerMapVal, toMergeOptsMapVal, isOptional)
			if err != nil {
				return merger, mergerOptions, err
			}
			continue
		}

		// TODO: good system needed
		//if reflect.ValueOf(toMergeVal).Kind() == reflect.Slice {
		//	mergerValueAsArray, mergerOptionAsArray, err := getBoolArray(mergerVal, mergerOption)
		//	if err != nil {
		//		return merger, mergerOptions, err
		//	}

		//	toMergeValueAsArray, toMergeOptionAsArray, err := getBoolArray(mergerVal, mergerOption)
		//	if err != nil {
		//		return merger, mergerOptions, err
		//	}

		//	merger[key] = append(mergerValueAsArray, toMergeValueAsArray...)
		//	mergerOptions[key] = append(mergerOptionAsArray, toMergeOptionAsArray...)

		//	if err != nil {
		//		return merger, mergerOptions, err
		//	}

		//	continue
		//}

		isMergerKeyOpt := true
		mergerOptKind := reflect.ValueOf(mergerOption).Kind()
		if mergerOptOk {
			if mergerOptKind != reflect.Bool {
				return merger, mergerOptions, fmt.Errorf("passed key '%s' in merger options is not of type bool ('%s' passed)", key, mergerOptKind)
			}
			isMergerKeyOpt = mergerOption.(bool)
		}

		isToMergeKeyOpt := isOptional
		toMergeOptKind := reflect.ValueOf(toMergeOption).Kind()
		if toMergeOptOk {
			if toMergeOptKind != reflect.Bool {
				return merger, mergerOptions, fmt.Errorf("passed key '%s' in to merge options is not of type bool ('%s' passed)", key, toMergeOptKind)
			}
			isToMergeKeyOpt = toMergeOption.(bool)
		}

		if isMergerKeyOpt && isToMergeKeyOpt {
			merger[key] = toMergeVal
			mergerOptions[key] = true
		} else if isMergerKeyOpt {
			merger[key] = toMergeVal
			mergerOptions[key] = false
		} else if isToMergeKeyOpt {
			mergerOptions[key] = false
		} else if mergerVal != toMergeVal {
			return merger, toMergeOptions, fmt.Errorf("passed key '%s' is unmergable (both merger and to merge are required)", key)
		}
	}

	return merger, mergerOptions, nil
}

func GetBsonMap(pathToFile string) (map[string]interface{}, error) {
	file, err := os.ReadFile(pathToFile)
	if err != nil {
		return nil, err
	}

	var bsonMap map[string]interface{}
	err = bson.Unmarshal(file, &bsonMap)
	if err != nil {
		return bsonMap, err
	}

	return bsonMap, nil
}

func SaveBsonMap(toSave map[string]interface{}, pathToFile string) error {
	content, err := bson.Marshal(toSave)
	if err != nil {
		return err
	}

	err = os.WriteFile(pathToFile, content, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func GetMapFromBytes(content []byte, configType int) (map[string]interface{}, error) {
	var err error
	var resultMap map[string]interface{}
	switch configType {
	case TextFile:
		break
	case JsonConfig:
		if content == nil {
			content = []byte("{}")
		}
		err = json.Unmarshal(content, &resultMap)
		break
	case YamlConfig:
		err = yaml.Unmarshal(content, &resultMap)
		break
	case TomlConfig:
		err = toml.Unmarshal(content, &resultMap)
		break
	default:
		return resultMap, fmt.Errorf("unsupported config type (FileType argument), passed '%d'", configType)
	}

	if err != nil {
		return resultMap, fmt.Errorf("could not obtain map from given array of bytes: %s", err)
	}

	if len(resultMap) == 0 {
		resultMap = make(map[string]interface{})
	}

	return resultMap, err
}
