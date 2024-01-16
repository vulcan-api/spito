package vrctFs

import "os"

type PrototypeLayer struct {
	// If ContentPath is specified and file exists in real fs, real file will be later overridden by this content
	// (We don't store content as string in order to make bson lightweight and fast accessible)
	ContentPath string `bson:",omitempty"`
	OptionsPath string `bson:",omitempty"`
	IsOptional  bool
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
