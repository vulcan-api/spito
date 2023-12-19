package vrctFs

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"os"
	"slices"
	"sort"
	"strings"
)

const virtualFSPathPrefix = "/tmp/spito-vrct/fs"

type FsVRCT struct {
	virtualFSPath  string
	fsRequirements []FsRequirement
}

func NewFsVRCT() (FsVRCT, error) {
	err := os.MkdirAll(virtualFSPathPrefix, os.ModePerm)
	if err != nil {
		return FsVRCT{}, err
	}

	dir, err := os.MkdirTemp(virtualFSPathPrefix, "")
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

// TODO: write test for following function
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

// TODO: check if it works
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
			if err := prototype.Read(prototypesDirPath + "/" + prototypeName); err != nil {
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
		fmt.Printf("[WARNING] Unrecognized file %s in %s\n", prototypeName, prototypesDirPath)
	}

	return nil
}

type FsRequirement struct {
	// In simpler words - how this rule appeared here
	ruleStackTrace   []string
	checkRequirement func() bool
}

const (
	ConfigFile = iota
	TextFile
)

type FilePrototype struct {
	Layers         []PrototypeLayer
	RealFileExists bool
	FileType       *int
	SelfPath       string `bson:"-"`
}

type PrototypeLayer struct {
	// If content path is specified and file exists in real fs, real file will be later overridden by this content
	// (We don't store content as string in order to make bson lightweight and fast accessible)
	ContentPath   *string `bson:",omitempty"`
	IsOptional    bool
	ConfigOptions []ConfigOption `bson:",omitempty"`
	ConfigType    *int           `bson:",omitempty"`
	// TODO: stack trace
}

func (p *FilePrototype) mergeLayers(optLayersIndexesToSkip []int) (PrototypeLayer, error) {
	pl := PrototypeLayer{
		IsOptional: false,
	}

	encounteredOptionalLayers := 0

	for i, layer := range p.Layers {
		if layer.IsOptional {
			encounteredOptionalLayers++
			if slices.Contains(optLayersIndexesToSkip, encounteredOptionalLayers-1) {
				continue
			}
		}

		var optionalInfoPrefix string
		if layer.IsOptional {
			optionalInfoPrefix = "optional"
		} else {
			optionalInfoPrefix = "non optional"
		}
		potentialConflictError := fmt.Errorf(
			"%s layer nr. %d is in conflict with previous layers", optionalInfoPrefix, i,
		)

		pl.ConfigOptions = append(pl.ConfigOptions, layer.ConfigOptions...)
		pl.IsOptional = layer.IsOptional && pl.IsOptional

		if pl.ConfigType != nil && layer.ConfigType != nil {
			return pl, potentialConflictError
		}
		pl.ConfigType = layer.ConfigType
	}

	// Not optional layers first
	sort.Slice(p.Layers, func(i, j int) bool {
		return !p.Layers[i].IsOptional && p.Layers[j].IsOptional
	})

	for i, layer := range p.Layers {
		if pl.ContentPath != nil && layer.ContentPath != nil && !layer.IsOptional {
			return pl, fmt.Errorf(
				"non optional layer nr. %d is in conflict with previous layers", i,
			)
		}
		if (layer.IsOptional && pl.ContentPath == nil) || !layer.IsOptional {
			pl.ContentPath = layer.ContentPath
		}
	}

	return pl, nil
}

func (p *FilePrototype) getDestinationPath() string {
	splitPath := strings.Split(p.SelfPath, "/")
	return strings.Join(splitPath[:len(splitPath)-1], "/")
}

func (p *FilePrototype) SimulateFile() ([]byte, error) {
	finalLayer, err := p.mergeLayers([]int{})
	if err != nil {
		return nil, err
	}

	var fileType int
	if p.FileType == nil {
		fileType = TextFile
	} else {
		fileType = *p.FileType
	}

	switch fileType {
	case TextFile:
		var filePath string

		if p.RealFileExists {
			filePath = p.getDestinationPath()
		} else {
			filePath = *finalLayer.ContentPath
		}

		file, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		return file, nil

	case ConfigFile:
		// TODO
		break
	}

	// Todo: handle other FileTypes
	panic("Unimplemented FileType simulation")
}

func (p *FilePrototype) Read(prototypePath string) error {
	p.SelfPath = prototypePath
	file, err := os.ReadFile(prototypePath)

	if os.IsNotExist(err) {
		return p.Save()
	} else if err != nil {
		return err
	}

	err = bson.Unmarshal(file, p)

	p.SelfPath = prototypePath
	return err
}

func (p *FilePrototype) Save() error {
	rawBson, err := bson.Marshal(p)
	if err != nil {
		return err
	}

	return os.WriteFile(p.SelfPath, rawBson, os.ModePerm)
}

func (p *FilePrototype) AddNewLayer(layer PrototypeLayer) error {
	p.Layers = append(p.Layers, layer)
	if err := p.Save(); err != nil {
		return err
	}

	return nil
}
