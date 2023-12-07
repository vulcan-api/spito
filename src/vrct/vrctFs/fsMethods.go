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
		return nil, err
	}

	if len(filePrototype.Layers) == 0 {
		return nil, fmt.Errorf("file %s not found", filePath)
	}

	return filePrototype.SimulateFile()
}
