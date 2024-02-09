package vrctFs

type FileType int

const (
	TextFile FileType = iota
	JsonConfig
	YamlConfig
	TomlConfig
)
