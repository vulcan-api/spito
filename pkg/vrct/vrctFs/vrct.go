package vrctFs

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
)

const VirtualFsPathPrefix = "/tmp/spito-vrct/fs"

type FsVRCT struct {
	virtualFSPath  string
	fsRequirements []FsRequirement
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

func (v *FsVRCT) DeleteRuntimeTemp() error {
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

	revertSteps, err := NewRevertSteps()
	if err != nil {
		return err
	}

	if err := v.mergeToRealFs(mergeDir, &revertSteps); err != nil {
		return err
	}

	return os.RemoveAll(mergeDir)
}

func (v *FsVRCT) Revert() error {
	// TODO: finish this one
	return nil
}

func (v *FsVRCT) mergeToRealFs(mergeDirPath string, revertSteps *RevertSteps) error {
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
		prototypePath := fmt.Sprintf("%s/%s.prototype.bson", v.virtualFSPath, entry.Name())

		if entry.IsDir() {
			_, err := os.Stat(realFsEntryPath)
			if err != nil && !os.IsNotExist(err) {
				return err
			}

			// If originally dir does not exist, then revert should delete it
			if os.IsNotExist(err) {
				revertSteps.RemoveDirAll(realFsEntryPath)
			}
			if err := os.MkdirAll(realFsEntryPath, os.ModePerm); err != nil {
				return err
			}
			if err := v.mergeToRealFs(mergeDirEntryPath, revertSteps); err != nil {
				return err
			}
			continue
		}

		filePrototype := FilePrototype{}
		err := filePrototype.Read(prototypePath, realFsEntryPath)
		if err != nil {
			return err
		}

		// TODO: add revert step here
		if err := revertSteps.BackupOldContent(realFsEntryPath); err != nil {
			return err
		}

		err = os.Remove(realFsEntryPath)
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		if err := os.Rename(mergeDirEntryPath, realFsEntryPath); err != nil {
			return err
		}

		filePrototype.IsApplied = true
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
			if err := prototype.Read(prototypesDirPath+"/"+prototypeName, destPath); err != nil {
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

type PrototypeLayer struct {
	// If ContentPath is specified and file exists in real fs, real file will be later overridden by this content
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
