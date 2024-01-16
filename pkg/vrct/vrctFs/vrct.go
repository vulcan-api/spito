package vrctFs

import (
	"os"
	"strings"
)

const VirtualFsPathPrefix = "/tmp/spito-vrct/fs"
const VirtualFilePostfix = ".prototype.bson"

type VRCTFs struct {
	virtualFSPath string
	revertSteps   RevertSteps
}

func NewFsVRCT() (VRCTFs, error) {
	err := os.MkdirAll(VirtualFsPathPrefix, os.ModePerm)
	revertSteps, err := NewRevertSteps()
	if err != nil {
		return VRCTFs{}, nil
	}

	err = os.MkdirAll(VirtualFsPathPrefix, os.ModePerm)
	if err != nil {
		return VRCTFs{}, err
	}

	dir, err := os.MkdirTemp(VirtualFsPathPrefix, "")
	if err != nil {
		return VRCTFs{}, err
	}

	return VRCTFs{
		virtualFSPath: dir,
		revertSteps:   revertSteps,
	}, nil
}

func (v *VRCTFs) DeleteRuntimeTemp() error {
	if err := v.revertSteps.DeleteRuntimeTemp(); err != nil {
		return err
	}
	return os.RemoveAll(v.virtualFSPath)
}

func (v *VRCTFs) Apply() error {
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

func (v *VRCTFs) Revert() error {
	return v.revertSteps.Apply()
}

func (v *VRCTFs) mergeToRealFs(mergeDirPath string) error {
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
