package vrctFs

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"os"
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
		//fmt.Printf("[WARNING] Unrecognized file %s in %s\n", prototypeName, prototypesDirPath)
	}

	return nil
}

func (p *FilePrototype) mergeLayers() (PrototypeLayer, error) {
	finalLayer := PrototypeLayer{
		IsOptional: false,
	}
	var err error

	// Non-optional layers first
	sort.Slice(p.Layers, func(i, j int) bool {
		return !p.Layers[i].IsOptional && p.Layers[j].IsOptional
	})

	for i, layer := range p.Layers {
		if p.FileType == TextFile {
			// TODO: check if contents differ
			if finalLayer.ContentPath != "" && layer.ContentPath != "" && !layer.IsOptional {
				return finalLayer, fmt.Errorf(
					"non optional layer nr. %d is in conflict with previous layers", i,
				)
			}
			if (layer.IsOptional && finalLayer.ContentPath == "") || !layer.IsOptional {
				finalLayer.ContentPath = layer.ContentPath
			}
		} else {
			// TODO: convert everything to bson and then compare

			//finalLayer.ContentPath
			//switch p.FileType {
			//case JsonConfig:
			//	config.MergeJson(finalLayer, layer)
			//	break
			//default:
			//	return finalLayer, fmt.Errorf("unknown file type '%d' of prototype file '%s'", p.FileType, p.SelfPath)
			//}
		}
	}

	return finalLayer, err
}

// TODO: fix this bro (it does nothing)
func (p *FilePrototype) getDestinationPath() string {
	splitPath := strings.Split(p.getVirtualPath(), "/")
	return strings.Join(splitPath[:len(splitPath)-1], "/")
}

// TODO: create this
func (p *FilePrototype) getRealPath() string {
	return ""
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

	return file, nil
}

func (p *FilePrototype) Read(vrctPrefix string, realPath string) error {
	prototypeFilePath := vrctPrefix + realPath

	pathEnd := strings.LastIndex(prototypeFilePath, "/") + 1
	path := prototypeFilePath[0:pathEnd]
	name := prototypeFilePath[pathEnd:]

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

func (p *FilePrototype) AddNewLayer(layer PrototypeLayer) error {
	// TODO: merge first and check if new layers is addable
	p.Layers = append(p.Layers, layer)
	if err := p.Save(); err != nil {
		return err
	}

	return nil
}

// TODO: create function creating new layer :>

func (layer *PrototypeLayer) GetContent() ([]byte, error) {
	file, err := os.ReadFile(layer.ContentPath)
	if err != nil {
		return file, err
	}

	return file, nil
}
