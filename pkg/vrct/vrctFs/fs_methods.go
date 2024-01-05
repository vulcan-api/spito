package vrctFs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// CreateFile TODO: check for conflict out of the box
func (v *FsVRCT) CreateFile(filePath string, content []byte, isOptional bool, fileType int) error {
	// TODO: allow non-absolute paths
	filePath, err := pathMustBeAbsolute(filePath)
	if err != nil {
		return err
	}
	dirPath := filepath.Dir(filePath)

	// TODO: consider pushing almost everything below in filePrototype.LoadOrCreate
	err = os.MkdirAll(fmt.Sprintf("%s%s", v.virtualFSPath, dirPath), os.ModePerm)
	if err != nil {
		return err
	}

	// TODO: create function that allows to merge json and xml configs
	filePrototype := FilePrototype{
		FileType: fileType,
	}
	err = filePrototype.Read(v.virtualFSPath, filePath)
	if err != nil {
		return err
	}

	prototypeLayer, err := filePrototype.CreateLayer(content, isOptional)
	if err != nil {
		return err
	}

	err = filePrototype.AddNewLayer(prototypeLayer)
	return err
}

func (v *FsVRCT) ReadFile(filePath string) ([]byte, error) {
	filePath, err := pathMustBeAbsolute(filePath)
	if err != nil {
		return nil, err
	}

	filePrototype := FilePrototype{}
	err = filePrototype.Read(v.virtualFSPath, filePath)
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
		if err := filePrototype.Read(v.virtualFSPath, path); err != nil {
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
