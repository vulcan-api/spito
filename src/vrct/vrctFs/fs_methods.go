package vrctFs

import (
	"fmt"
	"os"
	"path/filepath"
)

func (v *FsVRCT) CreateFile(filePath string, content []byte, isOptional bool) error {
	filePath, err := pathMustBeAbsolute(filePath)
	if err != nil {
		return err
	}
	dirPath := filepath.Dir(filePath)
	prototypeFilePath := fmt.Sprintf("%s%s.prototype.bson", v.virtualFSPath, filePath)
	randFileName := randomLetters(5)

	var contentPath *string = nil
	if content != nil {
		newContentPath := fmt.Sprintf("%s%s/%s", v.virtualFSPath, dirPath, randFileName)
		contentPath = &newContentPath

		err := os.MkdirAll(fmt.Sprintf("%s%s", v.virtualFSPath, dirPath), os.ModePerm)
		if err != nil {
			return err
		}

		if err := os.WriteFile(*contentPath, content, os.ModePerm); err != nil {
			return err
		}
	}

	filePrototype := FilePrototype{}
	err = filePrototype.Read(prototypeFilePath)
	if err != nil {
		return err
	}

	newLayer := PrototypeLayer{
		ContentPath:   contentPath,
		IsOptional:    isOptional,
		ConfigOptions: nil,
		ConfigType:    nil,
	}

	return filePrototype.AddNewLayer(newLayer)
}

func (v *FsVRCT) ReadFile(filePath string) ([]byte, error) {
	filePath, err := pathMustBeAbsolute(filePath)
	if err != nil {
		return nil, err
	}

	filePrototype := FilePrototype{}
	err = filePrototype.Read(fmt.Sprintf("%s%s.prototype.bson", v.virtualFSPath, filePath))
	if err != nil {
		file, err := os.ReadFile(filePath)
		if err != nil {
			return nil, os.ErrNotExist
		}

		return file, nil
	}

	if len(filePrototype.Layers) == 0 {
		return nil, os.ErrNotExist
	}

	return filePrototype.SimulateFile()
}

func (v *FsVRCT) ReadDir(path string) ([]os.DirEntry, error) {
	dirEntries := make(map[string]os.DirEntry)
	path, err := pathMustBeAbsolute(path)
	if err != nil {
		return nil, err
	}

	realFsEntries, err := os.ReadDir(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		for _, entry := range realFsEntries {
			dirEntries[entry.Name()] = entry
		}
	}

	vrctEntries, err := os.ReadDir(fmt.Sprintf("%s%s", v.virtualFSPath, path))
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		for _, entry := range vrctEntries {
			name := entry.Name()
			if name[len(name)-15:] != ".prototype.bson" && !entry.IsDir() {
				continue
			}
			if !entry.IsDir() {
				name = name[:len(name)-15]
			}

			dirEntries[name] = DirEntry{
				name:     name,
				isDir:    entry.IsDir(),
				type_:    entry.Type(), // TODO: consider changing it
				infoErr:  nil,          // TODO
				fileInfo: nil,          // TODO
			}
		}
	}

	res := make([]os.DirEntry, 0, len(dirEntries))

	for _, entry := range dirEntries {
		res = append(res, entry)
	}

	return res, nil
}
