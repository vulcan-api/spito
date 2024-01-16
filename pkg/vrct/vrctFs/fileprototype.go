package vrctFs

import (
	"encoding/json"
	"errors"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

type FilePrototype struct {
	Layers         []PrototypeLayer
	RealFileExists bool
	FileType       int
	Path           string `bson:"-"`
	Name           string `bson:"-"`
}

func (p *FilePrototype) getDestinationPath() string {
	newPath := strings.TrimPrefix(p.getVirtualPath(), VirtualFsPathPrefix)

	// Remove first slash
	newPath = newPath[1:]

	firstSlashIndex := strings.Index(newPath, "/")
	newPath = newPath[firstSlashIndex:]

	return strings.TrimSuffix(newPath, ".prototype.bson")
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

	if finalLayer.ContentPath == "" {
		filePath = p.getDestinationPath()
	} else {
		filePath = finalLayer.ContentPath
	}

	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var tempContentInterface map[string]interface{}
	if p.FileType != TextFile {
		err = bson.Unmarshal(file, &tempContentInterface)
		if err != nil {
			return file, err
		}
	}

	var fileContent []byte
	switch p.FileType {
	case JsonConfig:
		fileContent, err = json.Marshal(tempContentInterface)
		break
	case YamlConfig:
		fileContent, err = yaml.Marshal(tempContentInterface)
		break
	case TomlConfig:
		fileContent, err = toml.Marshal(tempContentInterface)
		break
	default:
		return file, nil
	}
	return fileContent, err
}

// TODO: think of splitting it up into to functions (read and load)
func (p *FilePrototype) Read(vrctPrefix string, realPath string) error {
	prototypeFilePath := vrctPrefix + realPath

	// TODO: instead of adding slash, use filepath.Join in appropriate lines
	path := filepath.Dir(prototypeFilePath)
	path += "/"
	name := filepath.Base(prototypeFilePath)

	p.Path = path
	p.Name = name
	file, err := os.ReadFile(p.getVirtualPath())

	if os.IsNotExist(err) {
		_, err := os.Stat(realPath)
		p.RealFileExists = !os.IsNotExist(err)

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

func (p *FilePrototype) CreateLayer(content []byte, options []byte, isOptional bool) (PrototypeLayer, error) {
	if p.Path == "" {
		return PrototypeLayer{}, errors.New("file prototype hasn't been loaded yet")
	}

	// TODO: check if file with random name already exist
	dir := filepath.Dir(p.Path)

	randFileName := randomLetters(5)
	contentPath := filepath.Join(dir, randFileName)

	randOptsName := randomLetters(5)
	optionsPath := filepath.Join(dir, randOptsName)

	tempConvertedContent, err := GetMapFromBytes(content, p.FileType)

	if p.FileType != TextFile {
		content, err = bson.Marshal(tempConvertedContent)
		if err != nil {
			return PrototypeLayer{}, err
		}
	}

	// TODO: think about changing ModePerm (spito can be run as root (dangerous))
	if err = os.WriteFile(contentPath, content, os.ModePerm); err != nil {
		return PrototypeLayer{}, err
	}

	if options == nil {
		options = []byte("{}")
	}

	var tempOptionalKeysMap map[string]interface{}
	err = json.Unmarshal(options, &tempOptionalKeysMap)
	if err != nil {
		return PrototypeLayer{}, err
	}

	optionalKeysBson, err := bson.Marshal(tempOptionalKeysMap)
	if err != nil {
		return PrototypeLayer{}, err
	}

	if err = os.WriteFile(optionsPath, optionalKeysBson, os.ModePerm); err != nil {
		return PrototypeLayer{}, err
	}

	newLayer := PrototypeLayer{
		ContentPath: contentPath,
		OptionsPath: optionsPath,
		IsOptional:  isOptional,
	}

	return newLayer, nil
}

func (p *FilePrototype) AddNewLayer(layer PrototypeLayer) error {
	// TODO: merge first and check if new layers is addable
	p.Layers = append(p.Layers, layer)
	if err := p.Save(); err != nil {
		return err
	}

	return nil
}

func (layer *PrototypeLayer) GetContent() ([]byte, error) {
	file, err := os.ReadFile(layer.ContentPath)
	if err != nil {
		return file, err
	}

	return file, nil
}

func (layer *PrototypeLayer) SetContent(content []byte) error {
	err := os.WriteFile(layer.ContentPath, content, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
