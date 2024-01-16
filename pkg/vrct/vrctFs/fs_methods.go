package vrctFs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// CreateFile function creating file
//
// Arguments:
//
//	filePath - path to file
//	content - content of file
//	optionalKeys - json or yaml document describing which key in config is optional
//	isOptional - default option in configs / is able to merge in text files
//	fileType - given 0 - text file, otherwise config specified in file_type.go
func (v *VRCTFs) CreateFile(filePath string, content []byte, optionalKeys []byte, isOptional bool, fileType int) error {
	filePath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}
	dirPath := filepath.Dir(filePath)

	err = os.MkdirAll(filepath.Join(v.virtualFSPath, dirPath), os.ModePerm)
	if err != nil {
		return err
	}

	filePrototype := FilePrototype{
		FileType: fileType,
	}
	err = filePrototype.Read(v.virtualFSPath, filePath)
	if err != nil {
		return err
	}

	prototypeLayer, err := filePrototype.CreateLayer(content, optionalKeys, isOptional)
	if err != nil {
		return err
	}

	err = filePrototype.AddNewLayer(prototypeLayer)
	return err
}

func (v *VRCTFs) ReadFile(filePath string) ([]byte, error) {
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

func (v *VRCTFs) Stat(path string) (os.FileInfo, error) {
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
			name:    name,
			size:    int64(len(content)),
			mode:    stat.Mode(),
			modTime: stat.ModTime(),
			isDir:   stat.IsDir(),
			sys:     stat.Sys(),
		}, nil
	}

	fileStat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	return FileInfo{
		name:    name,
		size:    fileStat.Size(),
		mode:    fileStat.Mode(),
		modTime: fileStat.ModTime(),
		isDir:   fileStat.IsDir(),
		sys:     stat.Sys(),
	}, nil
}

func (v *VRCTFs) ReadDir(path string) ([]os.DirEntry, error) {
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
				name:      name,
				isDir:     entry.IsDir(),
				entryType: entry.Type(),
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
