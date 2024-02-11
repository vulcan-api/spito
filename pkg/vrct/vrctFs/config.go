package vrctFs

import (
	"errors"
	"os"
	"path/filepath"
)

// CreateConfig function creating or updating configuration file
//
// Arguments:
//
//	filePath - Path to file
//	content - content of file
//	optionalKeys - json or yaml document describing which key in config is optional
//	isOptional - default option in configs / is able to merge in text files
//	fileType - given 0 - text file, otherwise config specified in file_type.go
func (v *VRCTFs) CreateConfig(filePath string, content []byte, optionalKeys []byte, isOptional bool, fileType FileType) error {
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

	if filePrototype.FileType == TextFile {
		return errors.New("trying to create file, where it's config type")
	}

	prototypeLayer, err := filePrototype.CreateLayer(content, optionalKeys, isOptional)
	if err != nil {
		return err
	}

	err = filePrototype.AddNewLayer(prototypeLayer, false)
	return err
}

func (v *VRCTFs) UpdateConfig(filePath string, content []byte, optionalKeys []byte, isOptional bool, fileType FileType) error {
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

	if filePrototype.FileType == TextFile {
		return errors.New("trying to create file, where it's config type")
	}

	originalContent, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	originalPrototypeLayer, err := filePrototype.CreateLayer(originalContent, nil, true)
	if err != nil {
		return err
	}

	err = filePrototype.AddNewLayer(originalPrototypeLayer, true)
	if err != nil {
		return err
	}

	prototypeLayer, err := filePrototype.CreateLayer(content, optionalKeys, isOptional)
	if err != nil {
		return err
	}

	err = filePrototype.AddNewLayer(prototypeLayer, false)
	return err
}
