package vrctFs

const (
	TextFile = iota
	JsonConfig
	YamlConfig
	TomlConfig
	IniConfig // TODO: add support (complex feature)
	XmlConfig // TODO: maybe add support (requires custom algorithm)
)
