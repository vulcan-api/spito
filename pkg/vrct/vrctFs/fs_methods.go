package vrctFs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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

func (v *FsVRCT) Stat(path string) (os.FileInfo, error) {
	path, err := pathMustBeAbsolute(path)
	if err != nil {
		return nil, err
	}

	splitPath := strings.Split(path, "/")
	name := splitPath[len(splitPath)-1]

	prototypePath := fmt.Sprintf("%s%s.prototype.bson", v.virtualFSPath, path)

	stat, err := os.Stat(prototypePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		filePrototype := FilePrototype{}
		if err := filePrototype.Read(prototypePath); err != nil {
			return nil, err
		}
		content, err := filePrototype.SimulateFile()
		if err != nil {
			return nil, err
		}

		return FileInfo{
			name:     name,
			size:     int64(len(content)),
			fileMode: stat.Mode(),    //TODO: consider changing it
			modTime:  stat.ModTime(), //TODO: same
			isDir:    stat.IsDir(),
		}, nil
	}

	fileStat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	return FileInfo{
		name:     name,
		size:     fileStat.Size(),    //TODO: consider changing it
		fileMode: fileStat.Mode(),    //TODO: same
		modTime:  fileStat.ModTime(), //TODO: same
		isDir:    fileStat.IsDir(),
	}, nil
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
				name:  name,
				isDir: entry.IsDir(),
				type_: entry.Type(), // TODO: consider changing it
				StatFn: func() (fs.FileInfo, error) {
					return v.Stat(path)
				},
			}
		}
	}

	res := make([]os.DirEntry, 0, len(dirEntries))

	for _, entry := range dirEntries {
		res = append(res, entry)
	}

	return res, nil
}
