package vrctFs

import "fmt"

const (
	JsonConfig = iota
	YamlConfig
	TomlConfig
)

type ConfigOption struct {
	List   []ConfigOption          `bson:",omitempty"`
	Map    map[string]ConfigOption `bson:",omitempty"`
	IsNull bool                    `bson:",omitempty"`
	Bool   *bool                   `bson:",omitempty"`
	Int    *int                    `bson:",omitempty"`
	Float  *float64                `bson:",omitempty"`
	String *string                 `bson:",omitempty"`
}

func (c *ConfigOption) MergeWith(c2 ConfigOption) error {
	if c2.IsNull {
		return nil
	}

	if c2.Bool != nil {
		if c.Bool != nil {
			// TODO: use special struct for this kind of error
			return fmt.Errorf("config conflict")
		}
		c.Bool = c2.Bool
		return nil
	}

	if c2.Int != nil {
		if c.Int != nil {
			// TODO: use special struct for this kind of error
			return fmt.Errorf("config conflict")
		}
		c.Int = c2.Int
		return nil
	}

	if c2.Float != nil {
		if c.Float != nil {
			// TODO: use special struct for this kind of error
			return fmt.Errorf("config conflict")
		}
		c.Float = c2.Float
		return nil
	}

	if c2.String != nil {
		if c.String != nil {
			// TODO: use special struct for this kind of error
			return fmt.Errorf("config conflict")
		}
		c.String = c2.String
		return nil
	}

	if len(c2.List) != 0 {
		for _, elem := range c2.List {
			_ = elem
			// TODO
		}
	}

	return fmt.Errorf("error occured during merge: config is neither merged nor null")
}
