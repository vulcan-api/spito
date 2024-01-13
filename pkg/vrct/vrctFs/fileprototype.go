package vrctFs

import (
	"gopkg.in/mgo.v2/bson"
	"os"
	"strings"
)

type FilePrototype struct {
	Layers         []PrototypeLayer
	RealFileExists bool
	FileType       *int
	SelfPath       string `bson:"-"`
}

func (p *FilePrototype) getDestinationPath() string {
	newPath := strings.TrimPrefix(p.SelfPath, VirtualFsPathPrefix)

	// Remove first slash
	newPath = newPath[1:]

	firstSlashIndex := strings.Index(newPath, "/")
	newPath = newPath[firstSlashIndex:]

	return strings.TrimSuffix(newPath, ".prototype.bson")
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

		if finalLayer.ContentPath == nil {
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

func (p *FilePrototype) Read(prototypePath string, destPath string) error {
	p.SelfPath = prototypePath
	file, err := os.ReadFile(prototypePath)

	if os.IsNotExist(err) {
		_, err := os.Stat(destPath)
		p.RealFileExists = !os.IsNotExist(err)

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
