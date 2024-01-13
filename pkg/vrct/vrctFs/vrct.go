package vrctFs

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

const VirtualFsPathPrefix = "/tmp/spito-vrct/fs"
const VirtualFilePostfix = ".prototype.bson"

type FsVRCT struct {
	virtualFSPath  string
	fsRequirements []FsRequirement
}

type FsRequirement struct {
	// In simpler words - how this rule appeared here
	ruleStackTrace   []string
	required         map[string]string
	abandoned        map[string]string
	checkRequirement func() bool
}

type FilePrototype struct {
	Layers         []PrototypeLayer
	RealFileExists bool
	FileType       int
	Path           string `bson:"-"`
	Name           string `bson:"-"`
}

type PrototypeLayer struct {
	// If content path is specified and file exists in real fs, real file will be later overridden by this content
	// (We don't store content as string in order to make bson lightweight and fast accessible)
	ContentPath string `bson:",omitempty"`
	OptionsPath string `bson:",omitempty"`
	IsOptional  bool
	// TODO: stack trace
}

func NewFsVRCT() (FsVRCT, error) {
	err := os.MkdirAll(VirtualFsPathPrefix, os.ModePerm)
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
	}, nil
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

	return mergeToRealFs(mergeDir)
}

func mergeToRealFs(mergeDirPath string) error {
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
		if entry.IsDir() {
			if err := os.MkdirAll(destPath+"/"+entry.Name(), os.ModePerm); err != nil {
				return err
			}
			if err := mergeToRealFs(mergeDirPath + "/" + entry.Name()); err != nil {
				return err
			}
			continue
		}

		err := os.Remove(destPath + "/" + entry.Name())
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		if err := os.Rename(mergeDirPath+"/"+entry.Name(), destPath+"/"+entry.Name()); err != nil {
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

func mergeConfigs(merger map[string]interface{}, mergerOptions map[string]interface{}, toMerge map[string]interface{}, toMergeOptions map[string]interface{}, isOptional bool) (map[string]interface{}, map[string]interface{}, error) {
	var err error
	// TODO: add merging arrays
	for key, toMergeVal := range toMerge {
		mergerVal, ok := merger[key]

		mergerOption, mergerOptOk := mergerOptions[key]
		toMergeOption, toMergeOptOk := toMergeOptions[key]

		// TODO: check for structure conflicts
		if !ok {
			merger[key] = toMergeVal
			if mergerOptions == nil {
				mergerOptions = make(map[string]interface{})
			}
			if toMergeOptOk {
				mergerOptions[key] = toMergeOption
			} else {
				mergerOptions[key] = isOptional
			}
			continue
		}

		if reflect.ValueOf(toMergeVal).Kind() == reflect.Map {
			if mergerValKind := reflect.ValueOf(mergerVal).Kind(); mergerValKind != reflect.Map {
				return merger, mergerOptions, fmt.Errorf("incompatibile structures: merger is not a map (it's '%s') while to merge is", mergerValKind)
			}

			var mergerOptsMapVal map[string]interface{}
			if mergerOptOk {
				if mergerOptsValKind := reflect.ValueOf(mergerOption).Kind(); mergerOptsValKind != reflect.Map {
					return merger, mergerOptions,
						fmt.Errorf("incompatible structures: options value behind key '%s' is not a map (it's '%s') while to merge's is", key, mergerOptsValKind)
				}

				mergerOptsMapVal = mergerOptions[key].(map[string]interface{})
			}

			var toMergeOptsMapVal map[string]interface{}
			if toMergeOptOk {
				if toMergeOptsValKind := reflect.ValueOf(toMergeOption).Kind(); toMergeOptOk && toMergeOptsValKind != reflect.Map {
					return merger, mergerOptions,
						fmt.Errorf("incompatible structures: options value behind key '%s' is not a map (it's '%s') while to merge's is", key, toMergeOptsValKind)
				}

				toMergeOptsMapVal = toMergeOptions[key].(map[string]interface{})
			}
			mergerMapVal := mergerVal.(map[string]interface{})
			toMergeMapVal := toMergeVal.(map[string]interface{})

			merger[key], mergerOptions[key], err = mergeConfigs(toMergeMapVal, mergerOptsMapVal, mergerMapVal, toMergeOptsMapVal, isOptional)
			if err != nil {
				return merger, mergerOptions, err
			}
			continue
		}

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

		if mergerOptions == nil {
			mergerOptions = make(map[string]interface{})
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

// TODO: fix this bro (it does nothing) and rename it to getRealPath
func (p *FilePrototype) getDestinationPath() string {
	splitPath := strings.Split(p.getVirtualPath(), "/")
	return strings.Join(splitPath[:len(splitPath)-1], "/")
}

func (p *FilePrototype) getVirtualPath() string {
	return p.Path + p.Name + VirtualFilePostfix
}

func (p *FilePrototype) SimulateFile() ([]byte, error) {
	finalLayer, err := p.mergeLayers()

	if err != nil {
		return nil, err
	}

	var filePath string

	if p.RealFileExists {
		filePath = p.getDestinationPath()
	} else {
		filePath = finalLayer.ContentPath
	}

	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var tempContentInterface map[string]interface{}
	if p.FileType != TextFile {
		err = bson.Unmarshal(file, &tempContentInterface)
		if err != nil {
			return file, err
		}
	}

	var fileContent []byte
	switch p.FileType {
	case JsonConfig:
		fileContent, err = json.Marshal(tempContentInterface)
		break
	case YamlConfig:
		fileContent, err = yaml.Marshal(tempContentInterface)
		break
	case TomlConfig:
		fileContent, err = toml.Marshal(tempContentInterface)
		break
	default:
		return file, nil
	}
	return fileContent, err
}

// TODO: think of splitting it up into to functions (read and load)
func (p *FilePrototype) Read(vrctPrefix string, realPath string) error {
	prototypeFilePath := vrctPrefix + realPath

	// TODO: instead of adding slash, use filepath.Join in appropriate lines
	path := filepath.Dir(prototypeFilePath)
	path += "/"
	name := filepath.Base(prototypeFilePath)

	p.Path = path
	p.Name = name
	file, err := os.ReadFile(p.getVirtualPath())

	if os.IsNotExist(err) {
		return p.Save()
	} else if err != nil {
		return err
	}

	err = bson.Unmarshal(file, p)

	// TODO: think about it (sus)
	p.Path = path
	p.Name = name
	return err
}

func (p *FilePrototype) Save() error {
	rawBson, err := bson.Marshal(p)
	if err != nil {
		return err
	}

	return os.WriteFile(p.getVirtualPath(), rawBson, os.ModePerm)
}

func (p *FilePrototype) CreateLayer(content []byte, options []byte, isOptional bool) (PrototypeLayer, error) {
	if p.Path == "" {
		return PrototypeLayer{}, errors.New("file prototype hasn't been loaded yet")
	}

	// TODO: check if file with random name already exist
	dir := filepath.Dir(p.Path)

	randFileName := randomLetters(5)
	contentPath := filepath.Join(dir, randFileName)

	randOptsName := randomLetters(5)
	optionsPath := filepath.Join(dir, randOptsName)

	tempConvertedContent, err := GetMapFromBytes(content, p.FileType)

	if p.FileType != TextFile {
		content, err = bson.Marshal(tempConvertedContent)
		if err != nil {
			return PrototypeLayer{}, err
		}
	}

	// TODO: think about changing ModePerm (spito can be run as root (dangerous))
	if err = os.WriteFile(contentPath, content, os.ModePerm); err != nil {
		return PrototypeLayer{}, err
	}

	if options == nil {
		options = []byte("{}")
	}

	var tempOptionalKeysMap map[string]interface{}
	err = json.Unmarshal(options, &tempOptionalKeysMap)
	if err != nil {
		return PrototypeLayer{}, err
	}

	optionalKeysBson, err := bson.Marshal(tempOptionalKeysMap)
	if err != nil {
		return PrototypeLayer{}, err
	}

	if err = os.WriteFile(optionsPath, optionalKeysBson, os.ModePerm); err != nil {
		return PrototypeLayer{}, err
	}

	newLayer := PrototypeLayer{
		ContentPath: contentPath,
		OptionsPath: optionsPath,
		IsOptional:  isOptional,
	}

	return newLayer, nil
}

func (p *FilePrototype) AddNewLayer(layer PrototypeLayer) error {
	// TODO: merge first and check if new layers is addable
	p.Layers = append(p.Layers, layer)
	if err := p.Save(); err != nil {
		return err
	}

	return nil
}

func (layer *PrototypeLayer) GetContent() ([]byte, error) {
	file, err := os.ReadFile(layer.ContentPath)
	if err != nil {
		return file, err
	}

	return file, nil
}

func (layer *PrototypeLayer) SetContent(content []byte) error {
	err := os.WriteFile(layer.ContentPath, content, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
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
